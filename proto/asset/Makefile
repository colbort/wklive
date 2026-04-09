.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/asset] generating asset pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/asset/enum.proto \
		proto/asset/model.proto \
		proto/asset/asset_app.proto \
		proto/asset/asset_admin.proto \
		proto/asset/asset_internal.proto 
	@echo "✅ [proto/asset] generated"

clean:
	rm -f *.pb.go