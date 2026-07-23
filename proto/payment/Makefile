.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/payment] generating payment pb..."
	@rm -f payment_*.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/payment/*.proto
	@echo "✅ [proto/payment] generated"

clean:
	rm -f *.pb.go
