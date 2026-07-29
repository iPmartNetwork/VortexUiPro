# Security Policy

## Supported Versions

| Version | Supported |
|:-------:|:---------:|
| 0.0.x   | ✅ Active |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please **do not** open a public issue.

Instead, report it privately via:

- **Telegram:** [@VortexUiPro](https://t.me/VortexUiPro)
- **Email:** security@vortexuipro.dev

Please include:
- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Any potential mitigations

We will:
1. Acknowledge receipt within 48 hours
2. Investigate and develop a fix
3. Release a patch as soon as possible
4. Credit the reporter (if desired)

## Scope

The following are considered in-scope for security reports:
- Authentication bypass
- Authorization / RBAC bypass
- SQL injection
- Remote code execution
- Sensitive data exposure
- Cross-site scripting (XSS)
- Cross-site request forgery (CSRF)

## Out of Scope

- Missing HTTP headers (without demonstrable impact)
- Self-XSS
- Rate limiting bypass (without data loss)
- Social engineering

## Disclosure

We follow coordinated disclosure:
- We fix the issue first
- We release a patched version
- We publicly disclose **after** the patch is available

Thank you for helping keep VortexUiPro secure! 🔐
