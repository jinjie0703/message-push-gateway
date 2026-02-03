# syntax=docker/dockerfile:1
# 指定使用 Dockerfile 的语法版本（BuildKit 语法），便于启用更现代的构建特性

FROM golang:1.24-alpine AS builder
# 第一阶段：使用 Go 官方镜像（Alpine 版）作为构建环境，并命名为 builder

WORKDIR /src
# 设置构建阶段的工作目录为 /src

RUN apk add --no-cache ca-certificates
# 安装 CA 证书：用于在 go mod download 等步骤中进行 HTTPS 访问（避免证书相关错误）

COPY go.mod go.sum ./
# 先只拷贝依赖描述文件，用于利用 Docker 分层缓存（依赖不变则无需重新下载）

RUN go mod download
# 下载 Go 依赖模块（会被缓存，提高后续构建速度）

COPY . .
# 再拷贝项目全部源码到镜像中

# Build a static binary (no CGO)
# 构建静态二进制：禁用 CGO，减少运行时依赖，更适合放到 distroless 镜像中运行
ENV CGO_ENABLED=0
# 禁用 CGO，强制生成纯 Go 的静态可执行文件（通常更易于跨环境运行）

RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server
# 编译服务端程序：
# -trimpath：去掉编译产物中的本地路径信息，提升可复现性/减少泄漏
# -ldflags "-s -w"：去符号表和调试信息，减小二进制体积
# -o /out/server：输出到 /out/server
# ./cmd/server：项目入口（常见 Go 多模块/多命令结构）

FROM gcr.io/distroless/static-debian12:nonroot
# 第二阶段：运行阶段镜像使用 distroless（极简、无包管理器、攻击面更小）
# static-debian12:nonroot：适合运行静态二进制，且默认非 root 用户

WORKDIR /app
# 设置运行阶段工作目录

COPY --from=builder /out/server /app/server
# 从 builder 阶段拷贝编译好的二进制到运行镜像中

EXPOSE 8080
# 声明容器对外提供的端口（文档/约定用途；实际映射由 docker run -p 决定）

ENV PORT=8080
# 设置环境变量 PORT，程序可读取该变量决定监听端口

USER nonroot:nonroot
# 指定以非 root 用户运行，提高安全性（避免容器内提权风险）

ENTRYPOINT ["/app/server"]
# 容器启动时执行的主进程：直接运行编译好的 server：项目入口（常见