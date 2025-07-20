

if [ -z "$GH_TOKEN" ]; then
  echo "Please set the GH_TOKEN environment variable to your GitHub OIDC access token."
  exit 1
fi

KUBECONFIG=./examples/auth/kubeconfig \
  kubectl config set-credentials testuser \
  --auth-provider=oidc \
  --auth-provider-arg=client-id=example-app  \
  --auth-provider-arg=client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0  \
  --auth-provider-arg=idp-issuer-url=https://host.docker.internal:5553/dex  \
  --auth-provider-arg=idp-certificate-authority=examples/auth/caddy-pki/intermediate.crt \
  --auth-provider-arg=id-token=${GH_TOKEN}


KUBECONFIG=./examples/auth/kubeconfig \
  kubectl config set-context test-context \
  --user=testuser \
  --cluster=k3d-mcp-test-auth

KUBECONFIG=./examples/auth/kubeconfig \
  kubectl config use-context test-context