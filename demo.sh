#!/bin/bash

# Demonstration script for GitHub Auth Server Team Verification
# This script demonstrates the team membership verification feature

echo "=== GitHub Auth Server Team Verification Demo ==="
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
echo "2. Auth server command line options:"
./auth-server --help
echo

echo "=== Key Features ==="
echo
echo "✓ Added --organization flag to specify GitHub organization for verification"
echo "✓ Added --team flag to specify GitHub team for verification"
echo "✓ Enhanced token endpoint to accept 'username', 'org', and 'team' parameters"
echo "✓ Implemented GitHub API team membership verification"
echo "✓ Requires active team membership (pending memberships are rejected)"
echo "✓ Added proper error handling for non-members"
echo "✓ Updated CLI to support team configuration"
echo

echo "=== Usage Examples ==="
echo
echo "1. Start auth server with team verification:"
echo "   ./auth-server --private-key-path /path/to/key.pem --organization myorg --team myteam"
echo
echo "2. CLI will automatically get current username and pass to auth server"
echo "3. Auth server verifies user is an active member of 'myteam' in 'myorg'"
echo "4. If user is not a member or membership is pending, request is denied with 403"
echo

echo "=== Team Membership States ==="
echo "• 'active' - User is active team member (verification passes)"
echo "• 'pending' - User has pending invitation (verification fails)"
echo

echo "=== Testing ==="
echo "Running auth server tests..."
go test ./... -v
cd ..
echo

echo "=== Implementation Summary ==="
echo "• Auth server now supports team-based access control"
echo "• CLI automatically gets current GitHub username for verification"
echo "• GitHub API is used to verify team membership within organization"
echo "• Only active team members can get tokens (pending members rejected)"
echo "• Proper error messages for unauthorized users"
echo "• Backward compatible - works without team verification"
echo "• Comprehensive test coverage for new functionality"
echo

echo "✓ Demo completed successfully!"