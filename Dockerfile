FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
COPY . .
RUN go build -o /bin/kv-demo ./cmd/kv-demo

FROM gcr.io/distroless/base-debian12
COPY --from=builder /bin/kv-demo /bin/kv-demo
ENV PORT=8080
EXPOSE 8080
CMD ["/bin/kv-demo"]
