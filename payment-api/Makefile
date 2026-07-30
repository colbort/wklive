API_FILE := api/core.api
GOCTL ?= $(shell go env GOPATH)/bin/goctl

.PHONY: gen check fmt verify fmt-api

gen: check
	@echo "🔧 generating from $(API_FILE) ..."
	@$(GOCTL) api go -api $(API_FILE) -dir .
	@$(MAKE) fmt
	@GOCACHE=/private/tmp/wklive-go-build-cache go mod tidy
	@echo "✅ generate done"

check:
	@test -f $(API_FILE) || (echo "❌ missing api file: $(API_FILE)"; exit 1)

fmt:
	@gofmt -w $$(find . -name "*.go" -type f)

verify:
	@test -f payment.go
	@test -f internal/handler/routes.go
	@test -f internal/types/types.go

fmt-api:
	goctl api format --dir ./api/
