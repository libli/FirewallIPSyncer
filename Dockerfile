# builder
FROM golang:1.20-bullseye as builder

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download -x

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -x -o FirewallIPSyncer .

# runner
FROM debian:bullseye-slim
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /build/FirewallIPSyncer /app/
CMD ["/app/FirewallIPSyncer"]