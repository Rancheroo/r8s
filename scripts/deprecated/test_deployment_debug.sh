#!/bin/bash

# Simple test to capture deployment API response
echo "Capturing deployment API response..."
echo ""

# Navigate to deployments view and capture debug output
(
  # Send keys to navigate: Enter (clusters) -> Enter (projects) -> Enter (namespaces) -> 2 (deployments)  
  sleep 2 && printf '\n' &&
  sleep 2 && printf '\n' &&
  sleep 2 && printf '\n' &&
  sleep 2 && printf '2' &&
  sleep 5
) | R9S_DEBUG=1 timeout 20s ./bin/r9s 2>&1 | grep -A 500 "deployments" | head -100
