#!/bin/bash

# Test script to capture deployment API response

echo "Testing deployment API response..."
echo ""

# Run r9s with debug mode and navigate to deployments
# We'll capture the API response for deployments
R9S_DEBUG=1 timeout 10s ./bin/r9s 2>&1 | grep -A 200 "API Response for.*deployments" | head -50
