#!/usr/bin/env bash
set -euo pipefail

echo "Installing development tools..."

# Detect architecture
ARCH=$(dpkg --print-architecture)          # amd64 | arm64
HADOLINT_ARCH="${ARCH}"
if [ "${ARCH}" = "amd64" ]; then
  HADOLINT_ARCH="x86_64"
fi

# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin"

# go tools
go install golang.org/x/tools/cmd/goimports@latest

# hadolint
HADOLINT_VERSION="v2.14.0"
curl -sSfL "https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/hadolint-Linux-${HADOLINT_ARCH}" \
  -o /usr/local/bin/hadolint
chmod +x /usr/local/bin/hadolint

# opentofu
TOFU_VERSION="1.11.5"
curl -sSfL "https://github.com/opentofu/opentofu/releases/download/v${TOFU_VERSION}/tofu_${TOFU_VERSION}_linux_${ARCH}.tar.gz" \
  | tar -xz -C /usr/local/bin tofu

# terraform-docs
TFDOCS_VERSION="v0.20.0"
curl -sSfL "https://github.com/terraform-docs/terraform-docs/releases/download/${TFDOCS_VERSION}/terraform-docs-${TFDOCS_VERSION}-linux-${ARCH}.tar.gz" \
  | tar -xz -C /usr/local/bin terraform-docs

# pre-commit
apt-get update -qq
apt-get install -y --no-install-recommends python3-pip > /dev/null 2>&1
pip3 install --quiet --break-system-packages pre-commit

echo "All tools installed."
echo ""
echo "Available make targets:"
echo "  make lint             - golangci-lint + hadolint"
echo "  make test             - unit tests + tofu test"
echo "  make test-integration - integration tests (requires Docker)"
echo "  make dev-apply        - build provider + terraform apply against docker-compose"
echo "  make dev-destroy      - terraform destroy"
echo ""
echo "Run 'pre-commit install' to enable git hooks."
