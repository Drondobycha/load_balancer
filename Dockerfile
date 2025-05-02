# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o balancer ./cmd

# Run stage
FROM scratch
WORKDIR /app
COPY --from=builder /app/balancer .
COPY config/local.yaml ./configs/config.yaml
EXPOSE 8080
ENTRYPOINT ["./balancer"]
CMD ["--config=configs/config.yaml"]