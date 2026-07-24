.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/liquidity] generating liquidity pb..."
	@rm -f *.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/liquidity/*.proto
	@echo "✅ [proto/liquidity] generated"

clean:
	rm -f *.pb.go
