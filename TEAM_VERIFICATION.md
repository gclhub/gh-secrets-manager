# Team Membership Verification Implementation

This document describes the implementation of team membership verification in the auth-server component.

## Summary

The auth-server has been enhanced to verify whether users belong to a specified team within an organization before they can retrieve a token. This adds an extra layer of security to ensure only authorized team members can access GitHub App tokens.

## Implementation Details

### Auth Server Changes

1. **New Command Line Arguments**:
   - `--organization`: GitHub organization name for team membership verification
   - `--team`: GitHub team name for membership verification

2. **Enhanced Token Endpoint**:
   - Accepts additional `username` query parameter
   - Accepts optional `org` query parameter (overrides server configuration)
   - Accepts optional `team` query parameter (overrides server configuration)

3. **Team Verification Logic**:
   - Uses GitHub API to check team membership within an organization
   - Requires active team membership (pending memberships are rejected)
   - Returns appropriate HTTP status codes (403 for non-members)

### CLI Changes

1. **Automatic Username Detection**:
   - CLI automatically gets current GitHub username using `gh` CLI
   - Username is passed to auth-server for verification

2. **Updated Client Options**:
   - Added `Username` field to `ClientOptions`
   - Added `Organization` and `Team` fields to `ClientOptions`
   - Enhanced token refresh logic to include verification parameters

3. **Configuration Support**:
   - Added organization and team fields to configuration file
   - CLI can pass team verification parameters to auth server

## Usage

### Starting Auth Server with Team Verification

```bash
# Start auth server with team verification
./auth-server \
  --private-key-path /path/to/app-private-key.pem \
  --organization myorg \
  --team myteam \
  --port 8080 \
  --verbose
```

### CLI Usage (No Changes Required)

The CLI automatically handles username detection and passes it to the auth server:

```bash
# Works as before - username is automatically detected
gh secrets-manager secrets list --org myorg
```

## Security Features

1. **Team Membership Verification**: Only active team members can get tokens
2. **Organization and Team Required**: Both organization and team must be specified for verification
3. **Active Membership Required**: Pending team memberships are rejected
4. **Graceful Degradation**: If no team is configured, verification is skipped
5. **Proper Error Handling**: Clear error messages for unauthorized users
6. **Backward Compatibility**: Existing deployments continue to work without changes

## API Endpoints

### Token Endpoint

```
POST /token?app-id=APP_ID&installation-id=INSTALLATION_ID&username=USERNAME&org=ORG&team=TEAM
```

**Parameters:**
- `app-id` (required): GitHub App ID
- `installation-id` (required): Installation ID
- `username` (optional): GitHub username for team verification
- `org` (optional): Organization name (overrides server configuration)
- `team` (optional): Team name (overrides server configuration)

**Response:**
- `200 OK`: Token granted (user is active team member)
- `403 Forbidden`: User not a member of required team or membership is pending
- `400 Bad Request`: Missing required parameters
- `500 Internal Server Error`: Server or GitHub API error

## Error Scenarios

1. **User Not Team Member**: Returns 403 with descriptive error message
2. **Pending Team Membership**: Returns 403 (only active memberships accepted)
3. **Missing Username**: Returns 400 when team verification is required
4. **Missing Organization**: Returns 400 when team is specified without organization
5. **GitHub API Errors**: Returns 500 with error details
6. **Network Issues**: Proper error propagation and logging

## Team Membership States

The GitHub API returns different states for team memberships:
- `active`: User is an active team member (verification passes)
- `pending`: User has a pending invitation (verification fails)

Only `active` memberships are accepted for token issuance.

## Testing

Comprehensive test coverage includes:
- Team membership verification scenarios
- Different membership states (active, pending)
- Error handling for various failure modes
- Integration tests for complete auth flow
- Unit tests for individual components

Run tests with:
```bash
cd auth-server
go test ./... -v
```