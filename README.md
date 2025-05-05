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

## GitHub App Authentication

This tool supports two authentication methods:
1. Personal Access Token (PAT) through GitHub CLI (default)
2. GitHub App-based authentication for enhanced security and temporary access

### Setting up the GitHub App

1. Create a new GitHub App in your organization:
   - Go to Organization Settings > Developer Settings > GitHub Apps
   - Create a new app with the following permissions:
     - Repository permissions:
       - Actions: Read & Write
       - Secrets: Read & Write
       - Variables: Read & Write
     - Organization permissions:
       - Actions: Read & Write
       - Secrets: Read & Write
       - Variables: Read & Write
   - Generate and download a private key
   - Note your App ID

2. Install the app in your organization:
   - After creating the app, install it in your organization
   - Note the Installation ID from the installation URL or via the API

### Running the Auth Server

#### Development Mode

1. Clone the auth-server component:
```bash
git clone <repo>/secrets-manager
cd secrets-manager/auth-server
```

2. Start the server with your GitHub App credentials:
```bash
go run cmd/server/main.go --port 8080 --private-key-path /path/to/private-key.pem
```

#### Production Mode

1. Build the auth server:
```bash
cd auth-server
go build -o bin/auth-server cmd/server/main.go
```

2. Deploy the server with your configuration:
```bash
./bin/auth-server \
  --port 443 \
  --private-key-path /path/to/private-key.pem
```

We recommend:
- Running behind a reverse proxy with TLS
- Using environment variables or a config management system for the private key
- Implementing additional access controls and rate limiting
- Monitoring server health and token usage

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

### Managing Secrets

```bash
# List organization secrets
gh secrets-manager secrets list --org myorg

# List repository secrets
gh secrets-manager secrets list --repo owner/repo

# List secrets for repositories with specific property
gh secrets-manager secrets list --org myorg --property language --value go

# Create/update organization secret
gh secrets-manager secrets set --org myorg --name SECRET_NAME --value "secret_value"

# Create/update repository secret
gh secrets-manager secrets set --repo owner/repo --name SECRET_NAME --value "secret_value"

# Import secrets from file
gh secrets-manager secrets set --org myorg --file secrets.json
gh secrets-manager secrets set --repo owner/repo --file secrets.csv

# Delete secret
gh secrets-manager secrets delete --org myorg --name SECRET_NAME
gh secrets-manager secrets delete --repo owner/repo --name SECRET_NAME
```

### Managing Variables

```bash
# List organization variables
gh secrets-manager variables list --org myorg

# List repository variables
gh secrets-manager variables list --repo owner/repo

# Create/update organization variable
gh secrets-manager variables set --org myorg --name VAR_NAME --value "value"

# Create/update repository variable
gh secrets-manager variables set --repo owner/repo --name VAR_NAME --value "value"

# Import variables from file
gh secrets-manager variables set --org myorg --file variables.json
gh secrets-manager variables set --repo owner/repo --file variables.csv

# Delete variable
gh secrets-manager variables delete --org myorg --name VAR_NAME
gh secrets-manager variables delete --repo owner/repo --name VAR_NAME
```

### Managing Dependabot Secrets

```bash
# List organization Dependabot secrets
gh secrets-manager dependabot list --org myorg

# List repository Dependabot secrets
gh secrets-manager dependabot list --repo owner/repo

# Create/update organization Dependabot secret
gh secrets-manager dependabot set --org myorg --name SECRET_NAME --value "secret_value"

# Create/update repository Dependabot secret
gh secrets-manager dependabot set --repo owner/repo --name SECRET_NAME --value "secret_value"

# Import Dependabot secrets from file
gh secrets-manager dependabot set --org myorg --file secrets.json
gh secrets-manager dependabot set --repo owner/repo --file secrets.csv

# Delete Dependabot secret
gh secrets-manager dependabot delete --org myorg --name SECRET_NAME
gh secrets-manager dependabot delete --repo owner/repo --name SECRET_NAME
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

When using the `--property` and `--value` flags, you can filter repositories by:

- `name`: Repository name
- `description`: Repository description
- `language`: Primary programming language
- `visibility`: Repository visibility (public/private)
- `is_private`: Private repository status (true/false)
- `has_issues`: Issues enabled status (true/false)
- `has_wiki`: Wiki enabled status (true/false)
- `archived`: Archive status (true/false)
- `disabled`: Disabled status (true/false)
- Repository topics (matches any topic)