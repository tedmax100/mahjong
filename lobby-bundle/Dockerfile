# 階段一：打包前端
FROM node:20-alpine AS client-builder

WORKDIR /app

# 複製 shared 模組
COPY shared/ /shared/

# 複製 lobby-client
COPY lobby-client/package*.json ./
RUN npm ci
COPY lobby-client/ .

# 構建前端（輸出到 /dist/lobby，然後移動到 /app/dist）
RUN npm run build && mkdir -p /app/dist && cp -r /dist/lobby/* /app/dist/

# 階段二：最終映像
FROM --platform=linux/amd64 alpine:3.19

WORKDIR /app

# 安裝 ca-certificates（HTTPS 需要）
RUN apk --no-cache add ca-certificates

# 複製預編譯的伺服器二進位（本地交叉編譯）
COPY lobby-bundle/mahjong-lobby-server .

# 複製前端靜態檔案
COPY --from=client-builder /app/dist ./public

# 環境變數（這些需要在 Zeabur 設定）
ENV PORT=3001
ENV GIN_MODE=release
ENV STATIC_DIR=./public

EXPOSE 3001

CMD ["./mahjong-lobby-server"]
