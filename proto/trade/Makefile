.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/trade] generating trade pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/trade/enum.proto \
		proto/trade/model.proto \
		proto/trade/trade_app.proto \
		proto/trade/trade_admin.proto \
		proto/trade/trade_internal.proto 
	@echo "✅ [proto/trade] generated"

clean:
	rm -f *.pb.go