FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY shared ./shared
COPY services/realtime ./services/realtime
WORKDIR /app/services/realtime
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /realtime ./cmd/server

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /realtime /app/realtime
EXPOSE 8003
ENV PORT=8003
CMD ["/app/realtime"]


