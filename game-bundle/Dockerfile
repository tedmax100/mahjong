# 階段一：打包前端
FROM node:20-alpine AS client-builder
WORKDIR /app
COPY shared/ /shared/
COPY game-client/package*.json ./
RUN npm ci
COPY game-client/ .
RUN npm run build

# 階段二：編譯 Go 後端
FROM golang:1.21-alpine AS server-builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
# 注意：這裡編譯的是 cmd/main.go (Game Server 入口)
RUN go build -o mahjong-game-server ./cmd/main.go

# 階段三：最終映像
FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=server-builder /app/mahjong-game-server .
COPY --from=client-builder /app/dist/game ./public

ENV STANDALONE_MODE=true
ENV STATIC_DIR=./public
ENV PORT=8080
EXPOSE 8080
CMD ["./mahjong-game-server"]
