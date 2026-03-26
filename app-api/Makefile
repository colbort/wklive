API_FILE := api/core.api

GOCTL ?= goctl

.PHONY: gen check fmt clean verify

# 一键生成
gen: check
	@echo "🔧 generating from $(API_FILE) ..."
	@$(GOCTL) api go -api $(API_FILE) -dir .
	@echo "✅ generate done"

# 检查 core.api 是否存在
check:
	@test -f $(API_FILE) || (echo "❌ missing api file: $(API_FILE)"; exit 1)
	@echo "✅ api file check ok"

# 格式化
fmt:
	@echo "🔧 gofmt ..."
	@gofmt -w $$(find . -name "*.go" -type f)

# 可选验证
verify:
	@echo "🔎 verifying generated files..."
	@test -f app.go || (echo "❌ app.go not found"; exit 1)
	@test -f internal/handler/routes.go || (echo "❌ internal/handler/routes.go not found"; exit 1)
	@test -f internal/types/types.go || (echo "❌ internal/types/types.go not found"; exit 1)
	@echo "✅ verify ok"

# 清理 goctl 生成的常见文件（谨慎使用）
clean:
	@echo "⚠️ cleaning generated files..."
	@rm -f app.go
	@rm -rf internal/handler
	@rm -rf internal/logic
	@rm -rf internal/types
	@rm -rf internal/svc
	@rm -f internal/config/config.go
	@echo "✅ clean done"