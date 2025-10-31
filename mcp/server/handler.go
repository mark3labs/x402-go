package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mark3labs/x402-go"
	"github.com/mark3labs/x402-go/mcp"
)

// X402Handler wraps an MCP HTTP handler and adds x402 payment verification
type X402Handler struct {
	mcpHandler  http.Handler
	config      *Config
	facilitator Facilitator
}

// NewX402Handler creates a new x402 payment handler
func NewX402Handler(mcpHandler http.Handler, config *Config) *X402Handler {
	if config == nil {
		config = DefaultConfig()
	}

	facilitator := NewHTTPFacilitator(config.FacilitatorURL)

	return &X402Handler{
		mcpHandler:  mcpHandler,
		config:      config,
		facilitator: facilitator,
	}
}

// ServeHTTP intercepts HTTP requests to check for x402 payments
func (h *X402Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only intercept POST requests (JSON-RPC calls)
	if r.Method != http.MethodPost {
		h.mcpHandler.ServeHTTP(w, r)
		return
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeError(w, nil, -32700, "Parse error", nil)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse JSON-RPC request
	var jsonrpcReq struct {
		JSONRPC string          `json:"jsonrpc"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
		ID      interface{}     `json:"id"`
	}
	if err := json.Unmarshal(bodyBytes, &jsonrpcReq); err != nil {
		h.writeError(w, nil, -32700, "Parse error", nil)
		return
	}

	// Only intercept tools/call methods
	if jsonrpcReq.Method != "tools/call" {
		h.mcpHandler.ServeHTTP(w, r)
		return
	}

	// Parse tool call params
	var toolParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
		Meta      *struct {
			AdditionalFields map[string]interface{} `json:"-"`
		} `json:"_meta"`
	}
	if err := json.Unmarshal(jsonrpcReq.Params, &toolParams); err != nil {
		h.writeError(w, jsonrpcReq.ID, -32602, "Invalid params", nil)
		return
	}

	// Unmarshal _meta separately to get AdditionalFields
	if len(jsonrpcReq.Params) > 0 {
		var params map[string]interface{}
		if err := json.Unmarshal(jsonrpcReq.Params, &params); err == nil {
			if meta, ok := params["_meta"].(map[string]interface{}); ok {
				if toolParams.Meta == nil {
					toolParams.Meta = &struct {
						AdditionalFields map[string]interface{} `json:"-"`
					}{}
				}
				toolParams.Meta.AdditionalFields = meta
			}
		}
	}

	// Check if tool requires payment
	requirements, needsPayment := h.checkPaymentRequired(toolParams.Name)
	if !needsPayment {
		// Free tool - pass through
		h.mcpHandler.ServeHTTP(w, r)
		return
	}

	// Tool requires payment - extract payment from _meta
	payment := h.extractPayment(toolParams.Meta)
	if payment == nil {
		// No payment provided - send 402 error
		h.sendPaymentRequiredError(w, jsonrpcReq.ID, requirements)
		return
	}

	// Find matching requirement
	requirement, err := h.findMatchingRequirement(payment, requirements)
	if err != nil {
		h.writeError(w, jsonrpcReq.ID, 402, fmt.Sprintf("Payment invalid: %v", err), nil)
		return
	}

	// Verify payment with facilitator
	ctx, cancel := context.WithTimeout(r.Context(), mcp.PaymentVerifyTimeout)
	defer cancel()

	verifyResp, err := h.facilitator.Verify(ctx, payment, *requirement)
	if err != nil {
		if h.config.Verbose {
			fmt.Printf("Payment verification failed: %v\n", err)
		}
		h.writeError(w, jsonrpcReq.ID, -32603, fmt.Sprintf("Verification failed: %v", err), nil)
		return
	}

	if !verifyResp.IsValid {
		if h.config.Verbose {
			fmt.Printf("Payment rejected: %s\n", verifyResp.InvalidReason)
		}
		h.writeError(w, jsonrpcReq.ID, 402, fmt.Sprintf("Payment invalid: %s", verifyResp.InvalidReason), nil)
		return
	}

	// Payment verified - settle if not verify-only mode
	var settleResp *x402.SettlementResponse
	if !h.config.VerifyOnly {
		settleCtx, settleCancel := context.WithTimeout(r.Context(), mcp.PaymentSettleTimeout)
		defer settleCancel()

		settleResp, err = h.facilitator.Settle(settleCtx, payment, *requirement)
		if err != nil {
			if h.config.Verbose {
				fmt.Printf("Payment settlement failed: %v\n", err)
			}
			// Return error with settlement response in error data
			errorData := map[string]interface{}{
				"x402/payment-response": map[string]interface{}{
					"success":     false,
					"network":     payment.Network,
					"payer":       verifyResp.Payer,
					"errorReason": err.Error(),
				},
			}
			h.writeError(w, jsonrpcReq.ID, -32603, fmt.Sprintf("Settlement failed: %v", err), errorData)
			return
		}

		if !settleResp.Success {
			if h.config.Verbose {
				fmt.Printf("Payment settlement unsuccessful: %s\n", settleResp.ErrorReason)
			}
			// Return error with settlement response in error data
			errorData := map[string]interface{}{
				"x402/payment-response": settleResp,
			}
			h.writeError(w, jsonrpcReq.ID, -32603, fmt.Sprintf("Settlement unsuccessful: %s", settleResp.ErrorReason), errorData)
			return
		}
	}

	// Payment successful - forward request and inject settlement response in result
	h.forwardWithSettlementResponse(w, r, bodyBytes, jsonrpcReq.ID, settleResp)
}

// checkPaymentRequired checks if a tool requires payment
func (h *X402Handler) checkPaymentRequired(toolName string) ([]x402.PaymentRequirement, bool) {
	requirements, exists := h.config.PaymentTools[toolName]
	if !exists || len(requirements) == 0 {
		return nil, false
	}

	// Set resource field on requirements
	for i := range requirements {
		if requirements[i].Resource == "" {
			requirements[i].Resource = fmt.Sprintf("mcp://tools/%s", toolName)
		}
	}

	return requirements, true
}

// extractPayment extracts payment from params._meta["x402/payment"]
func (h *X402Handler) extractPayment(meta *struct {
	AdditionalFields map[string]interface{} `json:"-"`
}) *x402.PaymentPayload {
	if meta == nil || meta.AdditionalFields == nil {
		return nil
	}

	paymentData, ok := meta.AdditionalFields["x402/payment"]
	if !ok {
		return nil
	}

	// Marshal and unmarshal to convert to PaymentPayload
	paymentBytes, err := json.Marshal(paymentData)
	if err != nil {
		return nil
	}

	var payment x402.PaymentPayload
	if err := json.Unmarshal(paymentBytes, &payment); err != nil {
		return nil
	}

	return &payment
}

// findMatchingRequirement finds a requirement that matches the payment
func (h *X402Handler) findMatchingRequirement(payment *x402.PaymentPayload, requirements []x402.PaymentRequirement) (*x402.PaymentRequirement, error) {
	for i := range requirements {
		req := &requirements[i]
		if req.Network == payment.Network && req.Scheme == payment.Scheme {
			return req, nil
		}
	}
	return nil, fmt.Errorf("no matching requirement for network %s and scheme %s", payment.Network, payment.Scheme)
}

// sendPaymentRequiredError sends a 402 error with payment requirements
func (h *X402Handler) sendPaymentRequiredError(w http.ResponseWriter, id interface{}, requirements []x402.PaymentRequirement) {
	errorData := map[string]interface{}{
		"x402Version": 1,
		"error":       "Payment required to access this resource",
		"accepts":     requirements,
	}

	h.writeError(w, id, 402, "Payment required", errorData)
}

// forwardWithSettlementResponse forwards the request and injects settlement response in result._meta
func (h *X402Handler) forwardWithSettlementResponse(w http.ResponseWriter, r *http.Request, requestBody []byte, requestID interface{}, settleResp *x402.SettlementResponse) {
	// Create a response recorder to capture the MCP handler's response
	recorder := &responseRecorder{
		headerMap:  make(http.Header),
		statusCode: http.StatusOK,
	}

	// Restore request body
	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	// Forward to MCP handler
	h.mcpHandler.ServeHTTP(recorder, r)

	// Parse response
	var jsonrpcResp struct {
		JSONRPC string          `json:"jsonrpc"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   interface{}     `json:"error,omitempty"`
		ID      interface{}     `json:"id"`
	}

	if err := json.Unmarshal(recorder.body.Bytes(), &jsonrpcResp); err != nil {
		// If we can't parse response, just forward it as-is
		w.WriteHeader(recorder.statusCode)
		_, _ = w.Write(recorder.body.Bytes())
		return
	}

	// Only inject settlement response if there's a result (not an error)
	if jsonrpcResp.Error == nil && jsonrpcResp.Result != nil && settleResp != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(jsonrpcResp.Result, &result); err == nil {
			// Get or create _meta
			meta, ok := result["_meta"].(map[string]interface{})
			if !ok {
				meta = make(map[string]interface{})
			}

			// Add settlement response
			meta["x402/payment-response"] = settleResp
			result["_meta"] = meta

			// Re-marshal result
			modifiedResult, err := json.Marshal(result)
			if err == nil {
				jsonrpcResp.Result = modifiedResult
			}
		}
	}

	// Write modified response
	responseBytes, err := json.Marshal(jsonrpcResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Copy headers
	for k, v := range recorder.headerMap {
		w.Header()[k] = v
	}

	w.WriteHeader(recorder.statusCode)
	_, _ = w.Write(responseBytes)
}

// writeError writes a JSON-RPC error response
func (h *X402Handler) writeError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	errorResp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}

	if data != nil {
		errorResp["error"].(map[string]interface{})["data"] = data
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 status
	_ = json.NewEncoder(w).Encode(errorResp)
}

// responseRecorder records HTTP responses for modification
type responseRecorder struct {
	headerMap  http.Header
	body       bytes.Buffer
	statusCode int
}

func (r *responseRecorder) Header() http.Header {
	return r.headerMap
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}
