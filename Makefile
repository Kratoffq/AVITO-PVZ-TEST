.PHONY: generate
generate: generate-openapi generate-grpc

.PHONY: generate-openapi
generate-openapi:
	@echo "Generating OpenAPI code..."
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@PATH="$(shell go env GOPATH)/bin:$(PATH)" oapi-codegen -package openapi -generate types,server,spec api/openapi/swagger.yaml > api/openapi/types.gen.go
	@PATH="$(shell go env GOPATH)/bin:$(PATH)" oapi-codegen -package http -generate types,server,spec api/openapi/swagger.yaml > internal/handler/http/openapi.gen.go
	@echo "OpenAPI code generation completed"

.PHONY: generate-grpc
generate-grpc:
	@echo "Generating gRPC code..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/pvz.proto
	@echo "gRPC code generation completed"

.PHONY: test
test:
	@echo "Running all tests..."
	@GOFLAGS="-ldflags=-linkmode=internal" go test -v -timeout 5m ./test/... -race
	@echo "Tests completed"

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@GOFLAGS="-ldflags=-linkmode=internal" go test -v -timeout 5m ./test/unit/... -race
	@echo "Unit tests completed"

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@GOFLAGS="-ldflags=-linkmode=internal" go test -v -timeout 5m ./test/integration/... -race
	@echo "Integration tests completed"

.PHONY: test-coverage-detailed
test-coverage-detailed:
	@echo "Running detailed coverage analysis..."
	@go test -coverprofile=coverage.out ./test/...
	@echo "\nCoverage by package:"
	@go tool cover -func=coverage.out
	@echo "\nGenerating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated in coverage.html" 