FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY ../../../../Downloads/go.mod go.sum ./
RUN go mod download
COPY ../../../../Downloads .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/api

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
