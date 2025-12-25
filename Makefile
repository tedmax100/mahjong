.PHONY: help install-cloudflared start stop tunnel tunnel-all dev dev-tunnel clean start-lobby dev-with-auth \
        docker-compile docker-buildx-setup docker-push-game docker-push-lobby docker-push-all docker-clean-binaries \
        version version-bump-game version-bump-lobby version-bump-all \
        docker-push-game-versioned docker-push-lobby-versioned docker-push-all-versioned \
        release-game release-lobby release-all deploy-game deploy-lobby deploy-all

# 顏色定義
GREEN  := \033[0;32m
YELLOW := \033[1;33m
BLUE   := \033[0;34m
RED    := \033[0;31m
NC     := \033[0m # No Color

# 配置
BACKEND_PORT := 8080
FRONTEND_PORT := 5173
TUNNEL_CONFIG := cloudflare-tunnel.yml

help: ## 顯示幫助信息
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  麻將遊戲 - 可用命令$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@echo ""
	@echo "$(YELLOW)快速開始:$(NC)"
	@echo "  $(GREEN)make tunnel-all$(NC)     - 啟動分離式前端 + 雙 Tunnel（推薦）"
	@echo "  $(GREEN)make start$(NC)          - 啟動舊版單一前端 + Tunnel"
	@echo "  $(GREEN)make stop$(NC)           - 停止所有服務"
	@echo ""
	@echo "$(YELLOW)所有命令:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

install-cloudflared: ## 安裝 cloudflared（如果尚未安裝）
	@echo "$(YELLOW)檢查 cloudflared 是否已安裝...$(NC)"
	@if ! command -v cloudflared &> /dev/null; then \
		echo "$(YELLOW)正在安裝 cloudflared...$(NC)"; \
		if [ "$$(uname)" = "Linux" ]; then \
			wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb; \
			sudo dpkg -i cloudflared-linux-amd64.deb; \
			rm cloudflared-linux-amd64.deb; \
		elif [ "$$(uname)" = "Darwin" ]; then \
			brew install cloudflared; \
		fi; \
		echo "$(GREEN)✓ cloudflared 安裝完成$(NC)"; \
	else \
		echo "$(GREEN)✓ cloudflared 已安裝$(NC)"; \
	fi

start: install-cloudflared ## 啟動舊版單一前端 + Cloudflare Tunnel
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  啟動麻將遊戲開發環境$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@echo ""
	@./start-tunnel.sh

tunnel-all: install-cloudflared ## 啟動分離式前端 + 雙 Tunnel（大廳+遊戲，推薦）
	@./start-tunnel-all.sh

stop: ## 停止所有服務
	@echo "$(YELLOW)正在停止所有服務...$(NC)"
	@pkill -f "go run cmd/main.go" 2>/dev/null || true
	@pkill -f "go run cmd/lobby/main.go" 2>/dev/null || true
	@pkill -f "vite" 2>/dev/null || true
	@pkill -f "cloudflared tunnel" 2>/dev/null || true
	@pkill -f "cloudflared" 2>/dev/null || true
	@pkill -f "npm run dev" 2>/dev/null || true
	@sleep 1
	@echo "$(GREEN)✓ 所有服務已停止$(NC)"

dev: ## 啟動本地開發服務器（不使用 tunnel）
	@echo "$(YELLOW)啟動本地開發環境...$(NC)"
	@$(MAKE) -j2 start-backend start-frontend

dev-tunnel: install-cloudflared ## 啟動開發服務器並通過 Cloudflare Tunnel 暴露到互聯網
	@echo "$(YELLOW)啟動開發環境並建立 Cloudflare Tunnel...$(NC)"
	@$(MAKE) create-tunnel-config
	@$(MAKE) -j3 start-backend start-frontend start-tunnel

start-backend: ## 啟動後端服務器
	@echo "$(GREEN)🚀 啟動後端服務器 (Port $(BACKEND_PORT))...$(NC)"
	@cd server && PORT=$(BACKEND_PORT) LOBBY_INTERNAL_SECRET=dev-internal-secret go run cmd/main.go

start-frontend: ## 啟動前端開發服務器
	@echo "$(GREEN)🚀 啟動前端開發服務器 (Port $(FRONTEND_PORT))...$(NC)"
	@cd client && npm run dev

start-tunnel: ## 啟動 Cloudflare Tunnel
	@echo "$(GREEN)🌐 啟動 Cloudflare Tunnel...$(NC)"
	@sleep 5  # 等待服務器啟動
	@cloudflared tunnel --config $(TUNNEL_CONFIG)

create-tunnel-config: ## 創建 Cloudflare Tunnel 配置文件
	@if [ ! -f $(TUNNEL_CONFIG) ]; then \
		echo "$(YELLOW)創建 Cloudflare Tunnel 配置...$(NC)"; \
		cat > $(TUNNEL_CONFIG) <<-EOF; \
		url: http://localhost:$(FRONTEND_PORT); \
		ingress:; \
		  - hostname: "*"; \
		    service: http://localhost:$(FRONTEND_PORT); \
		  - service: http_status:404; \
		EOF \
		echo "$(GREEN)✓ 配置文件已創建: $(TUNNEL_CONFIG)$(NC)"; \
	fi

tunnel-quick: install-cloudflared ## 快速啟動 tunnel（不使用配置文件）
	@echo "$(YELLOW)啟動開發環境並建立快速 Tunnel...$(NC)"
	@echo "$(BLUE)提示: 按 Ctrl+C 停止所有服務$(NC)"
	@trap 'kill 0' SIGINT; \
	(cd server && go run cmd/main.go) & \
	(cd client && npm run dev) & \
	sleep 5 && \
	cloudflared tunnel --url http://localhost:$(FRONTEND_PORT) & \
	wait

tunnel-backend: install-cloudflared ## 只為後端服務器建立 tunnel
	@echo "$(GREEN)🌐 為後端服務器建立 Tunnel...$(NC)"
	@cloudflared tunnel --url http://localhost:$(BACKEND_PORT)

clean: stop ## 清理臨時文件和停止所有服務
	@echo "$(YELLOW)清理臨時文件...$(NC)"
	@rm -f $(TUNNEL_CONFIG)
	@rm -f cloudflared*.deb 2>/dev/null || true
	@echo "$(GREEN)✓ 清理完成$(NC)"

logs: ## 查看服務器日誌
	@tail -f server.log

status: ## 檢查服務狀態
	@echo "$(BLUE)檢查服務狀態...$(NC)"
	@echo ""
	@echo "$(YELLOW)後端服務器 (Port $(BACKEND_PORT)):$(NC)"
	@if curl -s http://localhost:$(BACKEND_PORT) > /dev/null 2>&1; then \
		echo "  $(GREEN)✓ 運行中$(NC)"; \
	else \
		echo "  $(YELLOW)✗ 未運行$(NC)"; \
	fi
	@echo ""
	@echo "$(YELLOW)前端開發服務器 (Port $(FRONTEND_PORT)):$(NC)"
	@if curl -s http://localhost:$(FRONTEND_PORT) > /dev/null 2>&1; then \
		echo "  $(GREEN)✓ 運行中$(NC)"; \
	else \
		echo "  $(YELLOW)✗ 未運行$(NC)"; \
	fi
	@echo ""
	@echo "$(YELLOW)Cloudflared Tunnel:$(NC)"
	@if pgrep -f "cloudflared" > /dev/null; then \
		echo "  $(GREEN)✓ 運行中$(NC)"; \
	else \
		echo "  $(YELLOW)✗ 未運行$(NC)"; \
	fi

# 測試相關命令
test: ## 運行測試
	@cd client && npm run test:run

test-ui: ## 運行測試 UI
	@cd client && npm run test:ui

# 構建相關命令
# 本地構建用的版本資訊
LOCAL_GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LOCAL_BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LOCAL_LDFLAGS := -X mahjong/internal/version.GitCommit=$(LOCAL_GIT_COMMIT) \
                 -X mahjong/internal/version.Version=dev \
                 -X mahjong/internal/version.BuildTime=$(LOCAL_BUILD_TIME)

build: ## 構建項目
	@echo "$(YELLOW)構建前端...$(NC)"
	@cd client && npm run build
	@echo "$(YELLOW)構建遊戲伺服器...$(NC)"
	@cd server && go build -ldflags "$(LOCAL_LDFLAGS)" -o ../mahjong-server cmd/main.go
	@echo "$(YELLOW)構建 Lobby 伺服器...$(NC)"
	@cd server && go build -ldflags "$(LOCAL_LDFLAGS)" -o ../mahjong-lobby cmd/lobby/main.go
	@echo "$(GREEN)✓ 構建完成$(NC)"

# 安裝依賴
install: ## 安裝所有依賴
	@echo "$(YELLOW)安裝前端依賴...$(NC)"
	@cd client && npm install
	@echo "$(YELLOW)安裝後端依賴...$(NC)"
	@cd server && go mod download
	@echo "$(GREEN)✓ 依賴安裝完成$(NC)"

# Lobby & Auth Proxy 相關命令
start-lobby: ## 啟動 Lobby & Auth Proxy 伺服器 (Port 3001)
	@echo "$(GREEN)🏠 啟動 Lobby & Auth Proxy (Port 3001)...$(NC)"
	@cd server && go run cmd/lobby/main.go

dev-with-auth: ## 啟動本地開發環境（含 Lobby & Auth）
	@echo "$(YELLOW)啟動本地開發環境（含認證）...$(NC)"
	@$(MAKE) -j3 start-backend start-frontend start-lobby

# ============================================
# 分離式前端開發命令
# ============================================

dev-lobby: ## 啟動大廳前端開發 (Port 5174)
	@echo "$(GREEN)🏠 啟動大廳前端 (Port 5174)...$(NC)"
	@cd lobby-client && npm run dev

dev-game: ## 啟動遊戲前端開發 (Port 5175)
	@echo "$(GREEN)🎮 啟動遊戲前端 (Port 5175)...$(NC)"
	@cd game-client && npm run dev

dev-all: ## 啟動所有服務（分離模式）
	@echo "$(YELLOW)啟動分離模式開發環境...$(NC)"
	@$(MAKE) -j4 start-backend start-lobby dev-lobby dev-game

# ============================================
# 分離式前端構建命令
# ============================================

build-lobby-client: ## 構建大廳前端
	@echo "$(YELLOW)構建大廳前端...$(NC)"
	@cd lobby-client && npm run build
	@echo "$(GREEN)✓ 大廳前端構建完成 → dist/lobby/$(NC)"

build-game-client: ## 構建遊戲前端
	@echo "$(YELLOW)構建遊戲前端...$(NC)"
	@cd game-client && npm run build
	@echo "$(GREEN)✓ 遊戲前端構建完成 → dist/game/$(NC)"

build-game-bundle: build-game-client ## 構建遊戲獨立部署包
	@echo "$(YELLOW)構建遊戲獨立部署包...$(NC)"
	@mkdir -p game-bundle/dist
	@cd server && go build -o ../game-bundle/dist/mahjong-game-server cmd/main.go
	@cp -r dist/game/* game-bundle/dist/ 2>/dev/null || true
	@echo "$(GREEN)✓ 遊戲獨立部署包已準備: game-bundle/dist/$(NC)"

# ============================================
# Docker 相關命令
# ============================================

# DockerHub 配置
DOCKERHUB_USER := tedmax100
GAME_IMAGE := $(DOCKERHUB_USER)/mahjong-game
LOBBY_IMAGE := $(DOCKERHUB_USER)/mahjong-lobby

docker-build-game: ## 構建遊戲 Docker 映像（本地）
	@echo "$(YELLOW)構建遊戲 Docker 映像...$(NC)"
	@docker build -f game-bundle/Dockerfile -t mahjong-game:latest .
	@echo "$(GREEN)✓ Docker 映像構建完成: mahjong-game:latest$(NC)"

docker-run-game: ## 運行遊戲 Docker 容器
	@echo "$(YELLOW)運行遊戲 Docker 容器...$(NC)"
	@docker run -d -p 8080:8080 --name mahjong-game mahjong-game:latest
	@echo "$(GREEN)✓ 容器已啟動: http://localhost:8080$(NC)"

docker-stop-game: ## 停止遊戲 Docker 容器
	@docker stop mahjong-game 2>/dev/null || true
	@docker rm mahjong-game 2>/dev/null || true
	@echo "$(GREEN)✓ 容器已停止$(NC)"

# ============================================
# DockerHub 構建與推送命令
# ============================================

# 從 versions.json 讀取版本（用於 Docker 構建）
VERSIONS_FILE := versions.json
GAME_VERSION := $(shell cat $(VERSIONS_FILE) | grep '"game"' | sed 's/.*"game": *"\([^"]*\)".*/\1/')
LOBBY_VERSION := $(shell cat $(VERSIONS_FILE) | grep '"lobby"' | sed 's/.*"lobby": *"\([^"]*\)".*/\1/')

# 構建時注入的版本資訊
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS_GAME := -X mahjong/internal/version.GitCommit=$(GIT_COMMIT) \
                -X mahjong/internal/version.Version=$(GAME_VERSION) \
                -X mahjong/internal/version.BuildTime=$(BUILD_TIME)
LDFLAGS_LOBBY := -X mahjong/internal/version.GitCommit=$(GIT_COMMIT) \
                 -X mahjong/internal/version.Version=$(LOBBY_VERSION) \
                 -X mahjong/internal/version.BuildTime=$(BUILD_TIME)

docker-compile: ## 交叉編譯 Go 二進位（linux/amd64，含版本資訊）
	@echo "$(YELLOW)交叉編譯 Game Server...$(NC)"
	@echo "  Commit: $(GIT_COMMIT) | Version: $(GAME_VERSION)"
	@cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_GAME)" -o ../game-bundle/mahjong-game-server ./cmd/main.go
	@echo "$(GREEN)✓ Game Server 編譯完成$(NC)"
	@echo "$(YELLOW)交叉編譯 Lobby Server...$(NC)"
	@echo "  Commit: $(GIT_COMMIT) | Version: $(LOBBY_VERSION)"
	@cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS_LOBBY)" -o ../lobby-bundle/mahjong-lobby-server ./cmd/lobby/main.go
	@echo "$(GREEN)✓ Lobby Server 編譯完成$(NC)"

docker-buildx-setup: ## 設置 buildx 多架構構建器
	@echo "$(YELLOW)設置 buildx...$(NC)"
	@docker buildx create --name multiarch --driver docker-container --use 2>/dev/null || docker buildx use multiarch
	@echo "$(GREEN)✓ buildx 已就緒$(NC)"

docker-push-game: docker-compile docker-buildx-setup ## 構建並推送 Game 映像到 DockerHub
	@echo "$(YELLOW)構建並推送 Game 映像 (linux/amd64)...$(NC)"
	@docker buildx build --platform linux/amd64 -t $(GAME_IMAGE):latest -f game-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ 已推送: $(GAME_IMAGE):latest$(NC)"

docker-push-lobby: docker-compile docker-buildx-setup ## 構建並推送 Lobby 映像到 DockerHub
	@echo "$(YELLOW)構建並推送 Lobby 映像 (linux/amd64)...$(NC)"
	@docker buildx build --platform linux/amd64 -t $(LOBBY_IMAGE):latest -f lobby-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ 已推送: $(LOBBY_IMAGE):latest$(NC)"

docker-push-all: docker-compile docker-buildx-setup ## 構建並推送所有映像到 DockerHub
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  構建並推送所有 Docker 映像$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(YELLOW)構建並推送 Game 映像...$(NC)"
	@docker buildx build --platform linux/amd64 -t $(GAME_IMAGE):latest -f game-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ 已推送: $(GAME_IMAGE):latest$(NC)"
	@echo "$(YELLOW)構建並推送 Lobby 映像...$(NC)"
	@docker buildx build --platform linux/amd64 -t $(LOBBY_IMAGE):latest -f lobby-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ 已推送: $(LOBBY_IMAGE):latest$(NC)"
	@echo ""
	@echo "$(GREEN)======================================$(NC)"
	@echo "$(GREEN)  所有映像已推送完成！$(NC)"
	@echo "$(GREEN)======================================$(NC)"
	@echo "  $(BLUE)Game:$(NC)  $(GAME_IMAGE):latest"
	@echo "  $(BLUE)Lobby:$(NC) $(LOBBY_IMAGE):latest"

docker-clean-binaries: ## 清理本地編譯的二進位
	@rm -f game-bundle/mahjong-game-server lobby-bundle/mahjong-lobby-server
	@echo "$(GREEN)✓ 已清理本地二進位$(NC)"

# ============================================
# 安裝分離式前端依賴
# ============================================

install-lobby-client: ## 安裝大廳前端依賴
	@echo "$(YELLOW)安裝大廳前端依賴...$(NC)"
	@cd lobby-client && npm install
	@echo "$(GREEN)✓ 大廳前端依賴安裝完成$(NC)"

install-game-client: ## 安裝遊戲前端依賴
	@echo "$(YELLOW)安裝遊戲前端依賴...$(NC)"
	@cd game-client && npm install
	@echo "$(GREEN)✓ 遊戲前端依賴安裝完成$(NC)"

install-all: install install-lobby-client install-game-client ## 安裝所有依賴（含分離式前端）
	@echo "$(GREEN)✓ 所有依賴安裝完成$(NC)"

# ============================================
# 版本管理與發布命令
# ============================================

version: ## 顯示當前版本
	@echo "$(BLUE)當前版本:$(NC)"
	@echo "  $(YELLOW)Game:$(NC)  $(GAME_VERSION)"
	@echo "  $(YELLOW)Lobby:$(NC) $(LOBBY_VERSION)"

version-bump-game: ## 更新 Game 版本 (用法: make version-bump-game V=v0.0.9)
	@if [ -z "$(V)" ]; then \
		echo "$(RED)錯誤: 請指定版本號 V=vX.X.X$(NC)"; \
		echo "$(YELLOW)用法: make version-bump-game V=v0.0.9$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)更新 Game 版本: $(GAME_VERSION) → $(V)$(NC)"
	@sed -i 's/"game": *"[^"]*"/"game": "$(V)"/' $(VERSIONS_FILE)
	@echo "$(GREEN)✓ Game 版本已更新為 $(V)$(NC)"

version-bump-lobby: ## 更新 Lobby 版本 (用法: make version-bump-lobby V=v0.0.11)
	@if [ -z "$(V)" ]; then \
		echo "$(RED)錯誤: 請指定版本號 V=vX.X.X$(NC)"; \
		echo "$(YELLOW)用法: make version-bump-lobby V=v0.0.11$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)更新 Lobby 版本: $(LOBBY_VERSION) → $(V)$(NC)"
	@sed -i 's/"lobby": *"[^"]*"/"lobby": "$(V)"/' $(VERSIONS_FILE)
	@echo "$(GREEN)✓ Lobby 版本已更新為 $(V)$(NC)"

version-bump-all: ## 更新所有版本 (用法: make version-bump-all GAME_V=v0.0.9 LOBBY_V=v0.0.11)
	@if [ -z "$(GAME_V)" ] || [ -z "$(LOBBY_V)" ]; then \
		echo "$(RED)錯誤: 請指定版本號$(NC)"; \
		echo "$(YELLOW)用法: make version-bump-all GAME_V=v0.0.9 LOBBY_V=v0.0.11$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)更新版本:$(NC)"
	@echo "  Game:  $(GAME_VERSION) → $(GAME_V)"
	@echo "  Lobby: $(LOBBY_VERSION) → $(LOBBY_V)"
	@sed -i 's/"game": *"[^"]*"/"game": "$(GAME_V)"/' $(VERSIONS_FILE)
	@sed -i 's/"lobby": *"[^"]*"/"lobby": "$(LOBBY_V)"/' $(VERSIONS_FILE)
	@echo "$(GREEN)✓ 所有版本已更新$(NC)"

# ============================================
# Docker 版本化構建與推送
# ============================================

docker-push-game-versioned: docker-compile docker-buildx-setup ## 構建並推送 Game 映像（含版本標籤）
	@echo "$(YELLOW)構建並推送 Game 映像 $(GAME_VERSION)...$(NC)"
	@docker buildx build --platform linux/amd64 \
		-t $(GAME_IMAGE):$(GAME_VERSION) \
		-t $(GAME_IMAGE):latest \
		-f game-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ 已推送: $(GAME_IMAGE):$(GAME_VERSION)$(NC)"
	@echo "$(GREEN)✓ 已推送: $(GAME_IMAGE):latest$(NC)"

docker-push-lobby-versioned: docker-compile docker-buildx-setup ## 構建並推送 Lobby 映像（含版本標籤）
	@echo "$(YELLOW)構建並推送 Lobby 映像 $(LOBBY_VERSION)...$(NC)"
	@docker buildx build --platform linux/amd64 \
		-t $(LOBBY_IMAGE):$(LOBBY_VERSION) \
		-t $(LOBBY_IMAGE):latest \
		-f lobby-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ 已推送: $(LOBBY_IMAGE):$(LOBBY_VERSION)$(NC)"
	@echo "$(GREEN)✓ 已推送: $(LOBBY_IMAGE):latest$(NC)"

docker-push-all-versioned: docker-compile docker-buildx-setup ## 構建並推送所有映像（含版本標籤）
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  構建並推送所有 Docker 映像（含版本）$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(YELLOW)Game $(GAME_VERSION) | Lobby $(LOBBY_VERSION)$(NC)"
	@docker buildx build --platform linux/amd64 \
		-t $(GAME_IMAGE):$(GAME_VERSION) \
		-t $(GAME_IMAGE):latest \
		-f game-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ Game 映像已推送$(NC)"
	@docker buildx build --platform linux/amd64 \
		-t $(LOBBY_IMAGE):$(LOBBY_VERSION) \
		-t $(LOBBY_IMAGE):latest \
		-f lobby-bundle/Dockerfile --push .
	@echo "$(GREEN)✓ Lobby 映像已推送$(NC)"
	@echo ""
	@echo "$(GREEN)======================================$(NC)"
	@echo "$(GREEN)  所有映像已推送完成！$(NC)"
	@echo "$(GREEN)======================================$(NC)"
	@echo "  $(BLUE)Game:$(NC)  $(GAME_IMAGE):$(GAME_VERSION)"
	@echo "  $(BLUE)Lobby:$(NC) $(LOBBY_IMAGE):$(LOBBY_VERSION)"

# ============================================
# GitHub Release 命令
# ============================================

release-game: ## 創建 Game Release (用法: make release-game 或 make release-game V=v0.0.9)
	@VERSION=$${V:-$(GAME_VERSION)}; \
	TAG="game-$$VERSION"; \
	echo "$(YELLOW)創建 Game Release: $$TAG$(NC)"; \
	if [ -n "$(V)" ]; then \
		$(MAKE) version-bump-game V=$(V); \
		git add $(VERSIONS_FILE); \
		git commit -m "chore: bump game version to $(V)"; \
	fi; \
	git tag -a "$$TAG" -m "Game $$VERSION release"; \
	git push origin "$$TAG"; \
	gh release create "$$TAG" \
		--title "Game $$VERSION" \
		--notes "## Game Server Release $$VERSION" \
		--latest=false; \
	echo "$(GREEN)✓ Game Release 已創建: $$TAG$(NC)"

release-lobby: ## 創建 Lobby Release (用法: make release-lobby 或 make release-lobby V=v0.0.11)
	@VERSION=$${V:-$(LOBBY_VERSION)}; \
	TAG="lobby-$$VERSION"; \
	echo "$(YELLOW)創建 Lobby Release: $$TAG$(NC)"; \
	if [ -n "$(V)" ]; then \
		$(MAKE) version-bump-lobby V=$(V); \
		git add $(VERSIONS_FILE); \
		git commit -m "chore: bump lobby version to $(V)"; \
	fi; \
	git tag -a "$$TAG" -m "Lobby $$VERSION release"; \
	git push origin "$$TAG"; \
	gh release create "$$TAG" \
		--title "Lobby $$VERSION" \
		--notes "## Lobby Server Release $$VERSION" \
		--latest=false; \
	echo "$(GREEN)✓ Lobby Release 已創建: $$TAG$(NC)"

release-all: ## 創建兩個 Release (用法: make release-all 或 make release-all GAME_V=v0.0.9 LOBBY_V=v0.0.11)
	@if [ -n "$(GAME_V)" ] && [ -n "$(LOBBY_V)" ]; then \
		$(MAKE) version-bump-all GAME_V=$(GAME_V) LOBBY_V=$(LOBBY_V); \
		git add $(VERSIONS_FILE); \
		git commit -m "chore: bump versions - game $(GAME_V), lobby $(LOBBY_V)"; \
	fi
	@$(MAKE) release-game
	@$(MAKE) release-lobby

# ============================================
# 完整發布流程（版本 + Docker + Release）
# ============================================

deploy-game: ## 完整部署 Game (版本更新 + Docker 推送 + GitHub Release)
	@if [ -z "$(V)" ]; then \
		echo "$(RED)錯誤: 請指定版本號 V=vX.X.X$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  完整部署 Game $(V)$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@$(MAKE) version-bump-game V=$(V)
	@git add $(VERSIONS_FILE)
	@git commit -m "chore: bump game version to $(V)"
	@$(MAKE) docker-push-game-versioned
	@$(MAKE) release-game

deploy-lobby: ## 完整部署 Lobby (版本更新 + Docker 推送 + GitHub Release)
	@if [ -z "$(V)" ]; then \
		echo "$(RED)錯誤: 請指定版本號 V=vX.X.X$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  完整部署 Lobby $(V)$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@$(MAKE) version-bump-lobby V=$(V)
	@git add $(VERSIONS_FILE)
	@git commit -m "chore: bump lobby version to $(V)"
	@$(MAKE) docker-push-lobby-versioned
	@$(MAKE) release-lobby

deploy-all: ## 完整部署所有服務
	@if [ -z "$(GAME_V)" ] || [ -z "$(LOBBY_V)" ]; then \
		echo "$(RED)錯誤: 請指定版本號$(NC)"; \
		echo "$(YELLOW)用法: make deploy-all GAME_V=v0.0.9 LOBBY_V=v0.0.11$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  完整部署所有服務$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@$(MAKE) version-bump-all GAME_V=$(GAME_V) LOBBY_V=$(LOBBY_V)
	@git add $(VERSIONS_FILE)
	@git commit -m "chore: bump versions - game $(GAME_V), lobby $(LOBBY_V)"
	@$(MAKE) docker-push-all-versioned
	@$(MAKE) release-game
	@$(MAKE) release-lobby
	@echo "$(GREEN)======================================$(NC)"
	@echo "$(GREEN)  所有服務部署完成！$(NC)"
	@echo "$(GREEN)======================================$(NC)"