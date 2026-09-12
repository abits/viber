# Error handling

- Never silently swallow errors. If a return must be discarded, leave a one-line comment explaining why the failure is safe to ignore.
- Handle specific error types, not broad catch-alls. In Go: `errors.Is` / `errors.As`. In Python: named exception classes. Never `except:` or `catch (Throwable)`.
- Every error message names the operation and the offending input. Prefer `read config %s: %w` over `error reading file`.
