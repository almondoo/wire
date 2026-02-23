# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please use [GitHub's private vulnerability reporting](https://github.com/almondoo/wire/security/advisories/new) to submit your report.

You can expect:
- An acknowledgment within **48 hours**
- A status update within **7 days**
- A fix or mitigation plan within **30 days** for confirmed vulnerabilities

## Scope

This project is a compile-time code generator. It does not handle network traffic, user authentication, or sensitive data at runtime. However, we take security seriously for:

- Supply chain security (dependencies, build process)
- Code generation correctness (generated code should not introduce vulnerabilities)
- Repository and CI/CD security
