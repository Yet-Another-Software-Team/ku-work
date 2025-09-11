FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache build-base libwebp-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build a statically linked binary. This is important so the final
# image doesn't need the C libraries installed.
ENV CGO_ENABLED=1
RUN go build -ldflags '-extldflags "-static"' -o main .

# ---

FROM alpine:latest
WORKDIR /app

# Copy only the compiled binary from the "builder" stage
COPY --from=builder /app/main .

EXPOSE 8000

# Run the compiled app
CMD ["./main"]