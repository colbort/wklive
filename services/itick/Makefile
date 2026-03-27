GOCTL ?= goctl

# 当前服务 module（必须与你 services/itick/go.mod 的 module 一致）
MODULE := wklive/services/itick

# 仓库根目录（从 services/itick 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/itick 模块）
PROTO_DIR := $(REPO_ROOT)/proto/itick

# 自动查找所有 proto 文件
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

.PHONY: gen gen-rpc fix-imports tidy fmt clean-gen gen-model print-protos

gen: gen-rpc fix-imports fmt tidy clean-gen

# 查看匹配到的 proto 文件
print-protos:
	@echo "PROTO_DIR=$(PROTO_DIR)"
	@echo "PROTO_FILES=$(PROTO_FILES)"

# 一键：只生成包含 service 的 proto
gen-rpc:
	@echo "🔧 [services/itick] goctl rpc protoc..."
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
	@echo "✅ [services/itick] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/itick] fixing imports..."
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's#"$(MODULE)/wklive/proto/itick"#"wklive/proto/itick"#g'
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's#"wklive/services/itick/wklive/proto/itick"#"wklive/proto/itick"#g'
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's/itick_itick\./itick\./g'
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

.PHONY: gen-model
gen-model:
	goctl model mongo --type CoinKline --dir ./models