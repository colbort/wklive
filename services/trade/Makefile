GOCTL ?= goctl

# 当前服务 module（必须与你 services/trade/go.mod 的 module 一致）
MODULE := wklive/services/trade

# 仓库根目录（从 services/trade 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/trade 模块）
PROTO_DIR := $(REPO_ROOT)/proto/trade

# 自动查找所有 proto 文件
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

GOCTL ?= $(shell go env GOPATH)/bin/goctl

.PHONY: gen gen-rpc fix-imports tidy fmt clean-gen gen-model print-protos clean-main install-goctl

gen: check-goctl gen-rpc fix-imports fmt tidy clean-gen clean-main

check-goctl:
	@$(GOCTL) --version | grep "1.9.2" >/dev/null || \
		(echo "❌ 当前项目要求 goctl v1.9.2"; exit 1)

install-goctl:
	go install github.com/zeromicro/go-zero/tools/goctl@v1.9.2

# 查看匹配到的 proto 文件
print-protos:
	@echo "PROTO_DIR=$(PROTO_DIR)"
	@echo "PROTO_FILES=$(PROTO_FILES)"

# 一键：只生成包含 service 的 proto
gen-rpc:
	@echo "🔧 [services/trade] goctl rpc protoc..."
	@for file in $(PROTO_FILES); do \
		if grep -Eq '^[[:space:]]*service[[:space:]]+[A-Za-z0-9_]+' $$file; then \
			echo "➡️ generating rpc for $$file"; \
			$(GOCTL) rpc protoc $$file \
				--proto_path=$(REPO_ROOT) \
				--module=$(MODULE) \
				--go_out=. \
				--go-grpc_out=. \
				--zrpc_out=. \
				--style=gozero || exit $$?; \
		else \
			echo "⏭️ skip(no service): $$file"; \
		fi; \
	done
	@$(MAKE) fix-imports
	@$(MAKE) fmt
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
	@rm -rf tradeadmin.go tradeapp.go tradeinternal.go ./etc/tradeadmin.yaml ./etc/tradeapp.yaml ./etc/tradeinternal.yaml

.PHONY: gen-model
gen-model:
	# goctl model mongo --type CoinKline --dir ./models
	goctl model mysql ddl --src trade.sql --dir ./models -c