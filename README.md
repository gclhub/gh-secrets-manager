# gh-secrets-manager

A GitHub CLI extension for managing GitHub Actions secrets and variables, and Dependabot secrets at both organization and repository levels.

## Installation

```bash
gh extension install gclhub/gh-secrets-manager
```

## Features

- Manage secrets and variables at the organization level
- Manage secrets and variables for individual repositories
- Manage secrets and variables for repositories matching specific properties
- Import secrets and variables from JSON or CSV files
- List current values for variables and redacted secrets
- Update secret and variable values
- Delete secrets and variables
- Support for both GitHub Actions and Dependabot secrets
- Configure GitHub App authentication settings

## GitHub App Authentication

This tool supports two authentication methods:
1. GitHub App-based authentication for enhanced security and temporary access (preferred when configured)
2. Personal Access Token (PAT) through GitHub CLI (fallback)

When you have configured all required GitHub App settings (auth-server, app-id, and installation-id), the tool will automatically use GitHub App authentication. It only falls back to PAT authentication if:
- GitHub App configuration is missing or incomplete
- There's an error loading the configuration

### Setting up the GitHub App

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

### Initial Configuration

After setting up your GitHub App and auth server, configure the CLI extension to use them:

```bash
# Set the authentication server URL
gh secrets-manager config set auth-server https://your-auth-server.example.com

# Set your GitHub App ID
gh secrets-manager config set app-id YOUR_APP_ID

# Set your Installation ID
gh secrets-manager config set installation-id YOUR_INSTALLATION_ID

# Verify your configuration
gh secrets-manager config view
```

Once configured, the extension will automatically use GitHub App authentication for all commands without requiring additional flags.

### Running the Auth Server

#### Development Mode

1. Clone the auth-server component:
```bash
git clone <repo>/secrets-manager
cd secrets-manager/auth-server
```

2. Start the server with your GitHub App credentials:
```bash
# Basic auth server (no access control)
go run cmd/server/main.go --port 8080 --private-key-path /path/to/private-key.pem

# With team membership verification
go run cmd/server/main.go \
  --port 8080 \
  --private-key-path /path/to/private-key.pem \
  --organization myorg \
  --team myteam \
  --verbose
```

#### Production Mode

1. Build the auth server:
```bash
cd auth-server
go build -o bin/auth-server cmd/server/main.go
```

2. Deploy the server with your configuration:
```bash
# Basic auth server (no access control)
./bin/auth-server \
  --port 443 \
  --private-key-path /path/to/private-key.pem

# With team membership verification
./bin/auth-server \
  --port 443 \
  --private-key-path /path/to/private-key.pem \
  --organization myorg \
  --team myteam \
  --verbose
```

We recommend:
- Running behind a reverse proxy with TLS
- Using environment variables or a config management system for the private key
- Implementing additional access controls and rate limiting
- Monitoring server health and token usage
- Configuring team membership verification for access control

### Auth Server Endpoints

The auth server exposes two HTTP endpoints:

#### Health Check
```
GET /healthz
```
Returns 200 OK if the server is running. Useful for load balancer health checks and monitoring.

#### Token Generation
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
```

### Using GitHub App Authentication

To use the CLI with GitHub App authentication:

```bash
# List secrets using GitHub App authentication
gh secrets-manager secrets list \
  --org myorg \
  --app-id 123456 \
  --installation-id 987654 \
  --auth-server https://auth.example.com

# Import secrets with GitHub App authentication
gh secrets-manager secrets set \
  --org myorg \
  --file secrets.json \
  --app-id 123456 \
  --installation-id 987654 \
  --auth-server https://auth.example.com
```

The auth server will:
1. Generate a JWT using the GitHub App's private key
2. Exchange it for an installation access token
3. Return the token to the CLI
4. The CLI will use this token for all API calls
5. Tokens are automatically refreshed before expiration

## Usage

### Managing Configuration

The `config` subcommand allows you to manage persistent GitHub App authentication settings. Configuration is stored in the user's config directory under `gh/secrets-manager/config.json`.

### Available Commands

- `config view` - Display current configuration settings
- `config set <key> <value>` - Set a configuration value
- `config get <key>` - Get a specific configuration value
- `config delete <key>` - Delete a configuration value

### Configuration Keys

The following configuration keys are supported:

- `auth-server` - URL of the authentication server (e.g., https://auth.example.com)
- `app-id` - GitHub App ID (numeric)
- `installation-id` - GitHub App Installation ID (numeric)

### Examples

```bash
# View all current configuration
gh secrets-manager config view

# Get a specific configuration value
gh secrets-manager config get auth-server

# Set configuration values
gh secrets-manager config set auth-server https://auth.example.com
gh secrets-manager config set app-id 123456
gh secrets-manager config set installation-id 987654

# Delete a configuration value
gh secrets-manager config delete auth-server
```

### Configuration Storage

Configuration is automatically stored in:
- macOS: `~/Library/Application Support/gh/secrets-manager/config.json`
- Linux: `~/.config/gh/secrets-manager/config.json`
- Windows: `%APPDATA%\gh\secrets-manager\config.json`

The configuration file uses JSON format and is created automatically when you first set a value. File permissions are set to 0644 to ensure secure access.

When all required GitHub App settings (auth-server, app-id, and installation-id) are configured, the extension will automatically use GitHub App authentication for all commands without requiring additional command-line flags.

### Managing Secrets

```bash
# List organization secrets
gh secrets-manager secrets list --org myorg

# List repository secrets
gh secrets-manager secrets list --repo owner/repo

# List secrets for repositories with specific property
gh secrets-manager secrets list --org myorg --property team --prop_value backend

# Create/update organization secret
gh secrets-manager secrets set --org myorg --name API_KEY --value "secret123"

# Create/update repository secret
gh secrets-manager secrets set --repo owner/repo --name DB_PASS --value "dbpass"

# Import secrets from file
gh secrets-manager secrets set --org myorg --file secrets.json

# Set secret for specific repositories
gh secrets-manager secrets set --org myorg --property team --prop_value frontend --name NPM_TOKEN --value "npmtoken"

# Delete secret
gh secrets-manager secrets delete --org myorg --name API_KEY
gh secrets-manager secrets delete --repo owner/repo --name DB_PASS
```

### Managing Environment Secrets

You can manage secrets specific to GitHub Actions environments within a repository:

```bash
# List environment secrets
gh secrets-manager secrets list --repo owner/repo --environment prod

# Create/update environment secret
gh secrets-manager secrets set --repo owner/repo --environment prod --name SECRET_NAME --value "secret123"

# Delete environment secret
gh secrets-manager secrets delete --repo owner/repo --environment prod --name SECRET_NAME

# Import multiple environment secrets from file
gh secrets-manager secrets set --repo owner/repo --environment prod --file env-secrets.json
```

### Managing Variables

```bash
# List organization variables
gh secrets-manager variables list --org myorg

# List repository variables
gh secrets-manager variables list --repo owner/repo

# List variables for repositories with specific property
gh secrets-manager variables list --org myorg --property team --prop_value backend

# Create/update organization variable
gh secrets-manager variables set --org myorg --name VAR_NAME --value "value"

# Create/update repository variable
gh secrets-manager variables set --repo owner/repo --name VAR_NAME --value "value"

# Set variable for specific repositories
gh secrets-manager variables set --org myorg --property team --prop_value frontend --name CONFIG_ENV --value "production"

# Import variables from file
gh secrets-manager variables set --org myorg --file variables.json
gh secrets-manager variables set --repo owner/repo --file variables.csv

# Delete variable
gh secrets-manager variables delete --org myorg --name VAR_NAME
gh secrets-manager variables delete --repo owner/repo --name VAR_NAME
```

### Managing Environment Variables

Similarly, you can manage environment-specific variables:

```bash
# List environment variables
gh secrets-manager variables list --repo owner/repo --environment prod

# Create/update environment variable
gh secrets-manager variables set --repo owner/repo --environment prod --name VAR_NAME --value "value123"

# Delete environment variable
gh secrets-manager variables delete --repo owner/repo --environment prod --name VAR_NAME

# Import multiple environment variables from file
gh secrets-manager variables set --repo owner/repo --environment prod --file env-variables.json
```

### Managing Dependabot Secrets

```bash
# List organization Dependabot secrets
gh secrets-manager dependabot list --org myorg

# List repository Dependabot secrets
gh secrets-manager dependabot list --repo owner/repo

# List Dependabot secrets for repositories with specific property
gh secrets-manager dependabot list --org myorg --property team --prop_value backend

# Create/update organization Dependabot secret
gh secrets-manager dependabot set --org myorg --name NPM_TOKEN --value "npmtoken"

# Create/update repository Dependabot secret
gh secrets-manager dependabot set --repo owner/repo --name DOCKER_TOKEN --value "dockertoken"

# Set Dependabot secret for specific repositories
gh secrets-manager dependabot set --org myorg --property team --prop_value backend --name MAVEN_PASSWORD --value "mavenpass"

# Import Dependabot secrets from file
gh secrets-manager dependabot set --org myorg --file dependabot-secrets.json
gh secrets-manager dependabot set --repo owner/repo --file dependabot-secrets.csv

# Delete Dependabot secret
gh secrets-manager dependabot delete --org myorg --name NPM_TOKEN
gh secrets-manager dependabot delete --repo owner/repo --name DOCKER_TOKEN
```

## Input File Formats

### JSON

```json
{
  "SECRET_NAME": "secret_value",
  "ANOTHER_SECRET": "another_value"
}
```

or

```json
[
  {
    "name": "SECRET_NAME",
    "value": "secret_value"
  },
  {
    "name": "ANOTHER_SECRET",
    "value": "another_value"
  }
]
```

### CSV

```csv
name,value
SECRET_NAME,secret_value
ANOTHER_SECRET,another_value
```

## Repository Property Filtering

The `--property` and `--prop_value` flags allow you to target multiple repositories based on GitHub custom repository properties.

First, define custom properties in your organization's settings:
1. Go to your organization's settings
2. Navigate to Repository defaults > Properties
3. Add custom properties like `team`, `environment`, `service-tier`, etc.
4. Assign property values to your repositories

Then use these properties to target specific repositories:

```bash
# Set secrets for all backend team repositories
gh secrets-manager secrets set --org myorg --property team --prop_value backend --name DB_PASSWORD --value "secret123"

# Update config for all production tier repositories
gh secrets-manager variables set --org myorg --property service-tier --prop_value production --name API_URL --value "prod.api.example.com"

# Set variables for staging repositories
gh secrets-manager variables set --org myorg --property environment --prop_value staging --name LOG_LEVEL --value "debug"
```

Note: This feature requires that you have defined custom properties in your organization's settings and assigned values to repositories.