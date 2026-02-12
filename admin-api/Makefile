API_DIR := api
TMP_DIR := .gen-tmp

MODULES := auth_public auth_private system user payment
COMMON  := $(API_DIR)/common.api

GOCTL ?= goctl
MODULE_NAME := $(shell awk '/^module /{print $$2}' go.mod)

.PHONY: gen check clean clean-gen gen-all merge-all gen-modules merge-modules fix-imports verify

# 一键生成
gen: check clean-gen gen-all merge-all gen-modules merge-modules fix-imports verify clean
	@echo "✅ generate done"

# 检查：子模块 api 不能包含 syntax/info（只能在 common.api）
check:
	@for m in $(MODULES); do \
		if grep -nE '^\s*syntax\s*=' $(API_DIR)/$$m.api >/dev/null 2>&1; then \
			echo "❌ $(API_DIR)/$$m.api contains 'syntax = ...' (must be only in common.api)"; exit 1; \
		fi; \
		if grep -nE '^\s*info\s*\(' $(API_DIR)/$$m.api >/dev/null 2>&1; then \
			echo "❌ $(API_DIR)/$$m.api contains 'info(...)' (must be only in common.api)"; exit 1; \
		fi; \
	done
	@echo "✅ api files check ok"

clean-gen:
	@rm -rf $(TMP_DIR)
	@mkdir -p $(TMP_DIR)

# 生成全量 all.api：用于更新 routes/types/config/svc/admin.go（这些覆盖是安全的）
gen-all:
	@{ \
		set -e; \
		out="$(TMP_DIR)/all.api"; \
		: > "$$out"; \
		cat "$(COMMON)" >> "$$out"; \
		for m in $(MODULES); do \
			f="$(API_DIR)/$$m.api"; \
			test -f "$$f" || { echo "❌ missing api file: $$f"; exit 1; }; \
			cat "$$f" >> "$$out"; \
		done; \
	}
	@$(GOCTL) api go -api $(TMP_DIR)/all.api -dir $(TMP_DIR)/all


# 合并全量产物（覆盖更新：routes/types/config/svc/admin.go）
merge-all:
	@test -f $(TMP_DIR)/all/admin.go || (echo "❌ gen-all failed: admin.go not found"; exit 1)
	@mkdir -p internal/handler internal/types internal/config internal/svc etc
	@cp -f $(TMP_DIR)/all/internal/handler/routes.go internal/handler/routes.go
	@cp -f $(TMP_DIR)/all/internal/types/types.go internal/types/types.go
	@cp -n $(TMP_DIR)/all/internal/config/config.go internal/config/config.go 2>/dev/null || true
	@rsync -a --ignore-existing $(TMP_DIR)/all/internal/svc/ internal/svc/
	@cp -n $(TMP_DIR)/all/admin.go admin.go 2>/dev/null || true
	@cp -n $(TMP_DIR)/all/etc/*.yaml etc/ 2>/dev/null || true
	@cp -n $(TMP_DIR)/all/go.mod go.mod 2>/dev/null || true
	@cp -n $(TMP_DIR)/all/go.sum go.sum 2>/dev/null || true

# 每个模块单独生成到 tmp：用于“只合并新增 handler/logic（不覆盖你写过的）”
gen-modules:
	@for m in $(MODULES); do \
		mkdir -p $(TMP_DIR)/$$m; \
		cat $(COMMON) $(API_DIR)/$$m.api > $(TMP_DIR)/$$m/$$m.api; \
		$(GOCTL) api go -api $(TMP_DIR)/$$m/$$m.api -dir $(TMP_DIR)/$$m/out; \
	done

# handler/logic 只拷贝不存在的文件（ignore-existing），避免覆盖你写过的代码
merge-modules:
	@for m in $(MODULES); do \
		mkdir -p internal/handler/$$m internal/logic/$$m; \
		rsync -a --ignore-existing $(TMP_DIR)/$$m/out/internal/handler/$$m/ internal/handler/$$m/ 2>/dev/null || true; \
		rsync -a --ignore-existing $(TMP_DIR)/$$m/out/internal/logic/$$m/   internal/logic/$$m/   2>/dev/null || true; \
	done

# 自动修复所有 go 文件里的临时 import（包含 admin.go / internal/svc / internal/logic 等）
fix-imports:
	@echo "🔧 Fixing .gen-tmp imports to $(MODULE_NAME)/internal/..."
	@# ① 修复 .gen-tmp/all/internal/xxx -> internal/xxx
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
	  's#$(MODULE_NAME)/\.gen-tmp/all/internal/#$(MODULE_NAME)/internal/#g'
	@# ② 修复 .gen-tmp/<mod>/out/internal/xxx -> internal/xxx
	@find . -name "*.go" -type f -print0 | xargs -0 sed -i \
	  's#$(MODULE_NAME)/\.gen-tmp/[^"]*/out/internal/#$(MODULE_NAME)/internal/#g'
	@echo "✅ imports fixed"

# 验证：工程里不允许再出现 .gen-tmp import
verify:
	@echo "🔎 Verifying no .gen-tmp imports in *.go ..."
	@# 允许 .gen-tmp 目录存在，但不允许被 import
	@if grep -R --line-number --include="*.go" "\.gen-tmp" . >/dev/null 2>&1; then \
		echo "❌ Found .gen-tmp references in Go files:"; \
		grep -R --line-number --include="*.go" "\.gen-tmp" .; \
		exit 1; \
	else \
		echo "✅ no .gen-tmp imports"; \
	fi

# 清理临时目录
clean:
	@rm -rf $(TMP_DIR)
