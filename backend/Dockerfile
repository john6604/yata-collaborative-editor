FROM golang:1.26.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -o ./relay ./cmd/relay

FROM alpine AS runtime

WORKDIR /app

COPY --from=builder /app/relay ./relay/
CMD ["./relay/relay"]

EXPOSE 8181