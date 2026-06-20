.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/chat] generating chat pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/chat/enum.proto \
		proto/chat/model.proto \
		proto/chat/chat_app.proto \
		proto/chat/chat_admin.proto \
		proto/chat/chat_internal.proto 
	@echo "✅ [proto/chat] generated"

clean:
	rm -f *.pb.go