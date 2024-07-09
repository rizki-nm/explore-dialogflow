FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o binary .

FROM alpine:3
RUN apk update && apk add --no-cache ca-certificates tzdata && update-ca-certificates
COPY --from=builder /app/binary /app/binary
CMD ["/app/binary"]