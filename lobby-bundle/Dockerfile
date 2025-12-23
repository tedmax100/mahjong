# 階段一：打包前端
FROM node:20-alpine AS client-builder
WORKDIR /app

# Vite 環境變數（編譯時嵌入）
ARG VITE_AUTH_PROXY_URL
ENV VITE_AUTH_PROXY_URL=${VITE_AUTH_PROXY_URL}

COPY shared/ /shared/
COPY lobby-client/package*.json ./
RUN npm ci
COPY lobby-client/ .
RUN npm run build && mkdir -p /app/dist && cp -r dist/* /app/dist/

# 階段二：編譯 Go 後端
FROM golang:1.21-alpine AS server-builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN go build -o mahjong-lobby-server ./cmd/lobby/main.go

# 階段三：最終映像
FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=server-builder /app/mahjong-lobby-server .
COPY --from=client-builder /app/dist ./public

ENV PORT=3001
ENV GIN_MODE=release
ENV STATIC_DIR=./public
EXPOSE 3001
CMD ["./mahjong-lobby-server"]
