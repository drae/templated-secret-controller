FROM gcr.io/distroless/static:nonroot

# Goreleaser supplies the already-built architecture-specific binary directly in the Docker build context
# under the fixed name 'templated-secret-controller'. No shell or build tools needed.
COPY templated-secret-controller /templated-secret-controller

USER 65532:65532
ENTRYPOINT ["/templated-secret-controller"]