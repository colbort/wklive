GOCTL ?= goctl

# 当前服务 module（必须与你 services/system/go.mod 的 module 一致）
MODULE := wklive/services/system

# 仓库根目录（从 services/system 往上两级）
REPO_ROOT := ../..

# proto 所在位置（属于 rpc/system 模块）
PROTO := $(REPO_ROOT)/rpc/system/system.proto

.PHONY: gen gen-rpc fix-imports tidy fmt clean-gen

gen: gen-rpc fix-imports fmt tidy clean-gen

# 一键：生成 go-zero rpc 服务骨架 + 修复 import + gofmt + go mod tidy
gen-rpc:
	@echo "🔧 [services/system] goctl rpc protoc..."
	$(GOCTL) rpc protoc $(PROTO) \
		--proto_path=$(REPO_ROOT) \
		--module=$(MODULE) \
		--go_out=. \
		--go-grpc_out=. \
		--zrpc_out=. \
		--style=gozero
	@$(MAKE) fix-imports
	@$(MAKE) fmt
	@$(MAKE) tidy
	@echo "✅ [services/system] gen-rpc done"

# 修复 goctl 可能生成的错误 import 路径/别名
fix-imports:
	@echo "🔧 [services/system] fixing imports..."
	@# 1) 把任何形如 ".../wklive/proto/system" 的错误前缀统一改成 "wklive/proto/system"
	@# 常见错误：wklive/services/system/wklive/proto/system
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's#"$(MODULE)/wklive/proto/system"#"wklive/proto/system"#g'
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's#"wklive/services/system/wklive/proto/system"#"wklive/proto/system"#g'
	@# 2) 修复 goctl 生成的包别名 system_system.xxx -> system.xxx（如果有）
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
		's/system_system\./system\./g'
	@echo "✅ [services/system] imports fixed"

fmt:
	@echo "🔧 [services/system] gofmt..."
	@gofmt -w $$(find . -name "*.go" -type f)

tidy:
	@echo "🔧 [services/system] go mod tidy..."
	@go mod tidy

# 谨慎：只清理 goctl 生成的 server 目录（避免误删你写的 logic）
clean-gen:
	@rm -rf wklive/

.PHONY: gen-model
gen-model:
	goctl model mysql ddl --src system.sql --dir ./models
