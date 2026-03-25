GOCTL ?= goctl

# 当前服务 module（必须与你 services/user/go.mod 的 module 一致）
MODULE := wklive/services/user

# 仓库根目录（从 services/user 往上两级）
REPO_ROOT := ../..

# proto 目录（属于 proto/user 模块）
PROTO_DIR := $(REPO_ROOT)/proto/user

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
	@echo "🔧 [services/user] goctl rpc protoc..."
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
	@echo "✅ [services/user] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/user] fixing imports..."
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's#"$(MODULE)/wklive/proto/user"#"wklive/proto/user"#g'
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's#"wklive/services/user/wklive/proto/user"#"wklive/proto/user"#g'
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's/user_user\./user\./g'
	@echo "✅ [services/user] imports fixed"

fmt:
	@echo "🔧 [services/user] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/user] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

.PHONY: gen-model
gen-model:
	goctl model mysql ddl --src user.sql --dir ./models -c