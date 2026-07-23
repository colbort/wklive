.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/trade] generating trade pb..."
	@rm -f trade_*.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/trade/*.proto
	@echo "✅ [proto/trade] generated"

clean:
	rm -f *.pb.go
