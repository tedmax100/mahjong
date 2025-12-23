# 階段一：打包前端
FROM node:20-alpine AS client-builder

WORKDIR /app

# 複製 shared 模組
COPY shared/ /shared/

# 複製 game-client
COPY game-client/package*.json ./
RUN npm ci
COPY game-client/ .
RUN npm run build

# 階段二：最終映像
FROM --platform=linux/amd64 alpine:3.19

WORKDIR /app

# 安裝 ca-certificates（HTTPS 需要）
RUN apk --no-cache add ca-certificates

# 複製預編譯的伺服器二進位（本地交叉編譯）
COPY game-bundle/mahjong-game-server .

# 複製前端靜態檔案
COPY --from=client-builder /dist/game ./public

# 環境變數
ENV STANDALONE_MODE=true
ENV STATIC_DIR=./public
ENV PORT=8080

EXPOSE 8080

CMD ["./mahjong-game-server"]
