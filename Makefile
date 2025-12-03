.PHONY: help install-cloudflared start stop tunnel dev dev-tunnel clean

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
	@echo "  $(GREEN)make start$(NC)          - 啟動所有服務 + Cloudflare Tunnel"
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

start: install-cloudflared ## 啟動所有服務 + Cloudflare Tunnel（推薦）
	@echo "$(BLUE)======================================$(NC)"
	@echo "$(BLUE)  啟動麻將遊戲開發環境$(NC)"
	@echo "$(BLUE)======================================$(NC)"
	@echo ""
	@./start-tunnel.sh

stop: ## 停止所有服務
	@echo "$(YELLOW)正在停止所有服務...$(NC)"
	@pkill -f "go run cmd/main.go" 2>/dev/null || true
	@pkill -f "vite" 2>/dev/null || true
	@pkill -f "cloudflared tunnel" 2>/dev/null || true
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
	@cd server && go run cmd/main.go

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
build: ## 構建項目
	@echo "$(YELLOW)構建前端...$(NC)"
	@cd client && npm run build
	@echo "$(YELLOW)構建後端...$(NC)"
	@cd server && go build -o ../mahjong-server cmd/main.go
	@echo "$(GREEN)✓ 構建完成$(NC)"

# 安裝依賴
install: ## 安裝所有依賴
	@echo "$(YELLOW)安裝前端依賴...$(NC)"
	@cd client && npm install
	@echo "$(YELLOW)安裝後端依賴...$(NC)"
	@cd server && go mod download
	@echo "$(GREEN)✓ 依賴安裝完成$(NC)"
