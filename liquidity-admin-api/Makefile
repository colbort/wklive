API_FILE := api/core.api
GOCTL ?= $(shell go env GOPATH)/bin/goctl
GOCACHE ?= /private/tmp/wklive-go-build-cache
export GOCACHE

.PHONY: gen check-goctl fmt tidy verify

gen: check-goctl
	@echo "🔧 [liquidity-admin-api] generating..."
	@$(GOCTL) api go -api $(API_FILE) -dir . --style=gozero
	@$(MAKE) fmt
	@$(MAKE) tidy
	@$(MAKE) verify

check-goctl:
	@$(GOCTL) --version | grep "1.9.2" >/dev/null || \
		(echo "❌ 当前项目要求 goctl v1.9.2"; exit 1)

fmt:
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@go mod tidy

verify:
	@go test ./...
