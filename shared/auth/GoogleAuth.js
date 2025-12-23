/**
 * Google OAuth 認證管理（Auth Proxy 版本）
 * 透過 Auth Proxy 進行 Google OAuth，支援動態網址
 *
 * 這是共用模組，被 lobby-client 和 game-client 共用
 */
export class GoogleAuth {
  constructor(options = {}) {
    this.user = null;
    this.onSignIn = null;
    this.onSignOut = null;
    // Auth Proxy URL（從選項或環境變數或預設值）
    // 生產環境預設使用當前網域（lobby-bundle 前後端同服務）
    const defaultAuthProxyUrl = typeof window !== 'undefined' && window.location.hostname !== 'localhost'
      ? window.location.origin
      : 'http://localhost:3001';
    this.authProxyUrl = options.authProxyUrl ||
      (typeof import.meta !== 'undefined' && import.meta.env?.VITE_AUTH_PROXY_URL) ||
      defaultAuthProxyUrl;
    // localStorage key
    this.storageKey = options.storageKey || 'game_auth_token';
  }

  /**
   * 初始化認證系統
   * 檢查 URL hash 或 localStorage 中的 token
   */
  init() {
    // 1. 先檢查 URL hash 是否有新 token（OAuth callback）
    this.checkUrlToken();

    // 2. 再檢查 localStorage 中的現有 token
    this.checkStoredToken();

    // 3. 設定登入按鈕事件（如果存在）
    const loginBtn = document.getElementById('google-login-btn');
    if (loginBtn) {
      loginBtn.addEventListener('click', () => this.signIn());
    }
  }

  /**
   * 檢查 URL hash 中的 token（OAuth callback 後）
   */
  checkUrlToken() {
    const hash = window.location.hash;

    // 檢查是否有 token
    if (hash.startsWith('#token=')) {
      const token = hash.substring(7); // 移除 '#token='

      // 清除 URL hash（避免分享連結時洩漏 token）
      window.history.replaceState(null, '', window.location.pathname + window.location.search);

      // 驗證並儲存 token
      const payload = this.parseJwt(token);

      if (payload && this.isTokenValid(payload)) {
        this.setUser(payload, token);
        localStorage.setItem(this.storageKey, token);
        console.log('從 URL 取得 token，登入成功:', this.user.name);
      } else {
        console.warn('URL 中的 token 無效或已過期');
      }

      return;
    }

    // 檢查是否有錯誤
    if (hash.startsWith('#error=')) {
      const error = hash.substring(7);
      window.history.replaceState(null, '', window.location.pathname + window.location.search);
      console.error('OAuth 錯誤:', decodeURIComponent(error));
    }
  }

  /**
   * 檢查 localStorage 中的 token
   */
  checkStoredToken() {
    if (this.user) return; // 已從 URL 取得

    const token = localStorage.getItem(this.storageKey);
    if (!token) return;

    const payload = this.parseJwt(token);

    if (payload && this.isTokenValid(payload)) {
      this.setUser(payload, token);
      console.log('從 localStorage 恢復登入:', this.user.name);
    } else {
      // Token 無效或過期，清除
      localStorage.removeItem(this.storageKey);
      console.log('已清除過期的 token');
    }
  }

  /**
   * 設定使用者資料並觸發回調
   */
  setUser(payload, token) {
    this.user = {
      id: payload.sub,
      name: payload.name,
      picture: payload.picture,
      token: token,
    };

    if (this.onSignIn) {
      this.onSignIn(this.user);
    }
  }

  /**
   * 從外部設定使用者（用於跨站跳轉）
   */
  setUserFromExternal(userData) {
    this.user = {
      id: userData.userId || userData.id,
      name: userData.userName || userData.name,
      picture: userData.picture || null,
      token: userData.token,
    };

    // 儲存 token
    if (userData.token) {
      localStorage.setItem(this.storageKey, userData.token);
    }

    if (this.onSignIn) {
      this.onSignIn(this.user);
    }
  }

  /**
   * 檢查 token 是否仍然有效
   */
  isTokenValid(payload) {
    if (!payload || !payload.exp) return false;

    // exp 是秒數，轉換為毫秒
    const expTime = payload.exp * 1000;
    const now = Date.now();

    // 預留 1 分鐘緩衝
    return expTime > (now + 60000);
  }

  /**
   * 發起登入（導向 Auth Proxy）
   */
  signIn() {
    const origin = window.location.origin;
    const loginUrl = `${this.authProxyUrl}/login?origin=${encodeURIComponent(origin)}`;

    console.log('導向 Auth Proxy 登入:', loginUrl);
    window.location.href = loginUrl;
  }

  /**
   * 登出
   */
  signOut() {
    localStorage.removeItem(this.storageKey);
    this.user = null;

    console.log('已登出');

    if (this.onSignOut) {
      this.onSignOut();
    }
  }

  /**
   * 處理 401 錯誤（自動登出）
   * 應在 API 呼叫收到 401 時調用
   */
  handleUnauthorized() {
    console.warn('收到 401 Unauthorized，自動登出');
    this.signOut();
    // 可選：重新載入頁面以重置狀態
    // window.location.reload();
  }

  /**
   * 解析 JWT token
   */
  parseJwt(token) {
    try {
      const parts = token.split('.');
      if (parts.length !== 3) return null;

      const base64Url = parts[1];
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      return JSON.parse(jsonPayload);
    } catch (error) {
      console.error('JWT 解析失敗:', error);
      return null;
    }
  }

  /**
   * 獲取當前使用者
   */
  getUser() {
    return this.user;
  }

  /**
   * 獲取 token（用於 API 呼叫）
   */
  getToken() {
    return this.user?.token || localStorage.getItem(this.storageKey);
  }

  /**
   * 檢查是否已登入
   */
  isAuthenticated() {
    return this.user !== null;
  }

  /**
   * 建立帶有認證的 fetch wrapper
   * 自動附加 Bearer token 並處理 401
   */
  createAuthFetch() {
    return async (url, options = {}) => {
      const token = this.getToken();

      if (token) {
        options.headers = {
          ...options.headers,
          'Authorization': `Bearer ${token}`,
        };
      }

      const response = await fetch(url, options);

      if (response.status === 401) {
        this.handleUnauthorized();
      }

      return response;
    };
  }

  /**
   * 建立測試使用者（開發用）
   */
  createTestUser(name = null) {
    console.warn('使用測試使用者（僅供開發）');

    const uniqueId = 'player_' + Date.now() + '_' + Math.random().toString(36).substring(2, 11);
    const playerNumber = Math.floor(Math.random() * 4) + 1;
    const playerName = name || `玩家${playerNumber}`;

    this.user = {
      id: uniqueId,
      name: playerName,
      picture: `https://ui-avatars.com/api/?name=${encodeURIComponent(playerName)}&background=random`,
      token: 'test_token_' + Date.now(),
      isGuest: true,
    };

    if (this.onSignIn) {
      this.onSignIn(this.user);
    }

    return this.user;
  }
}
