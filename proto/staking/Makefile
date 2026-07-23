.PHONY: gen clean

REPO_ROOT := ../..

gen:
	@echo "🔧 [proto/staking] generating staking pb..."
	@rm -f staking_*.pb.go
	cd $(REPO_ROOT) && protoc \
		-I . \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/common/common.proto \
		proto/staking/*.proto
	@echo "✅ [proto/staking] generated"

clean:
	rm -f *.pb.go
