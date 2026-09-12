# Security

- Secrets, API keys, and tokens live only in `.env` (or a secret manager). `.env` must be in `.gitignore`.
- Never print or log URLs that contain an API key, session token, or other credential — mask or strip the sensitive segment before logging.
- Configuration that varies by environment: read from env vars, not hard-coded literals.
- Never log passwords, tokens, or PII, even at debug level.
