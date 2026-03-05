#!/bin/bash
echo "PATH is: $PATH"
which base64
which jq
echo "Testing base64..."
echo "test" | base64 | base64 --decode
echo ""
echo "Testing jq..."
echo '{"foo":"bar"}' | jq -r .foo
