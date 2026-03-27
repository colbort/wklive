.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/itick] generating itick pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/itick/common.proto \
		proto/itick/model.proto \
		proto/itick/itick_app.proto \
		proto/itick/itick_admin.proto 
	@echo "✅ [proto/itick] generated"

clean:
	rm -f *.pb.go