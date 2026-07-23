# 当前服务 module（必须与你 services/staking/go.mod 的 module 一致）
MODULE := wklive/services/staking

# 仓库根目录（从 services/staking 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/staking 模块）
PROTO_DIR := $(REPO_ROOT)/proto/staking


PROTO_FILE := $(PROTO_DIR)/staking.proto

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

# 所有 RPC service 统一在 staking.proto 中，以 multiple 模式分组生成。
gen-rpc:
	@echo "🔧 [services/staking] goctl rpc protoc..."
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
	@echo "✅ [services/staking] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/staking] fixing imports..."
ifeq ($(shell uname -s),Darwin)
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"$(MODULE)/wklive/proto/staking"#"wklive/proto/staking"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's#"wklive/services/staking/wklive/proto/staking"#"wklive/proto/staking"#g' {} +
	@find . -name "*.go" -type f -exec sed -i '' -e 's/staking_staking\./staking\./g' {} +
else
	@find . -name "*.go" -type f -exec sed -i -e 's#"$(MODULE)/wklive/proto/staking"#"wklive/proto/staking"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's#"wklive/services/staking/wklive/proto/staking"#"wklive/proto/staking"#g' {} +
	@find . -name "*.go" -type f -exec sed -i -e 's/staking_staking\./staking\./g' {} +
endif
	@echo "✅ [services/staking] imports fixed"

fmt:
	@echo "🔧 [services/staking] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/staking] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

clean-main:
	@rm -rf stakingadmin.go stakingapp.go stakingtask.go ./etc/stakingadmin.yaml ./etc/stakingapp.yaml ./etc/stakingtask.yaml

.PHONY: gen-model
gen-model:
	@$(MAKE) check-goctl
	@echo "🔧 [services/staking] generating decimal-safe models..."
	@$(GOCTL) model mysql ddl --src staking.sql --dir ./models -c
	@echo "✅ [services/staking] models generated (DECIMAL -> decimal.Decimal)"
