.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/payment] generating payment pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/payment/enum.proto \
		proto/payment/model.proto \
		proto/payment/payment_app.proto \
		proto/payment/payment_admin.proto 
	@echo "✅ [proto/payment] generated"

clean:
	rm -f *.pb.go