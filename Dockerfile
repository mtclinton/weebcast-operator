# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /workspace

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/
COPY pkg/ pkg/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager cmd/main.go

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /workspace/manager .

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65532:65532

ENTRYPOINT ["/manager"]

