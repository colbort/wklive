# 当前服务 module（必须与你 services/liquidity/go.mod 的 module 一致）
MODULE := wklive/services/liquidity

# 仓库根目录（从 services/liquidity 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/liquidity 模块）
PROTO_DIR := $(REPO_ROOT)/proto/liquidity

PROTO_FILE := $(PROTO_DIR)/liquidity.proto

GOCTL ?= $(shell go env GOPATH)/bin/goctl
GOCACHE ?= /private/tmp/wklive-go-build-cache
export GOCACHE

.PHONY: gen gen-rpc fix-imports tidy fmt clean-gen gen-model print-protos clean-main install-goctl

gen: check-goctl gen-rpc

check-goctl:
	@$(GOCTL) --version | grep "1.9.2" >/dev/null || \
		(echo "❌ 当前项目要求 goctl v1.9.2"; exit 1)

install-goctl:
	go install github.com/zeromicro/go-zero/tools/goctl@v1.9.2

# 查看匹配到的 proto 文件
print-protos:
	@echo "PROTO_DIR=$(PROTO_DIR)"
	@echo "PROTO_FILE=$(PROTO_FILE)"

# 所有 RPC service 统一在 liquidity.proto 中，以 multiple 模式分组生成。
gen-rpc:
	@echo "🔧 [services/liquidity] goctl rpc protoc..."
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
	@$(MAKE) clean-gen
	@$(MAKE) clean-main
	@$(MAKE) tidy
	@echo "✅ [services/liquidity] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/liquidity] fixing imports..."
ifeq ($(shell uname -s),Darwin)
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"$(MODULE)/wklive/proto/liquidity"#"wklive/proto/liquidity"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"wklive/services/liquidity/wklive/proto/liquidity"#"wklive/proto/liquidity"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's/liquidity_liquidity\./liquidity\./g' {} +
else
	@find . -name "*.go" -type f -exec sed -i -e 's#"$(MODULE)/wklive/proto/liquidity"#"wklive/proto/liquidity"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's#"wklive/services/liquidity/wklive/proto/liquidity"#"wklive/proto/liquidity"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's/liquidity_liquidity\./liquidity\./g' {} +
endif
	@echo "✅ [services/liquidity] imports fixed"

fmt:
	@echo "🔧 [services/liquidity] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/liquidity] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删手写 logic）
clean-gen:
	@rm -rf wklive/

clean-main:
	@rm -rf liquidityadmin.go liquidityapp.go liquidityinternal.go liquiditytask.go ./etc/liquidityadmin.yaml ./etc/liquidityapp.yaml ./etc/liquidityinternal.yaml ./etc/liquiditytask.yaml

.PHONY: gen-model
gen-model:
	@$(MAKE) check-goctl
	@echo "🔧 [services/liquidity] generating decimal-safe models..."
	@$(GOCTL) model mysql ddl --src liquidity.sql --dir ./models -c
	@echo "✅ [services/liquidity] models generated (DECIMAL -> decimal.Decimal)"
