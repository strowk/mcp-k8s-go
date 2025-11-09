# Example of setup authentication with dex, k3d and mcp-k8s

## Pre-clean

```bash
rm examples/auth/caddy-pki/*.crt
rm examples/auth/caddy-pki/*.key
```

## Configure dex

Go to Github, create new OAuth app, get client ID and secret.

<!-- TODO: clean up this, most likely not needed part about github, since example does not include teams support anymore -->

Then create `dex-config.yaml` file in `examples/auth/` directory based on `dex-config-template.yaml`:

```bash
cp examples/auth/dex-config-template.yaml examples/auth/dex-config.yaml
```
and fill dex-config.yaml with your client ID and secret after this block:

```yaml
connectors:
  - type: github
    id: github
    name: GitHub
    config:
```

## Start dex

```bash
cd examples/auth/
docker compose up
```

## Start k3d and configure

```bash
# k3d cluster delete mcp-test-auth # uncomment to delete existing cluster
cd examples/auth/
./k3d.sh
kubectl apply -f ./k8s-manifest.yaml
```

## Run server

```bash
go run main.go \
  --oidc-client-id=example-app \
  --oidc-client-secret=ZXhhbXBsZS1hcHAtc2VjcmV0 \
  --oidc-redirect-url=http://localhost:8080/callback \
  --oidc-auth-url=http://localhost:5556/dex/auth \
  --oidc-token-url=http://localhost:5556/dex/token \
  --remote-hostname=localhost \
  --remote-port=8080 

```

Wait till the server has started on 8080.

## Run example client

```bash
cd examples/auth/example_client
go run main.go
```

Follow URL, allow access.

<!-- ## Try access token

Grab access token from the server log, put it to GH_TOKEN env var (`export GH_TOKEN=`), then:

```bash
examples/auth/configure.sh
```

Now run `code ~/.kube/config`, grab cluster with name `k3d-mcp-test-auth`, and replace cluster in `kubeconfig`.

Now you can access the cluster with `kubectl`:

```bash
KUBECONFIG=examples/auth/kubeconfig kubectl get pods -n default
KUBECONFIG=examples/auth/kubeconfig kubectl get pods -n kube-system # this should fail unless you added kube-system access
```

``` -->

Change `k8s-manifest.yaml` to your needs (f.e add kube-system access), then reapply it:

```bash
kubectl apply -f ./k8s-manifest.yaml

Delete the rolebinding if you want to remove kube-system access:

```bash
kubectl delete -n kube-system rolebinding admin-test-team
```

## Cleanup

Stop all running processes in terminals, then remove k3d cluster:

```bash
k3d cluster delete mcp-test-auth
```