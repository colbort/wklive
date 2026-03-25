.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/user] generating user pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/user/common.proto \
		proto/user/model.proto \
		proto/user/user_app.proto \
		proto/user/user_admin.proto 
	@echo "✅ [proto/user] generated"

clean:
	rm -f *.pb.go