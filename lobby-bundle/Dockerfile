# 階段一：打包前端
FROM node:20-alpine AS client-builder
WORKDIR /app

# Vite 環境變數（編譯時嵌入）
ARG VITE_AUTH_PROXY_URL
ARG VITE_GAME_SERVER_URL
ENV VITE_AUTH_PROXY_URL=${VITE_AUTH_PROXY_URL}
ENV VITE_GAME_SERVER_URL=${VITE_GAME_SERVER_URL}

COPY shared/ /shared/
COPY lobby-client/package*.json ./
RUN npm ci
COPY lobby-client/ .
RUN npm run build

# 階段二：最終映像（使用預編譯的二進位）
FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates

# 複製預編譯的 Go 二進位（由 make docker-compile 產生）
COPY lobby-bundle/mahjong-lobby-server .
COPY --from=client-builder /dist/lobby ./public

ENV PORT=3001
ENV GIN_MODE=release
ENV STATIC_DIR=./public
EXPOSE 3001
CMD ["./mahjong-lobby-server"]
