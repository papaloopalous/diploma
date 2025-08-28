FROM golang:1.24-alpine AS builder
RUN apk add --no-cache gcc musl-dev libffi-dev openssl-dev pkgconfig git
WORKDIR /go/src/tarantool_api
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o tarantool_api ./cmd/

FROM alpine:latest AS runtime
WORKDIR /app
COPY --from=builder /go/src/tarantool_api/tarantool_api /app/
RUN apk add --no-cache openssl
RUN chmod +x ./tarantool_api
EXPOSE 8080/tcp
ENTRYPOINT ./tarantool_api