FROM golang:1.22-alpine AS builder
WORKDIR /workspace

COPY go.work ./
COPY shared/ ./shared/
COPY services/driver/ ./services/driver/

WORKDIR /workspace/services/driver
RUN go build -o /app/driver ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/driver ./driver
EXPOSE 8080
CMD ["./driver"]
