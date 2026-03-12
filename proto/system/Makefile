.PHONY: gen clean

gen:
	@echo "🔧 [rpc/system] generating system pb..."
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		system.proto
	@echo "✅ [rpc/system] generated: system.pb.go / system_grpc.pb.go"

clean:
	rm -f *.pb.go
