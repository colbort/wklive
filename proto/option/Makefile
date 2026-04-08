.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/option] generating option pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/option/enum.proto \
		proto/option/model.proto \
		proto/option/option_app.proto \
		proto/option/option_admin.proto
	@echo "✅ [proto/option] generated"

clean:
	rm -f *.pb.go