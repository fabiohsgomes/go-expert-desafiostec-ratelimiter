# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main ./examples/main.go

# Final stage
FROM scratch

# Copy the binary from builder stage
COPY --from=builder /app/main /main

# Expose port
EXPOSE 8080

# Run the application
CMD ["/main"]