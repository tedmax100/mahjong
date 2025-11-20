/**
 * Google OAuth 认证管理
 */
export class GoogleAuth {
  constructor() {
    this.user = null;
    this.onSignIn = null;
  }

  /**
   * 初始化Google Identity Services
   */
  init() {
    if (typeof google !== 'undefined' && google.accounts) {
      google.accounts.id.initialize({
        client_id: '28393022343-u385bvffeq6p3l7k7fdltrjerpr04286.apps.googleusercontent.com',
        callback: this.handleCredentialResponse.bind(this)
      });

      google.accounts.id.renderButton(
        document.getElementById('google-login-btn'),
        { theme: 'outline', size: 'large', text: 'signin_with', locale: 'zh_TW' }
      );

      // 自动登入提示
      google.accounts.id.prompt();
    }
  }

  /**
   * 处理Google登入响应
   */
  handleCredentialResponse(response) {
    const credential = response.credential;
    const payload = this.parseJwt(credential);

    this.user = {
      id: payload.sub,
      email: payload.email,
      name: payload.name,
      picture: payload.picture,
      token: credential
    };

    console.log('Google登入成功:', this.user);

    if (this.onSignIn) {
      this.onSignIn(this.user);
    }
  }

  /**
   * 手动触发登入（备用方案）
   */
  signIn() {
    if (typeof google !== 'undefined' && google.accounts) {
      google.accounts.id.prompt();
    } else {
      // 如果Google Identity Services未加载，使用简单的测试登入
      console.warn('Google Identity Services未加载，使用测试用户');

      // 为每个浏览器标签生成唯一ID
      const uniqueId = 'player_' + Date.now() + '_' + Math.random().toString(36).substring(2, 11);
      const playerNumber = Math.floor(Math.random() * 4) + 1;

      this.user = {
        id: uniqueId,
        email: `player${playerNumber}@test.com`,
        name: `玩家${playerNumber}`,
        picture: `https://ui-avatars.com/api/?name=P${playerNumber}&background=random`,
        token: 'test_token_' + Date.now()
      };

      console.log('创建测试用户:', this.user);

      if (this.onSignIn) {
        this.onSignIn(this.user);
      }
    }
  }

  /**
   * 登出
   */
  signOut() {
    if (typeof google !== 'undefined' && google.accounts) {
      google.accounts.id.disableAutoSelect();
    }
    this.user = null;
  }

  /**
   * 解析JWT token
   */
  parseJwt(token) {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  }

  /**
   * 获取当前用户
   */
  getUser() {
    return this.user;
  }
}
