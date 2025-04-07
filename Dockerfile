FROM gcr.io/distroless/static-debian12
USER nonroot:nonroot
COPY --chown=nonroot:nonroot mcp-k8s-go /
ENTRYPOINT ["/mcp-k8s-go"]
