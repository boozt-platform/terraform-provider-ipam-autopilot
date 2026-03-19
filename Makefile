# Copyright 2026 Boozt Fashion AB
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0

REPO_ROOT    := $(shell pwd)
PROVIDER_DIR := $(REPO_ROOT)/provider
LOCAL_DEV_DIR := $(REPO_ROOT)/examples/local-dev

.PHONY: help lint test test-integration build-provider dev-setup dev-plan dev-apply dev-destroy

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

## ── Code quality ────────────────────────────────────────────────────────────

lint: ## Run golangci-lint on container and provider
	cd $(REPO_ROOT)/container && golangci-lint run ./...
	cd $(REPO_ROOT)/provider  && golangci-lint run ./...

fmt: ## Format all Go code
	cd $(REPO_ROOT)/container && gofmt -w .
	cd $(REPO_ROOT)/provider  && gofmt -w .

## ── Tests ───────────────────────────────────────────────────────────────────

test: ## Run unit tests (container + provider)
	cd $(REPO_ROOT)/container && go test ./...
	cd $(REPO_ROOT)/provider  && go test ./...

test-integration: ## Run integration tests (requires Docker)
	cd $(REPO_ROOT)/container && go test -tags=integration -timeout=10m ./tests/...

## ── Local development ───────────────────────────────────────────────────────

build-provider: ## Build the Terraform provider binary
	cd $(PROVIDER_DIR) && go build -o terraform-provider-ipam-autopilot .

dev-setup: build-provider ## Build provider and generate dev.tfrc
	@echo 'provider_installation {'                                           > $(LOCAL_DEV_DIR)/dev.tfrc
	@echo '  dev_overrides {'                                                >> $(LOCAL_DEV_DIR)/dev.tfrc
	@echo '    "boozt-platform/ipam-autopilot" = "$(PROVIDER_DIR)"'         >> $(LOCAL_DEV_DIR)/dev.tfrc
	@echo '  }'                                                              >> $(LOCAL_DEV_DIR)/dev.tfrc
	@echo '  direct {}'                                                      >> $(LOCAL_DEV_DIR)/dev.tfrc
	@echo '}'                                                                >> $(LOCAL_DEV_DIR)/dev.tfrc
	@echo ""
	@echo "Ready. Run:"
	@echo "  cd $(LOCAL_DEV_DIR)"
	@echo "  export TF_CLI_CONFIG_FILE=./dev.tfrc GCP_IDENTITY_TOKEN=localdev"
	@echo "  terraform plan"

dev-plan: dev-setup ## terraform plan against local docker-compose stack
	cd $(LOCAL_DEV_DIR) && TF_CLI_CONFIG_FILE=./dev.tfrc GCP_IDENTITY_TOKEN=localdev terraform plan

dev-apply: dev-setup ## terraform apply against local docker-compose stack
	cd $(LOCAL_DEV_DIR) && TF_CLI_CONFIG_FILE=./dev.tfrc GCP_IDENTITY_TOKEN=localdev terraform apply -auto-approve

dev-destroy: dev-setup ## terraform destroy local dev resources
	cd $(LOCAL_DEV_DIR) && TF_CLI_CONFIG_FILE=./dev.tfrc GCP_IDENTITY_TOKEN=localdev terraform destroy -auto-approve
