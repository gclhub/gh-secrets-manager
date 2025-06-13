# Auth Server Documentation

The auth server component of gh-secrets-manager provides GitHub App-based authentication for enhanced security and temporary access tokens.

## Setting up the GitHub App

1. Create a new GitHub App in your organization:
   - Go to Organization Settings > Developer Settings > GitHub Apps
   - Create a new app with the following permissions:
     - Repository permissions:
       - Actions: Read & Write
       - Secrets: Read & Write
       - Variables: Read & Write
     - Organization permissions:
       - Secrets: Read & Write
       - Variables: Read & Write
   - Generate and download a private key
   - Note your App ID

2. Install the app in your organization:
   - After creating the app, install it in your organization
   - Note the Installation ID from the installation URL or via the API

## Running the Auth Server

### Development Mode

1. Clone the auth-server component:
```bash
git clone <repo>/secrets-manager
cd secrets-manager/auth-server
```

2. Start the server with your GitHub App credentials:
```bash
# Basic auth server (no access control)
go run cmd/server/main.go --port 8080 --private-key-path /path/to/private-key.pem

# With team membership verification (organization is optional and auto-detected)
go run cmd/server/main.go \
  --port 8080 \
  --private-key-path /path/to/private-key.pem \
  --team myteam \
  --verbose

# With explicit organization override
go run cmd/server/main.go \
  --port 8080 \
  --private-key-path /path/to/private-key.pem \
  --organization myorg \
  --team myteam \
  --verbose
```

### Production Mode

1. Build the auth server:
```bash
cd auth-server
go build -o bin/auth-server cmd/server/main.go
```

2. Deploy the server with your configuration:
```bash
./bin/auth-server \
  --port 443 \
  --private-key-path /path/to/private-key.pem \
  --team myteam \
  --verbose
```

Production deployment recommendations:
- Run behind a reverse proxy with TLS
- Use environment variables or a config management system for the private key
- Implement additional access controls and rate limiting
- Monitor server health and token usage
- Configure team membership verification for access control

## API Endpoints

The auth server exposes two HTTP endpoints:

### Health Check
```
GET /healthz
```
Returns 200 OK if the server is running. Useful for load balancer health checks and monitoring.

### Token Generation
```
POST /token?app-id=APP_ID&installation-id=INSTALLATION_ID&username=USERNAME&org=ORG&team=TEAM
```
Generates a GitHub installation access token. If team verification is configured, the user must be an active member of the specified team.

Parameters:
- `app-id` (required) - The GitHub App ID
- `installation-id` (required) - The installation ID for the organization
- `username` (optional) - GitHub username for team membership verification
- `org` (optional) - Organization name (overrides server configuration)
- `team` (optional) - Team name (overrides server configuration)

Response (200 OK):
```json
{
    "token": "ghs_xxxxxxxxxxxx",
    "expires_at": "2025-05-16T19:47:43Z"
}
```

Error Response (400, 401, 500):
```json
{
    "message": "error description"
}
