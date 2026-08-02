# 当前服务 module（必须与你 services/asset/go.mod 的 module 一致）
MODULE := wklive/services/asset

# 仓库根目录（从 services/asset 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/asset 模块）
PROTO_DIR := $(REPO_ROOT)/proto/asset


PROTO_FILE := $(PROTO_DIR)/asset.proto

GOCTL ?= $(shell go env GOPATH)/bin/goctl

.PHONY: gen gen-rpc sync-proto fix-imports tidy fmt clean-gen gen-model print-protos clean-main install-goctl

gen: check-goctl gen-rpc sync-proto fix-imports fmt tidy clean-gen clean-main

check-goctl:
	@$(GOCTL) --version | grep "1.9.2" >/dev/null || \
		(echo "❌ 当前项目要求 goctl v1.9.2"; exit 1)

install-goctl:
	go install github.com/zeromicro/go-zero/tools/goctl@v1.9.2

# 查看匹配到的 proto 文件
print-protos:
	@echo "PROTO_DIR=$(PROTO_DIR)"
	@echo "PROTO_FILE=$(PROTO_FILE)"

# 所有 RPC service 统一在 asset.proto 中，以 multiple 模式分组生成。
gen-rpc:
	@echo "🔧 [services/asset] goctl rpc protoc..."
	@$(GOCTL) rpc protoc $(PROTO_FILE) \
		--multiple \
		--proto_path=$(REPO_ROOT) \
		--module=$(MODULE) \
		--go_out=. \
		--go-grpc_out=. \
		--zrpc_out=. \
		--style=gozero
	@$(MAKE) fix-imports
	@$(MAKE) fmt
	@$(MAKE) tidy
	@echo "✅ [services/asset] gen-rpc done"

# goctl 在服务目录生成的协议代码只用于生成 zrpc 桩，clean-gen 会删除该临时目录。
# 显式同步仓库共享的 proto Go 代码，避免 .proto 与 proto/asset/*.pb.go 漂移。
sync-proto:
	@command -v protoc >/dev/null || (echo "❌ protoc not found"; exit 1)
	@command -v protoc-gen-go >/dev/null || (echo "❌ protoc-gen-go not found"; exit 1)
	@command -v protoc-gen-go-grpc >/dev/null || (echo "❌ protoc-gen-go-grpc not found"; exit 1)
	@echo "🔧 [services/asset] syncing shared proto Go code..."
	@cd $(REPO_ROOT) && protoc --proto_path=. \
		--go_out=. --go_opt=module=wklive \
		--go-grpc_out=. --go-grpc_opt=module=wklive \
		proto/asset/enum.proto proto/asset/model.proto proto/asset/asset.proto
	@echo "✅ [services/asset] shared proto Go code synced"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/asset] fixing imports..."
ifeq ($(shell uname -s),Darwin)
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"$(MODULE)/wklive/proto/asset"#"wklive/proto/asset"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"wklive/services/asset/wklive/proto/asset"#"wklive/proto/asset"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's/asset_asset\./asset\./g' {} +
else
	@find . -name "*.go" -type f -exec sed -i -e 's#"$(MODULE)/wklive/proto/asset"#"wklive/proto/asset"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's#"wklive/services/asset/wklive/proto/asset"#"wklive/proto/asset"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's/asset_asset\./asset\./g' {} +
endif
	@echo "✅ [services/asset] imports fixed"

fmt:
	@echo "🔧 [services/asset] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/asset] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

clean-main:
	@rm -rf assetadmin.go assetapp.go assetinternal.go ./etc/assetadmin.yaml ./etc/assetapp.yaml ./etc/assetinternal.yaml

.PHONY: gen-model
gen-model:
	@$(MAKE) check-goctl
	@echo "🔧 [services/asset] generating decimal-safe models..."
	@$(GOCTL) model mysql ddl --src asset.sql --dir ./models -c
	@echo "✅ [services/asset] models generated (DECIMAL -> decimal.Decimal)"
