# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of gh-secrets-manager seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Reporting Process

1. **Do Not** create a public GitHub issue for the vulnerability.
2. Email your findings to [gclhub@github.com]. Include:
   - A description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact
   - Any possible mitigations

### What to Expect

- You will receive an acknowledgment within 48 hours.
- We will investigate and keep you informed of our progress.
- Once fixed, we will notify you and publicly disclose the issue.

## Security Considerations

### Authentication

1. **GitHub App Authentication** (Recommended)
   - Uses temporary tokens with limited scope
   - Requires a secure auth server setup
   - Implements team-based access controls
   - Tokens are automatically rotated

2. **Personal Access Token**
   - Use only when GitHub App authentication is not possible
   - Ensure tokens have minimal required permissions
   - Rotate tokens regularly

### Secret Storage

- Secrets are never stored locally
- All secret values are encrypted in transit
- Configuration files are created with secure permissions (0644)

### Auth Server Security

When deploying the auth server:
- Use HTTPS with valid certificates
- Implement proper access controls
- Monitor for unusual activity
- Keep the private key secure
- Enable audit logging

## Best Practices

1. **Access Control**
   - Use GitHub App authentication when possible
   - Implement team-based access controls
   - Review access permissions regularly

2. **Secret Management**
   - Rotate secrets regularly
   - Use environment-specific secrets
   - Audit secret access and changes

3. **Configuration**
   - Keep auth server credentials secure
   - Use secure communication channels
   - Regularly update dependencies

4. **Monitoring**
   - Enable verbose logging in production
   - Monitor auth server access
   - Review GitHub audit logs
