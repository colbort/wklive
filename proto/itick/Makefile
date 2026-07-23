.PHONY: gen clean check-protoc install-protoc-plugins

REPO_ROOT := ../..
GOBIN := $(shell go env GOPATH)/bin
PROTOC_GEN_GO := $(GOBIN)/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(GOBIN)/protoc-gen-go-grpc

gen: check-protoc
	@echo "🔧 [proto/itick] generating itick pb..."
	@rm -f itick_admin*.pb.go itick_app*.pb.go itick_internal*.pb.go itick_task*.pb.go
	cd $(REPO_ROOT) && PATH="$(GOBIN):$$PATH" protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/itick/enum.proto \
		proto/itick/model.proto \
		proto/itick/itick.proto
	@echo "✅ [proto/itick] generated"

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
