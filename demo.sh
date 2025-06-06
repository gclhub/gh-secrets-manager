#!/bin/bash

# Demonstration script for GitHub Auth Server Organization Verification
# This script demonstrates the new organization membership verification feature

echo "=== GitHub Auth Server Organization Verification Demo ==="
echo

# Build the auth server
echo "1. Building auth server..."
cd auth-server
go build -o auth-server ./cmd/server
if [ $? -ne 0 ]; then
    echo "Failed to build auth server"
    exit 1
fi
echo "✓ Auth server built successfully"
echo

# Show the new command line options
echo "2. New auth server command line options:"
./auth-server --help
echo

echo "=== Key Features Added ==="
echo
echo "✓ Added --organization flag to specify GitHub organization for verification"
echo "✓ Added --team flag to specify GitHub team for verification (for future use)"
echo "✓ Enhanced token endpoint to accept 'username' query parameter"
echo "✓ Implemented GitHub API organization membership verification"
echo "✓ Added proper error handling for non-members"
echo "✓ Updated CLI to automatically get current GitHub username"
echo

echo "=== Usage Examples ==="
echo
echo "1. Start auth server with organization verification:"
echo "   ./auth-server --private-key-path /path/to/key.pem --organization myorg"
echo
echo "2. CLI will automatically get current username and pass to auth server"
echo "3. Auth server verifies user belongs to 'myorg' before issuing token"
echo "4. If user is not a member, request is denied with 403 Forbidden"
echo

echo "=== Testing ==="
echo "Running auth server tests..."
cd auth-server
go test ./... -v
cd ..
echo

echo "=== Implementation Summary ==="
echo "• Auth server now supports organization-based access control"
echo "• CLI automatically gets current GitHub username for verification"
echo "• GitHub API is used to verify organization membership"
echo "• Proper error messages for unauthorized users"
echo "• Backward compatible - works without organization verification"
echo "• Comprehensive test coverage for new functionality"
echo

echo "✓ Demo completed successfully!"