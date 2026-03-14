.PHONY: gen clean

gen:
	@echo "🔧 [proto/system] generating system pb..."
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		*.proto
	@echo "✅ [proto/system] generated: system.pb.go / system_grpc.pb.go"

clean:
	rm -f *.pb.go
