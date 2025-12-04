# Weebcast Operator Makefile

# Image URL to use all building/pushing image targets
IMG ?= weebcast/weebcast-operator:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: ## Run golangci-lint against code.
	golangci-lint run

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for cross-platform support
	docker buildx build --platform linux/amd64,linux/arm64 -t ${IMG} --push .

##@ Deployment

.PHONY: install
install: ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/crd/

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/
	kubectl apply -f config/rbac/
	kubectl apply -f config/manager/

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/manager/
	kubectl delete -f config/rbac/
	kubectl delete -f config/crd/

.PHONY: deploy-samples
deploy-samples: ## Deploy sample AnimeMonitor resources.
	kubectl apply -f config/samples/

.PHONY: undeploy-samples
undeploy-samples: ## Remove sample AnimeMonitor resources.
	kubectl delete -f config/samples/

##@ Local Development

.PHONY: dev
dev: install ## Run operator, API worker, and frontend locally.
	@echo "🚀 Starting local development environment..."
	@echo ""
	@echo "Starting services:"
	@echo "  - Operator:  go run cmd/main.go"
	@echo "  - API:       http://localhost:8787"
	@echo "  - Frontend:  http://localhost:8000"
	@echo ""
	@trap 'kill 0' EXIT; \
	go run cmd/main.go & \
	(cd website/worker && wrangler dev --local --port 8787) & \
	(cd website/frontend && python3 -m http.server 8000) & \
	wait

.PHONY: dev-operator
dev-operator: install ## Run only the operator locally.
	go run cmd/main.go

.PHONY: dev-api
dev-api: ## Run only the Cloudflare Worker API locally.
	cd website/worker && wrangler dev --local --port 8787

.PHONY: dev-frontend
dev-frontend: ## Run only the frontend locally.
	@echo "Frontend available at http://localhost:8000"
	cd website/frontend && python3 -m http.server 8000

.PHONY: sync
sync: ## Sync K8s AnimeMonitor data to local API worker.
	@./scripts/sync-to-local.sh

.PHONY: dev-watch
dev-watch: ## Run sync in a loop (every 30s) to keep local API updated.
	@echo "Watching for changes and syncing every 30s..."
	@while true; do ./scripts/sync-to-local.sh; sleep 30; done

##@ Utilities

.PHONY: mod-tidy
mod-tidy: ## Run go mod tidy.
	go mod tidy

.PHONY: mod-download
mod-download: ## Run go mod download.
	go mod download

.PHONY: clean
clean: ## Clean build artifacts.
	rm -rf bin/
	rm -f cover.out

.PHONY: logs
logs: ## View operator logs.
	kubectl logs -f -n weebcast-system deployment/weebcast-operator

.PHONY: status
status: ## Check AnimeMonitor status.
	kubectl get animemonitors -A -o wide

