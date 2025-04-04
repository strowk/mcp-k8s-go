FROM gcr.io/distroless/static-debian12
ENTRYPOINT ["/mcp-k8s-go"]
COPY mcp-k8s-go /
