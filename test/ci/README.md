## End-to-end test

These tests are marked with the `integration` build tag, which means they are excluded from normal test runs.

### Running the integration tests

To run these tests (requires a working Kubernetes cluster):

```bash
go test -tags=integration ./test/ci/ -timeout 60m -v
```

### Running a specific test

```bash
go test -tags=integration ./test/ci/ -timeout 60m -v -run TestSecretTemplate_Full_Lifecycle
```

See `./test/ci/env.go` for required environment variables for some tests.
