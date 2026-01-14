let prompt = $"
STUDY scratch/comments.md
STUDY scratch/NOTES.md

Pick the one most important issue to fix and then fix it.

If there are no issues left to fix, simply kill this PID: ($nu.pid)

Important:
- make sure all tests pass when done
- git commit after you are done
- save any notes or note any problems in scratch/NOTES.md 
"

for i in 1..10 {
  print $"Iteration #($i)..."
  opencode run $prompt -m anthropic/claude-sonnet-4-5
  print "Done!"
}
