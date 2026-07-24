# 当前服务 module（必须与你 services/trade/go.mod 的 module 一致）
MODULE := wklive/services/trade

# 仓库根目录（从 services/trade 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/trade 模块）
PROTO_DIR := $(REPO_ROOT)/proto/trade


PROTO_FILE := $(PROTO_DIR)/trade.proto

GOCTL ?= $(shell go env GOPATH)/bin/goctl

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

# 所有 RPC service 统一在 trade.proto 中，以 multiple 模式分组生成。
gen-rpc:
	@echo "🔧 [services/trade] goctl rpc protoc..."
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
	@echo "✅ [services/trade] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/trade] fixing imports..."
ifeq ($(shell uname -s),Darwin)
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"$(MODULE)/wklive/proto/trade"#"wklive/proto/trade"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"wklive/services/trade/wklive/proto/trade"#"wklive/proto/trade"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's/trade_trade\./trade\./g' {} +
else
	@find . -name "*.go" -type f -exec sed -i -e 's#"$(MODULE)/wklive/proto/trade"#"wklive/proto/trade"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's#"wklive/services/trade/wklive/proto/trade"#"wklive/proto/trade"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's/trade_trade\./trade\./g' {} +
endif
	@echo "✅ [services/trade] imports fixed"

fmt:
	@echo "🔧 [services/trade] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/trade] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

clean-main:
	@rm -rf tradeadmin.go tradeapp.go tradeinternal.go tradetask.go ./etc/tradeadmin.yaml ./etc/tradeapp.yaml ./etc/tradeinternal.yaml ./etc/tradetask.yaml

.PHONY: gen-model
gen-model:
	@$(MAKE) check-goctl
	@echo "🔧 [services/trade] generating decimal-safe models..."
	@$(GOCTL) model mysql ddl --src trade.sql --dir ./models -c
	@echo "✅ [services/trade] models generated (DECIMAL -> decimal.Decimal)"
