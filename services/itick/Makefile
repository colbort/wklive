# 当前服务 module（必须与你 services/itick/go.mod 的 module 一致）
MODULE := wklive/services/itick

# 仓库根目录（从 services/itick 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/itick 模块）
PROTO_DIR := $(REPO_ROOT)/proto/itick

PROTO_FILE := $(PROTO_DIR)/itick.proto

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
	@echo "PROTO_FILE=$(PROTO_FILE)"

# 四个 RPC service 统一在 itick.proto 中，以 multiple 模式分组生成。
gen-rpc:
	@echo "🔧 [services/itick] goctl rpc protoc..."
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
	@echo "✅ [services/itick] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/itick] fixing imports..."
ifeq ($(shell uname -s),Darwin)
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"$(MODULE)/wklive/proto/itick"#"wklive/proto/itick"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"wklive/services/itick/wklive/proto/itick"#"wklive/proto/itick"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's/itick_itick\./itick\./g' {} +
else
	@find . -name "*.go" -type f -exec sed -i -e 's#"$(MODULE)/wklive/proto/itick"#"wklive/proto/itick"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's#"wklive/services/itick/wklive/proto/itick"#"wklive/proto/itick"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's/itick_itick\./itick\./g' {} +
endif
	@echo "✅ [services/itick] imports fixed"

fmt:
	@echo "🔧 [services/itick] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/itick] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

clean-main:
	@rm -rf itickadmin.go itickapp.go itickinternal.go iticktask.go ./etc/itickadmin.yaml ./etc/itickapp.yaml ./etc/itickinternal.yaml ./etc/iticktask.yaml

.PHONY: gen-model
gen-model:
	@$(MAKE) check-goctl
	@echo "🔧 [services/itick] generating decimal-safe models..."
	@$(GOCTL) model mysql ddl --src itick.sql --dir ./models -c
	@echo "✅ [services/itick] models generated (DECIMAL -> decimal.Decimal)"
