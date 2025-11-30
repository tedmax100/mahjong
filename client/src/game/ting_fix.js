
  /**
   * 檢查自己摸牌後可以執行的動作（自摸、聽牌、暗槓）
   */
  checkSelfActions() {
    const myPlayer = this.players[this.myPosition];
    if (!myPlayer || !myPlayer.tiles || this.myPosition !== this.currentTurn) {
      return;
    }

    // Ask the server to check for possible actions
    if (this.ws) {
      this.ws.sendAction('check_ting');
    }
  }

  /**
   * 处理服务器返回的听牌结果
   * @param {Object} data - key: discardable tile, value: list of winning tiles
   */
  handleTingResult(data) {
    this.possibleTingDiscards = data;
    const actions = [];

    if (Object.keys(this.possibleTingDiscards).length > 0) {
        actions.push('ready');
        console.log('🎯 聽牌！可打出的牌:', Object.keys(this.possibleTingDiscards));
    }

    // For now, let's not check for self-drawn win on the client
    // as the canHu logic is buggy. Server will handle it.

    if (actions.length > 0) {
      this.actionButtons.show([...actions, 'cancel']);
      this.pendingActions = actions;
    }
  }
