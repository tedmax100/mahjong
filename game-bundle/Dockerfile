# 階段一：編譯 Go 伺服器
FROM golang:1.21-alpine AS server-builder

WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mahjong-game-server ./cmd/main.go

# 階段二：打包前端
FROM node:20-alpine AS client-builder

WORKDIR /app

# 複製 shared 模組
COPY shared/ /shared/

# 複製 game-client
COPY game-client/package*.json ./
RUN npm ci
COPY game-client/ .
RUN npm run build

# 階段三：最終映像
FROM alpine:3.19

WORKDIR /app

# 安裝 ca-certificates（HTTPS 需要）
RUN apk --no-cache add ca-certificates

# 複製伺服器二進位
COPY --from=server-builder /mahjong-game-server .

# 複製前端靜態檔案
COPY --from=client-builder /app/dist ./public

# 環境變數
ENV STANDALONE_MODE=true
ENV STATIC_DIR=./public
ENV PORT=8080

EXPOSE 8080

CMD ["./mahjong-game-server"]
