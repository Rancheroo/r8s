#!/bin/bash
# coderabbit-reply.sh - Easy threaded replies to CodeRabbit comments
# Usage: ./coderabbit-reply.sh <comment_id> <message>
# Example: ./coderabbit-reply.sh 2831504030 "@coderabbit Acknowledged"

set -e

COMMENT_ID=${1:-""}
MESSAGE=${2:-""}
PR_NUMBER=${3:-"64"}
REPO="Rancheroo/r8s"

if [ -z "$COMMENT_ID" ] || [ -z "$MESSAGE" ]; then
    echo "Usage: $0 <comment_id> <message> [pr_number]"
    echo "Example: $0 2831504030 '@CodeRabbitAI Fixed in commit abc123'"
    exit 1
fi

# Get current commit
COMMIT=$(git rev-parse HEAD)

# Get comment details using gh (no jq needed)
echo "Fetching comment $COMMENT_ID..."
COMMENT_JSON=$(gh api repos/$REPO/pulls/comments/$COMMENT_ID 2>/dev/null || echo "")

if [ -z "$COMMENT_JSON" ]; then
    echo "❌ Error: Could not fetch comment $COMMENT_ID"
    exit 1
fi

# Extract path and line using grep (no jq dependency)
PATH=$(echo "$COMMENT_JSON" | grep '"path"' | head -1 | sed 's/.*"path": "\([^"]*\)".*/\1/')
LINE=$(echo "$COMMENT_JSON" | grep '"line"' | head -1 | sed 's/.*"line": \([0-9]*\).*/\1/')

# Validate we have required fields
if [ -z "$PATH" ] || [ "$PATH" == "null" ]; then
    echo "❌ Error: Could not determine file path"
    exit 1
fi

# Use original_line if line is null or 0
if [ -z "$LINE" ] || [ "$LINE" == "null" ] || [ "$LINE" == "0" ]; then
    LINE=$(echo "$COMMENT_JSON" | grep '"original_line"' | head -1 | sed 's/.*"original_line": \([0-9]*\).*/\1/')
fi

# Default to line 1 if still empty
if [ -z "$LINE" ] || [ "$LINE" == "null" ] || [ "$LINE" == "0" ]; then
    LINE=1
fi

echo "Replying to comment $COMMENT_ID on $PATH:$LINE"

# Create threaded reply
JSON_BODY=$(cat <<EOF
{
  "body": "$MESSAGE",
  "commit_id": "$COMMIT",
  "path": "$PATH",
  "line": $LINE,
  "in_reply_to": $COMMENT_ID
}
EOF
)

RESPONSE=$(echo "$JSON_BODY" | gh api repos/$REPO/pulls/$PR_NUMBER/comments --input - 2>&1)

if echo "$RESPONSE" | grep -q '"html_url"'; then
    COMMENT_URL=$(echo "$RESPONSE" | grep '"html_url"' | head -1 | sed 's/.*"html_url": "\([^"]*\)".*/\1/')
    echo "✅ Posted: $COMMENT_URL"
else
    echo "❌ Failed"
    echo "$RESPONSE" | head -5
    exit 1
fi
