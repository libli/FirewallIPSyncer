# builder
FROM golang:1.20-bullseye as builder

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download -x

COPY . .

# 编译server
WORKDIR /build/cmd/server
RUN go build -ldflags="-w -s" -x -o /build/server

# 编译client
WORKDIR /build/cmd/client
RUN go build -ldflags="-w -s" -x -o /build/client

# runner
FROM debian:bullseye-slim
ENV TZ=Asia/Shanghai
# 设置环境变量
ENV TYPE="server"
WORKDIR /app
COPY --from=builder /build/server /app/server
COPY --from=builder /build/client /app/client
# 设置启动命令
CMD ["/bin/sh", "-c", "if [ \"$TYPE\" = \"client\" ]; then /app/client; else /app/server; fi"]