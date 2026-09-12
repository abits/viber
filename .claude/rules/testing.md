# Testing

- Every new function or type gets a test unless it is a pure wiring shim.
- Mock external dependencies: network, filesystem beyond the test framework's temp-dir helper, databases, subprocess calls.
- Arrange-Act-Assert layout — visually or with blank lines. One behaviour per test case.
- **Never run a test snippet you just generated without first saving it to a discrete file.** Ad-hoc `-e` / `-c` runs are non-reproducible and leave no trace.
- **Never delete a file created as part of testing.** If cleanup is needed, use the framework's temp-dir helper (`t.TempDir`, `tmp_path`, etc.).
- Any folder that holds test output artifacts must be listed in `.gitignore`.
- Never commit commented-out tests. Delete them or mark them skipped with a reason.
