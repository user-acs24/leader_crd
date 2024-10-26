
# Use Golang image to build the binary
FROM golang:1.20 as builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /k8s-leader-election .

# Build the final image
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /k8s-leader-election /usr/local/bin/k8s-leader-election

# Set the entry point to run the binary
ENTRYPOINT ["/usr/local/bin/k8s-leader-election"]
