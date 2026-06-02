# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| latest (main branch) | ✅ |
| older releases | ❌ |

traCtl is in early development. Only the current `main` branch receives security fixes.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Report security issues privately by:
1. Opening a GitHub issue with the label `security` marked as **private/confidential**, or
2. Emailing the maintainers directly (see README for contact details)

Include as much detail as possible:
- Type of vulnerability (e.g. injection, path traversal, credential leak)
- Which component is affected (`internal/sandbox`, `cmd/localapi`, etc.)
- Steps to reproduce
- Potential impact

## Response Timeline

| Step | Timeline |
|------|----------|
| Acknowledgement of report | Within 48 hours |
| Initial assessment | Within 5 business days |
| Fix or mitigation | Depends on severity — critical within 7 days |
| Public disclosure | After fix is released |

## Scope

**In scope:**
- JavaScript sandbox escape (`internal/sandbox`)
- Credential leakage in logs or traces
- Path traversal via overlay or file endpoints
- Authentication bypass in `cmd/localapi`
- Remote code execution via workflow files

**Out of scope:**
- Vulnerabilities in third-party dependencies (report upstream)
- Denial of service against the local API (it assumes a trusted local environment)
- Issues requiring physical access to the machine

## Security Architecture Notes

- The JS sandbox (`internal/sandbox`) uses goja — a pure-Go engine with no I/O access
- `cmd/localapi` binds to `localhost` only and has no authentication (trusted local network assumption)
- Authorization headers are masked in all diagnostic output
- See [ADR-006](docs/adr/ADR-006-security-trust-model.md) for the full security trust model
