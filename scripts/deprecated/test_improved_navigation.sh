#!/bin/bash

# Test script for r9s with improved navigation support
# This script tests the fixed navigation in r9s, including handling of offline mode

echo -e "\n\033[1;36m===== R9S Offline Navigation Test =====\033[0m"
echo -e "\033[1;33mThis test launches r9s and allows you to test the navigation improvements:\033[0m"
echo -e "  • Fixed refresh for all views - no more frozen loading screens"
echo -e "  • Proper namespace counts in project view"
echo -e "  • Context-appropriate status bar messages"
echo -e "  • Improved breadcrumb navigation"
echo -e "  • Automatic fallback to mock data when offline"

echo -e "\n\033[1;36m===== Navigation Testing Instructions =====\033[0m"
echo -e "\033[1;33mUse these keys to navigate the UI:\033[0m"
echo -e "  • \033[1;37mEnter\033[0m - Select and navigate deeper (Cluster → Project → Namespace → Pod)"
echo -e "  • \033[1;37mEsc\033[0m - Go back one level"
echo -e "  • \033[1;37md\033[0m - Describe the selected resource (when on a Pod)"
echo -e "  • \033[1;37mr\033[0m - Refresh the current view (fixed to work on all views)"
echo -e "  • \033[1;37mq\033[0m - Quit"

echo -e "\n\033[1;36m===== Expected Results =====\033[0m"
echo -e "\033[1;33mVerify these improvements:\033[0m"
echo -e "  1. App starts at Clusters view (not directly in Pods)"
echo -e "  2. Refreshing any screen with 'r' works properly"
echo -e "  3. Status bar shows appropriate counts for each view"
echo -e "  4. Project view shows correct namespace counts" 
echo -e "  5. Navigation flow works: Clusters → Projects → Namespaces → Pods"
echo -e "  6. Breadcrumbs show the correct navigation path"

echo -e "\n\033[1;32mStarting r9s in 3 seconds...\033[0m"
sleep 3

# Execute r9s
./bin/r9s
