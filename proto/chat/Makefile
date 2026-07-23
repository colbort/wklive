.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/chat] generating chat pb..."
	@rm -f chat_*.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/chat/*.proto
	@echo "✅ [proto/chat] generated"

clean:
	rm -f *.pb.go
