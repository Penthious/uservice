SHELL := /bin/bash
BUILD_REF := "local"

run:
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

build:
	go build -ldflags "-X main.build=${BUILD_REF}"

admin:
	go run app/tooling/admin/main.go

test:
	go test ./... -count=1
	staticcheck -checks=all ./...
# ======================================================================================================================
# Building Containers
VERSION := 1.0

all: sales

sales:
	docker build \
		-f zarf/docker/Dockerfile.sales-api \
		-t sales-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ======================================================================================================================
# Running from within k8s/kind
KIND_CLUSTER := starter-cluster

kind-up:
	kind create cluster \
		--image kindest/node:v1.23.0@sha256:49824ab1727c04e56a21a5d8372a402fcd32ea51ac96a2706a12af38934f81ac \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/kind-config.yaml
	kubectl config set-context --current --namespace=sales-system

kind-load:
	cd zarf/k8s/kind/sales-pod; kustomize edit set image sales-api-image=sales-api-amd64:$(VERSION)
	kind load docker-image sales-api-amd64:$(VERSION) --name $(KIND_CLUSTER)

kind-apply:
	kustomize build zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-describe:
	kubectl describe pod -l app=sales

kind-restart:
	kubectl rollout restart deployment sales-pod

kind-update: all kind-load kind-restart
kind-update-apply: all kind-load kind-apply kind-restart

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

# ======================================================================================================================
# Modules support
tidy:
	go mod tidy
	go mod vendor
