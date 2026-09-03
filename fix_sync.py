with open("internal/core/sync.go", "r") as f:
    content = f.read()

# Replace raw newlines inside strings with actual \n characters
import re
# Wait, it's easier to just strip the raw newlines and insert \n.
# Actually I'll just sed the specific lines.
