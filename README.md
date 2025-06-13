# gh-secrets-manager

A GitHub CLI extension for managing GitHub Actions secrets and variables, and Dependabot secrets at both organization and repository levels.

## Features

- Manage GitHub Actions secrets at organization and repository levels
- Handle Dependabot secrets
- Manage GitHub Actions variables
- Support for both public and private repositories
- Secure secret value handling
- Batch operations support

## Installation

```bash
gh extension install gclhub/gh-secrets-manager
```

## Quick Start

1. **Install the extension**:
```bash
gh extension install gclhub/gh-secrets-manager
```

2. **Set up authentication** (choose one method):

   **Option A: GitHub App Authentication** (Recommended for organizations)
   1. Set up the auth server following the [Auth Server Documentation](docs/AUTH_SERVER.md)
   2. Configure the CLI to use your auth server:
   ```bash
   # Point to your auth server
   gh secrets-manager config set auth-server https://your-auth-server.example.com
   gh secrets-manager config set app-id YOUR_APP_ID
   gh secrets-manager config set installation-id YOUR_INSTALLATION_ID
   ```

   **Option B: Personal Access Token**
   ```bash
   gh auth login
   ```

3. **Start managing secrets**:
```bash
# Organization secrets
gh secrets-manager secrets list --org myorg
gh secrets-manager secrets set --org myorg --name DEPLOY_KEY --value "mysecret"

# Repository secrets
gh secrets-manager secrets list --repo owner/repo
gh secrets-manager secrets set --repo owner/repo --name API_KEY --value "mysecret"

# Variables
gh secrets-manager variables set --repo owner/repo --name ENV --value "production"

# Dependabot secrets
gh secrets-manager dependabot set --org myorg --name NPM_TOKEN --value "npmtoken"
```

See the [Configuration](#configuration) section for detailed setup instructions.

## Usage

### Managing Repository Secrets

```bash
# List all secrets in a repository
gh secrets-manager secrets list --repo owner/repo

# Set a new secret
gh secrets-manager secrets set --repo owner/repo --name API_KEY --value "mysecret"

# Remove a secret
gh secrets-manager secrets remove --repo owner/repo --name API_KEY
```

### Managing Organization Secrets

```bash
# List all organization secrets
gh secrets-manager secrets list --org myorg

# Set an organization secret
gh secrets-manager secrets set --org myorg --name DEPLOY_KEY --value "orgkey"

# Remove an organization secret
gh secrets-manager secrets remove --org myorg --name DEPLOY_KEY
```

### Managing Variables

```bash
# List all variables
gh secrets-manager variables list --repo owner/repo

# Set a new variable
gh secrets-manager variables set --repo owner/repo --name ENV --value "production"
```

### Managing Dependabot Secrets

```bash
# List Dependabot secrets
gh secrets-manager dependabot list --repo owner/repo

# Set a Dependabot secret
gh secrets-manager dependabot set --repo owner/repo --name NPM_TOKEN --value "token123"
```

## Configuration

This extension supports two authentication methods: GitHub App (recommended) and Personal Access Token (PAT).

### GitHub App Authentication Setup

1. Set up a GitHub App in your organization (see [Auth Server Documentation](docs/AUTH_SERVER.md))
2. Configure the CLI:
```bash
# Set the authentication server URL
gh secrets-manager config set auth-server https://your-auth-server.example.com

# Set your GitHub App credentials
gh secrets-manager config set app-id YOUR_APP_ID
gh secrets-manager config set installation-id YOUR_INSTALLATION_ID
```

### Personal Access Token Setup

The CLI automatically uses your GitHub CLI authentication. Just ensure you're logged in:
```bash
gh auth login
```

### Managing Configuration

```bash
# View all current settings
gh secrets-manager config view

# View specific setting
gh secrets-manager config get auth-server
```

For detailed information about setting up and running the authentication server, see [Auth Server Documentation](docs/AUTH_SERVER.md).

## Troubleshooting

- Ensure you have the latest version of the GitHub CLI installed
- Check that you have the necessary permissions in the organization/repository
- For detailed logs, add the `--verbose` flag to any command

## Support

- [Report Issues](https://github.com/gclhub/gh-secrets-manager/issues)
- [Contributing Guide](docs/CONTRIBUTING.md)
- [Security Policy](docs/SECURITY.md)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Authentication

This tool supports two authentication methods:

1. **GitHub App Authentication (Recommended)**:
   - Enhanced security with temporary access tokens
   - Configurable through the `config` command
   - Requires a running auth server
   - See [Auth Server Documentation](docs/AUTH_SERVER.md) for setup instructions

2. **Personal Access Token (Fallback)**:
   - Uses your GitHub CLI authentication
   - No additional configuration needed
   - Less secure than GitHub App authentication

The tool automatically uses GitHub App authentication when configured, falling back to PAT only if:
- GitHub App configuration is missing or incomplete
- There's an error loading the configuration
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