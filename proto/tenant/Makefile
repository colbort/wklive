.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/tenant] generating tenant pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/tenant/enum.proto \
		proto/tenant/config.proto \
		proto/tenant/tenant.proto
	@echo "✅ [proto/tenant] generated"

clean:
	rm -f *.pb.go