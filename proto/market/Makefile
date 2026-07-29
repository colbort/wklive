.PHONY: gen clean check-protoc install-protoc-plugins

REPO_ROOT := ../..
GOBIN := $(shell go env GOPATH)/bin
PROTOC_GEN_GO := $(GOBIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(GOBIN)/protoc-gen-go-grpc

gen: check-protoc
	@echo "🔧 [proto/market] generating market pb..."
	@rm -f market_admin*.pb.go market_app*.pb.go market_internal*.pb.go market_task*.pb.go
	cd $(REPO_ROOT) && PATH="$(GOBIN):$$PATH" protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/market/enum.proto \
		proto/market/model.proto \
		proto/market/market.proto
	@echo "✅ [proto/market] generated"

check-protoc:
	@command -v protoc >/dev/null || (echo "❌ protoc 未安装"; exit 1)
	@$(PROTOC_GEN_GO) --version >/dev/null 2>&1 || \
		(echo "❌ protoc-gen-go 未安装，请执行: make install-protoc-plugins"; exit 1)
	@$(PROTOC_GEN_GO_GRPC) --version >/dev/null 2>&1 || \
		(echo "❌ protoc-gen-go-grpc 未安装，请执行: make install-protoc-plugins"; exit 1)

install-protoc-plugins:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

clean:
	rm -f *.pb.go
