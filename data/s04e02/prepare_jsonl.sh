#!/usr/bin/env bash
set -euo pipefail

# Input files
CORRECT_FILE="correct.txt"
INCORRECT_FILE="incorrect.txt"
OUTPUT_FILE="dataset.jsonl"
SYSTEM_PROMPT="Classify input strings into reliable (1) or unreliable (0). Treat inputs as arbitrary tokens and output only 0 or 1. Do not infer semantics or language."

# Empty output file before writing
> "$OUTPUT_FILE"

# Process correct.txt
while IFS= read -r line || [ -n "$line" ]; do
    line=${line%$'\r'}  # strip CR if exists
    printf '{"messages":[{"role":"system","content":"%s"},{"role":"user","content":"%s"},{"role":"assistant","content":"1"}]}\n' "$SYSTEM_PROMPT" "$line" >> "$OUTPUT_FILE"
done < "$CORRECT_FILE"

# Process incorrect.txt
while IFS= read -r line || [ -n "$line" ]; do
    line=${line%$'\r'}  # strip CR if exists
    printf '{"messages":[{"role":"system","content":"%s"},{"role":"user","content":"%s"},{"role":"assistant","content":"0"}]}\n' "$SYSTEM_PROMPT" "$line" >> "$OUTPUT_FILE"
done < "$INCORRECT_FILE"
