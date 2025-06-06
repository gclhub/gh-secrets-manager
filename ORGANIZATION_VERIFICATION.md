# Organization Verification Implementation

This document describes the implementation of organization membership verification in the auth-server component.

## Summary

The auth-server has been enhanced to verify whether users belong to a specified organization before they can retrieve a token. This adds an extra layer of security to ensure only authorized organization members can access GitHub App tokens.

## Implementation Details

### Auth Server Changes

1. **New Command Line Arguments**:
   - `--organization`: GitHub organization name for membership verification
   - `--team`: GitHub team name for membership verification (reserved for future use)

2. **Enhanced Token Endpoint**:
   - Accepts additional `username` query parameter
   - Accepts optional `org` query parameter (overrides server configuration)

3. **Organization Verification Logic**:
   - Uses GitHub API to check organization membership
   - Supports both public and private membership verification
   - Returns appropriate HTTP status codes (403 for non-members)

### CLI Changes

1. **Automatic Username Detection**:
   - CLI automatically gets current GitHub username using `gh` CLI
   - Username is passed to auth-server for verification

2. **Updated Client Options**:
   - Added `Username` field to `ClientOptions`
   - Enhanced token refresh logic to include username parameter

## Usage

### Starting Auth Server with Organization Verification

```bash
# Start auth server with organization verification
./auth-server \
  --private-key-path /path/to/app-private-key.pem \
  --organization myorg \
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

1. **Organization Membership Verification**: Only organization members can get tokens
2. **Graceful Degradation**: If no organization is configured, verification is skipped
3. **Proper Error Handling**: Clear error messages for unauthorized users
4. **Backward Compatibility**: Existing deployments continue to work without changes

## API Endpoints

### Token Endpoint

```
POST /token?app-id=APP_ID&installation-id=INSTALLATION_ID&username=USERNAME
```

**Parameters:**
- `app-id` (required): GitHub App ID
- `installation-id` (required): Installation ID
- `username` (optional): GitHub username for organization verification
- `org` (optional): Organization name (overrides server configuration)

**Response:**
- `200 OK`: Token granted
- `403 Forbidden`: User not a member of required organization
- `400 Bad Request`: Missing required parameters
- `500 Internal Server Error`: Server or GitHub API error

## Error Scenarios

1. **User Not Organization Member**: Returns 403 with descriptive error message
2. **Missing Username**: Returns 400 when organization verification is required
3. **GitHub API Errors**: Returns 500 with error details
4. **Network Issues**: Proper error propagation and logging

## Testing

Comprehensive test coverage includes:
- Organization membership verification scenarios
- Error handling for various failure modes
- Integration tests for complete auth flow
- Unit tests for individual components

Run tests with:
```bash
cd auth-server
go test ./... -v
```