.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/system] generating system pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/system/config.proto \
		proto/system/system.proto
	@echo "✅ [proto/system] generated"

clean:
	rm -f *.pb.go