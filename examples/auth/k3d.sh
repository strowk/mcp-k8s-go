#!/bin/bash

#   --k3s-arg "--kube-apiserver-arg=--oidc-issuer-url=https://caddy:443/dex@server:*" \

k3d cluster create \
  --k3s-arg "--kube-apiserver-arg=--oidc-issuer-url=https://host.docker.internal:5553/dex@server:*" \
  --k3s-arg "--kube-apiserver-arg=--oidc-username-claim=email@server:*" \
  --k3s-arg "--kube-apiserver-arg=--oidc-groups-claim=groups@server:*" \
  --k3s-arg "--kube-apiserver-arg=--oidc-username-prefix=oidc:@server:*" \
  --k3s-arg "--kube-apiserver-arg=--oidc-groups-prefix=oidc:@server:*" \
  --k3s-arg "--kube-apiserver-arg=--oidc-client-id=example-app@server:*" \
  --k3s-arg "--kube-apiserver-arg=--oidc-ca-file=/caddy-pki/intermediate.crt@server:*" \
  --network auth_default \
  --volume "$(pwd -W)/caddy-pki:/caddy-pki@server:*" \
  mcp-test-auth

# curl https://caddy:443/dex/.well-known/openid-configuration

