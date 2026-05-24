const canvas = document.getElementById("chain-canvas");
const ctx = canvas.getContext("2d");

let width = 0;
let height = 0;
let nodes = [];

function resize() {
  const dpr = Math.min(window.devicePixelRatio || 1, 2);
  width = window.innerWidth;
  height = window.innerHeight;
  canvas.width = Math.floor(width * dpr);
  canvas.height = Math.floor(height * dpr);
  canvas.style.width = `${width}px`;
  canvas.style.height = `${height}px`;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

  const count = Math.max(28, Math.min(70, Math.floor(width / 24)));
  nodes = Array.from({ length: count }, (_, index) => ({
    x: Math.random() * width,
    y: Math.random() * height,
    vx: (Math.random() - 0.5) * 0.24,
    vy: (Math.random() - 0.5) * 0.24,
    pulse: Math.random() * Math.PI * 2,
    block: index % 9 === 0,
  }));
}

function draw() {
  ctx.clearRect(0, 0, width, height);
  ctx.fillStyle = "#10110e";
  ctx.fillRect(0, 0, width, height);

  for (const node of nodes) {
    node.x += node.vx;
    node.y += node.vy;
    node.pulse += 0.015;

    if (node.x < -20) node.x = width + 20;
    if (node.x > width + 20) node.x = -20;
    if (node.y < -20) node.y = height + 20;
    if (node.y > height + 20) node.y = -20;
  }

  for (let i = 0; i < nodes.length; i += 1) {
    for (let j = i + 1; j < nodes.length; j += 1) {
      const a = nodes[i];
      const b = nodes[j];
      const dx = a.x - b.x;
      const dy = a.y - b.y;
      const distance = Math.sqrt(dx * dx + dy * dy);
      if (distance < 150) {
        const alpha = 1 - distance / 150;
        ctx.strokeStyle = `rgba(228, 185, 91, ${alpha * 0.16})`;
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(a.x, a.y);
        ctx.lineTo(b.x, b.y);
        ctx.stroke();
      }
    }
  }

  for (const node of nodes) {
    const glow = (Math.sin(node.pulse) + 1) / 2;
    if (node.block) {
      ctx.fillStyle = `rgba(221, 59, 59, ${0.78 + glow * 0.18})`;
      ctx.fillRect(node.x - 5, node.y - 5, 10, 10);
    } else {
      ctx.fillStyle = `rgba(84, 199, 136, ${0.45 + glow * 0.24})`;
      ctx.beginPath();
      ctx.arc(node.x, node.y, 2.4, 0, Math.PI * 2);
      ctx.fill();
    }
  }

  ctx.fillStyle = "rgba(16, 17, 14, 0.34)";
  ctx.fillRect(0, 0, width, height);
  requestAnimationFrame(draw);
}

window.addEventListener("resize", resize);
resize();
draw();
