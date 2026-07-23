.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/asset] generating asset pb..."
	@rm -f asset_*.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/asset/*.proto
	@echo "✅ [proto/asset] generated"

clean:
	rm -f *.pb.go
