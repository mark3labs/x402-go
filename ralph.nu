for i in 1..10 {
  print $"Iteration #($i)..."
  opencode run "Prompt: " -f scratch/PROMPT.md -m anthropic/claude-sonnet-4-5
  print "Done!"
}
