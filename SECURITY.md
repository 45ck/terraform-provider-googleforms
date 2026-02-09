# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it
responsibly.

**Do NOT create a public GitHub issue for security vulnerabilities.**

Instead, please send a detailed report to the project maintainers via
GitHub's private vulnerability reporting feature or email.

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and provide a timeline for
resolution.

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.x.x   | Yes (pre-release) |

## Security Considerations

This provider handles sensitive data:

- **Credentials**: Service account keys and OAuth tokens are used for API access.
  Always use Application Default Credentials or environment variables instead of
  hardcoding credentials in Terraform configurations.

- **State files**: Terraform state may contain form content and configuration.
  Use encrypted remote state backends (S3 with SSE, GCS with CMEK, Terraform Cloud).

- **OAuth scopes**: The provider requests `forms.body` and `drive.file` scopes
  (minimal required). It never requests full `drive` scope by default.

- **Debug logging**: When `TF_LOG=DEBUG` is set, API request/response bodies
  may appear in logs. Authorization headers and credential content are never logged.
