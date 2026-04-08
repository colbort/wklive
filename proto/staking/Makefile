.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/staking] generating staking pb..."
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/staking/enum.proto \
		proto/staking/model.proto \
		proto/staking/staking_app.proto \
		proto/staking/staking_admin.proto
	@echo "✅ [proto/staking] generated"

clean:
	rm -f *.pb.go