#/bin/bash

go install go.uber.org/mock/mockgen@latest
mockgen -source=internal/k8s/pool.go -destination=internal/k8s/mock/pool_mock.go