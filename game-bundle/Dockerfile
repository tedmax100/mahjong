# 階段一：打包前端
FROM node:20-alpine AS client-builder
WORKDIR /app
COPY shared/ /shared/
COPY game-client/package*.json ./
RUN npm ci
COPY game-client/ .
RUN npm run build

# 階段二：最終映像（使用預編譯的二進位）
FROM alpine:3.19
WORKDIR /app
RUN apk --no-cache add ca-certificates

# 複製預編譯的 Go 二進位（由 make docker-compile 產生）
COPY game-bundle/mahjong-game-server .
COPY --from=client-builder /dist/game ./public

ENV STANDALONE_MODE=true
ENV STATIC_DIR=./public
ENV PORT=8080
EXPOSE 8080
CMD ["./mahjong-game-server"]
