# 当前服务 module（必须与你 services/option/go.mod 的 module 一致）
MODULE := wklive/services/option

# 仓库根目录（从 services/option 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/option 模块）
PROTO_DIR := $(REPO_ROOT)/proto/option


PROTO_FILE := $(PROTO_DIR)/option.proto

GOCTL ?= $(shell go env GOPATH)/bin/goctl
MODEL_DDL := option.sql

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

# 所有 RPC service 统一在 option.proto 中，以 multiple 模式分组生成。
gen-rpc:
	@echo "🔧 [services/option] goctl rpc protoc..."
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
	@echo "✅ [services/option] gen-rpc done"

# goctl 在服务目录生成的协议代码只用于生成 zrpc 桩，clean-gen 会删除该临时目录。
# 显式同步仓库共享的 proto Go 代码，避免 .proto 与 proto/option/*.pb.go 漂移。
sync-proto:
	@command -v protoc >/dev/null || (echo "❌ protoc not found"; exit 1)
	@command -v protoc-gen-go >/dev/null || (echo "❌ protoc-gen-go not found"; exit 1)
	@command -v protoc-gen-go-grpc >/dev/null || (echo "❌ protoc-gen-go-grpc not found"; exit 1)
	@echo "🔧 [services/option] syncing shared proto Go code..."
	@cd $(REPO_ROOT) && protoc --proto_path=. \
		--go_out=. --go_opt=module=wklive \
		--go-grpc_out=. --go-grpc_opt=module=wklive \
		proto/option/enum.proto proto/option/model.proto proto/option/option.proto
	@echo "✅ [services/option] shared proto Go code synced"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/option] fixing imports..."
ifeq ($(shell uname -s),Darwin)
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"$(MODULE)/wklive/proto/option"#"wklive/proto/option"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"wklive/services/option/wklive/proto/option"#"wklive/proto/option"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's/option_option\./option\./g' {} +
else
	@find . -name "*.go" -type f -exec sed -i -e 's#"$(MODULE)/wklive/proto/option"#"wklive/proto/option"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's#"wklive/services/option/wklive/proto/option"#"wklive/proto/option"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's/option_option\./option\./g' {} +
endif
	@echo "✅ [services/option] imports fixed"

fmt:
	@echo "🔧 [services/option] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/option] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

clean-main:
	@rm -rf optionadmin.go optionapp.go optioninternal.go optiontask.go ./etc/optionadmin.yaml ./etc/optionapp.yaml ./etc/optioninternal.yaml ./etc/optiontask.yaml

.PHONY: gen-model
gen-model: $(MODEL_DDL)
	@$(MAKE) check-goctl
	@echo "🔧 [services/option] generating decimal-safe models..."
	@$(GOCTL) model mysql ddl --src $(MODEL_DDL) --dir ./models -c
	@echo "✅ [services/option] models generated (DECIMAL -> decimal.Decimal)"
