#!/bin/bash
# scripts/review-reply.sh - Automate CodeRabbit replies
# Usage: ./scripts/review-reply.sh <PR_NUMBER> [message]

set -e

PR=$1
DEFAULT_MSG=$2

if [ -z "$PR" ]; then
    echo "Usage: $0 <PR_NUMBER> [message]"
    echo "  If message is provided, it replies to ALL unanswered comments with it."
    echo "  Otherwise, it enters interactive mode."
    exit 1
fi

REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo "🔍 Analyzing PR #$PR in $REPO..."

# 1. Get all comments
echo "📥 Fetching comments..."
ALL_COMMENTS=$(gh api repos/$REPO/pulls/$PR/comments --paginate --jq '.')

# 2. Find answered IDs (comments that are replies to something)
ANSWERED_IDS=$(echo "$ALL_COMMENTS" | jq -r '.[].in_reply_to_id | select(. != null)')

# 3. Find CodeRabbit root comments that are NOT answered
# We iterate through all comments where user is coderabbitai[bot] and it's a root comment (in_reply_to_id is null)
# Then checks if its ID is in ANSWERED_IDS

UNANSWERED=$(echo "$ALL_COMMENTS" | jq -r --argjson answered "$(echo "$ANSWERED_IDS" | jq -R . | jq -s .)" '
  .[] | 
  select(.user.login == "coderabbitai[bot]" and .in_reply_to_id == null) | 
  select(.id as $id | $answered | index($id) | not) |
  {id, path, line, body: (.body | split("\n")[0])} | 
  @base64
')

COUNT=0
for row in $UNANSWERED; do
    COUNT=$((COUNT + 1))
done

if [ "$COUNT" -eq 0 ]; then
    echo "✅ No unanswered CodeRabbit comments found!"
    exit 0
fi

echo "⚠️  Found $COUNT unanswered comments."

for row in $UNANSWERED; do
    _jq() {
     echo ${row} | base64 --decode | jq -r ${1}
    }
    
    ID=$(_jq '.id')
    PATH=$(_jq '.path')
    LINE=$(_jq '.line')
    BODY=$(_jq '.body')
    
    echo ""
    echo "----------------------------------------------------------------"
    echo "📄 File: $PATH:$LINE"
    echo "💬 Comment: $BODY..."
    echo "----------------------------------------------------------------"
    
    MSG=""
    if [ -n "$DEFAULT_MSG" ]; then
        MSG="$DEFAULT_MSG"
    else
        read -p "Reply (Enter to skip, 'q' to quit): " MSG
    fi
    
    if [ "$MSG" == "q" ]; then
        break
    fi
    
    if [ -n "$MSG" ]; then
        echo "🚀 Replying..."
        gh api repos/$REPO/pulls/$PR/comments/$ID/replies -f body="$MSG" > /dev/null
        echo "✅ Replied."
    else
        echo "⏭️  Skipped."
    fi
done
