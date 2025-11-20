/**
 * 麻将牌素材生成工具
 * 使用Canvas API生成PNG图片
 */

const { createCanvas } = require('canvas');
const fs = require('fs');
const path = require('path');

// 配置
const TILE_WIDTH = 60;
const TILE_HEIGHT = 80;
const OUTPUT_DIR = path.join(__dirname, '../client/public/assets/tiles');

// 确保输出目录存在
if (!fs.existsSync(OUTPUT_DIR)) {
  fs.mkdirSync(OUTPUT_DIR, { recursive: true });
}

// 颜色配置
const COLORS = {
  background: '#F5F0E8',
  border: '#8B7355',
  borderDark: '#5C4A3A',
  wan: '#2C5F2D',      // 万子 - 深绿
  tong: '#1E5A8E',     // 筒子 - 深蓝
  tiao: '#8B4513',     // 条子 - 棕色
  wind: '#1A1A1A',     // 风牌 - 黑色
  dragon: {
    zhong: '#C41E3A',  // 中 - 红色
    fa: '#2C5F2D',     // 发 - 绿色
    bai: '#4169E1'     // 白 - 蓝色边框
  },
  flower: '#8B008B'    // 花牌 - 紫色
};

/**
 * 绘制牌的基础框架
 */
function drawTileBase(ctx) {
  // 白色背景
  ctx.fillStyle = COLORS.background;
  ctx.fillRect(0, 0, TILE_WIDTH, TILE_HEIGHT);

  // 外边框
  ctx.strokeStyle = COLORS.borderDark;
  ctx.lineWidth = 2;
  ctx.strokeRect(1, 1, TILE_WIDTH - 2, TILE_HEIGHT - 2);

  // 内边框
  ctx.strokeStyle = COLORS.border;
  ctx.lineWidth = 1;
  ctx.strokeRect(3, 3, TILE_WIDTH - 6, TILE_HEIGHT - 6);
}

/**
 * 绘制数字（万子、筒子、条子）
 */
function drawNumber(ctx, number, type) {
  const color = type === 'wan' ? COLORS.wan :
                type === 'tong' ? COLORS.tong : COLORS.tiao;

  ctx.fillStyle = color;
  ctx.font = 'bold 24px Arial';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';

  // 绘制数字
  ctx.fillText(number.toString(), TILE_WIDTH / 2, TILE_HEIGHT * 0.6);

  // 绘制类型文字
  ctx.font = 'bold 14px Arial';
  const typeText = type === 'wan' ? '萬' : type === 'tong' ? '筒' : '條';
  ctx.fillText(typeText, TILE_WIDTH / 2, TILE_HEIGHT * 0.3);

  // 绘制装饰圆点或线条
  if (type === 'tong') {
    drawTongPattern(ctx, number);
  } else if (type === 'tiao') {
    drawTiaoPattern(ctx, number);
  }
}

/**
 * 绘制筒子图案
 */
function drawTongPattern(ctx, number) {
  const positions = getPatternPositions(number);
  ctx.fillStyle = COLORS.tong;

  positions.forEach(([x, y]) => {
    ctx.beginPath();
    ctx.arc(x, y, 3, 0, Math.PI * 2);
    ctx.fill();
  });
}

/**
 * 绘制条子图案
 */
function drawTiaoPattern(ctx, number) {
  const centerX = TILE_WIDTH / 2;
  const startY = 15;
  const spacing = 3;

  ctx.strokeStyle = COLORS.tiao;
  ctx.lineWidth = 2;

  for (let i = 0; i < number; i++) {
    const y = startY + i * spacing;
    ctx.beginPath();
    ctx.moveTo(centerX - 8, y);
    ctx.lineTo(centerX + 8, y);
    ctx.stroke();
  }
}

/**
 * 获取圆点图案位置
 */
function getPatternPositions(number) {
  const cx = TILE_WIDTH / 2;
  const positions = {
    1: [[cx, 20]],
    2: [[cx - 6, 18], [cx + 6, 18]],
    3: [[cx - 8, 17], [cx, 17], [cx + 8, 17]],
    4: [[cx - 6, 15], [cx + 6, 15], [cx - 6, 23], [cx + 6, 23]],
    5: [[cx - 8, 14], [cx + 8, 14], [cx, 19], [cx - 8, 24], [cx + 8, 24]],
    6: [[cx - 8, 13], [cx, 13], [cx + 8, 13], [cx - 8, 21], [cx, 21], [cx + 8, 21]],
    7: [[cx - 9, 12], [cx - 3, 12], [cx + 3, 12], [cx + 9, 12], [cx - 6, 19], [cx, 19], [cx + 6, 19]],
    8: [[cx - 9, 11], [cx - 3, 11], [cx + 3, 11], [cx + 9, 11], [cx - 9, 18], [cx - 3, 18], [cx + 3, 18], [cx + 9, 18]],
    9: [[cx - 9, 10], [cx - 3, 10], [cx + 3, 10], [cx + 9, 10], [cx - 6, 16], [cx, 16], [cx + 6, 16], [cx - 9, 22], [cx + 9, 22]]
  };

  return positions[number] || [];
}

/**
 * 绘制风牌
 */
function drawWind(ctx, wind) {
  const windChars = { dong: '東', nan: '南', xi: '西', bei: '北' };

  ctx.fillStyle = COLORS.wind;
  ctx.font = 'bold 32px Arial';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(windChars[wind], TILE_WIDTH / 2, TILE_HEIGHT / 2);
}

/**
 * 绘制三元牌
 */
function drawDragon(ctx, dragon) {
  const dragonChars = { zhong: '中', fa: '發', bai: '白' };
  const color = dragon === 'zhong' ? COLORS.dragon.zhong :
                dragon === 'fa' ? COLORS.dragon.fa : COLORS.dragon.bai;

  ctx.fillStyle = color;
  ctx.font = 'bold 32px Arial';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';

  if (dragon === 'bai') {
    // 白板用边框表示
    ctx.strokeStyle = color;
    ctx.lineWidth = 3;
    ctx.strokeRect(10, 15, TILE_WIDTH - 20, TILE_HEIGHT - 30);
  } else {
    ctx.fillText(dragonChars[dragon], TILE_WIDTH / 2, TILE_HEIGHT / 2);
  }
}

/**
 * 绘制花牌
 */
function drawFlower(ctx, flower) {
  const flowerChars = {
    chun: '春', xia: '夏', qiu: '秋', dong: '冬',
    mei: '梅', lan: '蘭', zhu: '竹', ju: '菊'
  };

  ctx.fillStyle = COLORS.flower;
  ctx.font = 'bold 28px Arial';
  ctx.textAlign = 'center';
  ctx.textBaseline = 'middle';
  ctx.fillText(flowerChars[flower], TILE_WIDTH / 2, TILE_HEIGHT / 2);

  // 添加小字标注
  ctx.font = '10px Arial';
  ctx.fillText('花', TILE_WIDTH / 2, TILE_HEIGHT - 12);
}

/**
 * 绘制牌背
 */
function drawTileBack(ctx) {
  // 深绿色背景
  ctx.fillStyle = '#1B5E20';
  ctx.fillRect(0, 0, TILE_WIDTH, TILE_HEIGHT);

  // 边框
  ctx.strokeStyle = COLORS.borderDark;
  ctx.lineWidth = 2;
  ctx.strokeRect(1, 1, TILE_WIDTH - 2, TILE_HEIGHT - 2);

  // 花纹
  ctx.strokeStyle = '#2E7D32';
  ctx.lineWidth = 1;

  for (let i = 0; i < 5; i++) {
    const y = 15 + i * 15;
    ctx.beginPath();
    ctx.moveTo(10, y);
    ctx.lineTo(TILE_WIDTH - 10, y);
    ctx.stroke();
  }

  // 中心装饰
  ctx.fillStyle = '#388E3C';
  ctx.fillRect(TILE_WIDTH / 2 - 8, TILE_HEIGHT / 2 - 8, 16, 16);
}

/**
 * 保存Canvas为PNG
 */
function saveTile(canvas, filename) {
  const buffer = canvas.toBuffer('image/png');
  const filepath = path.join(OUTPUT_DIR, filename);
  fs.writeFileSync(filepath, buffer);
  console.log(`✓ ${filename}`);
}

/**
 * 主函数
 */
function generateTiles() {
  console.log('开始生成麻将牌素材...\n');

  // 生成万子 1-9
  for (let i = 1; i <= 9; i++) {
    const canvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
    const ctx = canvas.getContext('2d');
    drawTileBase(ctx);
    drawNumber(ctx, i, 'wan');
    saveTile(canvas, `wan-${i}.png`);
  }

  // 生成筒子 1-9
  for (let i = 1; i <= 9; i++) {
    const canvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
    const ctx = canvas.getContext('2d');
    drawTileBase(ctx);
    drawNumber(ctx, i, 'tong');
    saveTile(canvas, `tong-${i}.png`);
  }

  // 生成条子 1-9
  for (let i = 1; i <= 9; i++) {
    const canvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
    const ctx = canvas.getContext('2d');
    drawTileBase(ctx);
    drawNumber(ctx, i, 'tiao');
    saveTile(canvas, `tiao-${i}.png`);
  }

  // 生成风牌
  ['dong', 'nan', 'xi', 'bei'].forEach(wind => {
    const canvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
    const ctx = canvas.getContext('2d');
    drawTileBase(ctx);
    drawWind(ctx, wind);
    saveTile(canvas, `${wind}.png`);
  });

  // 生成三元牌
  ['zhong', 'fa', 'bai'].forEach(dragon => {
    const canvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
    const ctx = canvas.getContext('2d');
    drawTileBase(ctx);
    drawDragon(ctx, dragon);
    saveTile(canvas, `${dragon}.png`);
  });

  // 生成花牌
  ['chun', 'xia', 'qiu', 'dong', 'mei', 'lan', 'zhu', 'ju'].forEach(flower => {
    const canvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
    const ctx = canvas.getContext('2d');
    drawTileBase(ctx);
    drawFlower(ctx, flower);
    saveTile(canvas, `flower-${flower}.png`);
  });

  // 生成牌背
  const backCanvas = createCanvas(TILE_WIDTH, TILE_HEIGHT);
  const backCtx = backCanvas.getContext('2d');
  drawTileBack(backCtx);
  saveTile(backCanvas, 'back.png');

  console.log(`\n✓ 所有素材生成完成！共 ${fs.readdirSync(OUTPUT_DIR).length} 个文件`);
  console.log(`输出目录: ${OUTPUT_DIR}`);
}

// 执行生成
try {
  generateTiles();
} catch (error) {
  console.error('生成失败:', error.message);
  console.log('\n请先安装依赖: npm install canvas');
}
