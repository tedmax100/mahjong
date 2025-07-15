// Phaser 遊戲設定
const config = {
    type: Phaser.AUTO,
    parent: 'game-container',
    width: 900,
    height: 600,
    backgroundColor: '#228B22',
    scene: {
        preload: preload,
        create: create,
        update: update
    }
};

let game = new Phaser.Game(config);

let socket = null;
let playerId = null;
let roomId = null;
let handTiles = []; // 玩家手牌

function preload() {
    // 你可以在這裡載入麻將牌圖檔（可用文字先代替）
}

function create() {
    this.add.text(400, 20, '台灣16張麻將', { fontSize: '32px', color: '#fff' });

    // 簡單畫出牌桌
    let centerX = this.cameras.main.centerX;
    let centerY = this.cameras.main.centerY;
    this.add.circle(centerX, centerY, 200, 0x006400);

    // 牌區（手牌）
    drawHand.call(this);

    // 操作按鈕
    let actions = ['出牌', '碰', '槓', '胡'];
    actions.forEach((act, i) => {
        let btn = this.add.text(50, 500 + i*30, act, { fontSize: '24px', color: '#fff', backgroundColor: '#444' })
            .setInteractive()
            .on('pointerdown', () => {
                if (socket) socket.send(JSON.stringify({ action: act, playerId, roomId }));
            });
    });
}

// 畫手牌（可自行改進）
function drawHand() {
    let startX = 200, startY = 500;
    for (let i = 0; i < handTiles.length; i++) {
        this.add.text(startX + i * 35, startY, handTiles[i] || '🀄', { fontSize: '28px', color: '#fff' });
    }
}

function update() {
    // 遊戲狀態更新
}

// 連接 WebSocket
function connectWS(room) {
    // 這裡請填寫你的後端 WebSocket 伺服器位址
    socket = new WebSocket('ws://localhost:8080/ws?room=' + room);

    socket.onopen = () => {
        console.log('WebSocket 連線成功');
        // 你可在這裡送出登入訊息
        socket.send(JSON.stringify({ action: 'join', roomId: room }));
    };

    socket.onmessage = (event) => {
        let msg = JSON.parse(event.data);
        console.log('收到訊息:', msg);

        // 根據收到的訊息更新遊戲狀態
        if (msg.type === 'hand') {
            handTiles = msg.tiles;
            game.scene.scenes[0].scene.restart(); // 重新畫面
        }
        // 其他訊息處理...
    };

    socket.onclose = () => {
        console.log('WebSocket 關閉');
    };
}

// UI 控制
document.getElementById('createBtn').onclick = () => {
    // 呼叫後端API建立房間（這裡僅模擬）
    roomId = Math.floor(Math.random()*100000).toString();
    document.getElementById('roomInput').value = roomId;
    connectWS(roomId);
    alert('已建立房間，ID: ' + roomId);
};

document.getElementById('joinBtn').onclick = () => {
    roomId = document.getElementById('roomInput').value;
    if (!roomId) return alert('請輸入房間ID');
    connectWS(roomId);
    alert('加入房間: ' + roomId);
};

document.getElementById('aiBtn').onclick = () => {
    if (socket) socket.send(JSON.stringify({ action: 'addAI', roomId }));
};

document.getElementById('startBtn').onclick = () => {
    if (socket) socket.send(JSON.stringify({ action: 'start', roomId }));
};

const mahjongEmojis = [
  '🀇','🀈','🀉','🀊','🀋','🀌','🀍','🀎','🀏', // 萬1~9
  '🀐','🀑','🀒','🀓','🀔','🀕','🀖','🀗','🀘', // 筒1~9
  '🀙','🀚','🀛','🀜','🀝','🀞','🀟','🀠','🀡', // 索1~9
  '🀀','🀁','🀂','🀃','🀄','🀅','🀆'           // 東南西北中發白
];
