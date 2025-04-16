FROM gcr.io/distroless/static:nonroot

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG TARGETVARIANT=""

WORKDIR /

COPY build/templated-secret-controller_${TARGETOS}_${TARGETARCH}*/templated-secret-controller /templated-secret-controller

# Use nonroot user from distroless image
USER 65532:65532

ENTRYPOINT ["/templated-secret-controller"]