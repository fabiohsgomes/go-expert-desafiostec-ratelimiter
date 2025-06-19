go-install:
	@echo "Installing Go tools"
	go install github.com/google/wire/cmd/wire@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/ktr0731/evans@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Go tools installed."
PHONY: go-install