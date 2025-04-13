## End-to-end test

These tests are now run directly through the GitHub Actions workflow.

To run manually:

```bash
go test ./test/ci/ -timeout 60m -v
# Or to run a specific test:
go test ./test/ci/ -timeout 60m -v -run TestSecretTemplate_Full_Lifecycle
```

See `./test/ci/env.go` for required environment variables for some tests.
