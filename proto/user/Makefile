.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/user] generating user pb..."
	@rm -f user_*.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/user/*.proto
	@echo "✅ [proto/user] generated"

clean:
	rm -f *.pb.go
