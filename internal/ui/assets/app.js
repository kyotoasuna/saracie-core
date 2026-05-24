const $ = (id) => document.getElementById(id);

let loadedWallet = null;
let timer = null;

async function api(path, options = {}) {
  const response = await fetch(path, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text.trim() || response.statusText);
  }
  return response.json();
}

function post(path, body = {}) {
  return api(path, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

function show(message) {
  const toast = $("toast");
  toast.textContent = message;
  toast.classList.add("visible");
  clearTimeout(timer);
  timer = setTimeout(() => toast.classList.remove("visible"), 3600);
}

function walletInput() {
  return {
    wallet: $("wallet-file").value.trim() || "saracie.wallet",
    passphrase: $("wallet-passphrase").value,
  };
}

function applyWallet(info) {
  loadedWallet = info;
  $("wallet-state").textContent = "Open";
  $("wallet-address").textContent = info.address;
  $("balance-address").value = info.address;
  $("miner-address").value = info.address;
}

async function refresh() {
  const status = await api("/api/status");
  $("height").textContent = status.chain.height;
  $("reward").textContent = status.chain.block_reward;
  $("mempool").textContent = status.chain.mempool_count;
  $("tip-hash").textContent = status.chain.tip_hash;

  const miner = status.miner;
  $("blocks-mined").textContent = `${miner.blocks_mined} blocks`;
  $("last-reward").textContent = miner.last_reward || "0.00000000";
  $("last-height").textContent = miner.last_height ? `Height ${miner.last_height}` : "No blocks";
  $("last-hash").textContent = miner.last_hash || miner.last_error || "waiting for mining activity";

  const pill = $("miner-pill");
  pill.textContent = miner.running ? "Running" : "Stopped";
  pill.classList.toggle("running", miner.running);
}

async function createWallet() {
  const info = await post("/api/wallet/create", walletInput());
  applyWallet(info);
  show("Wallet created");
}

async function openWallet() {
  const info = await post("/api/wallet/open", walletInput());
  applyWallet(info);
  show("Wallet opened");
}

async function checkBalance(event) {
  event.preventDefault();
  const address = $("balance-address").value.trim();
  const balance = await api(`/api/balance?address=${encodeURIComponent(address)}`);
  $("confirmed-balance").textContent = balance.confirmed;
  $("spendable-balance").textContent = balance.spendable;
  show("Balance updated");
}

async function sendTransaction(event) {
  event.preventDefault();
  const input = walletInput();
  const tx = await post("/api/send-file", {
    ...input,
    to: $("send-to").value.trim(),
    amount: $("send-amount").value.trim(),
    fee: $("send-fee").value.trim(),
  });
  $("send-result").textContent = tx.id;
  show("Transaction added to mempool");
  await refresh();
}

async function startMiner() {
  const address = $("miner-address").value.trim();
  await post("/api/miner/start", { address });
  show("Miner started");
  await refresh();
}

async function stopMiner() {
  await post("/api/miner/stop", {});
  show("Miner stopped");
  await refresh();
}

function wire() {
  $("refresh").addEventListener("click", () => refresh().catch((error) => show(error.message)));
  $("create-wallet").addEventListener("click", () => createWallet().catch((error) => show(error.message)));
  $("open-wallet").addEventListener("click", () => openWallet().catch((error) => show(error.message)));
  $("balance-form").addEventListener("submit", (event) => checkBalance(event).catch((error) => show(error.message)));
  $("send-form").addEventListener("submit", (event) => sendTransaction(event).catch((error) => show(error.message)));
  $("start-miner").addEventListener("click", () => startMiner().catch((error) => show(error.message)));
  $("stop-miner").addEventListener("click", () => stopMiner().catch((error) => show(error.message)));
}

const canvas = $("network-canvas");
const ctx = canvas.getContext("2d");
let width = 0;
let height = 0;
let points = [];

function resizeCanvas() {
  const dpr = Math.min(window.devicePixelRatio || 1, 2);
  width = window.innerWidth;
  height = window.innerHeight;
  canvas.width = Math.floor(width * dpr);
  canvas.height = Math.floor(height * dpr);
  canvas.style.width = `${width}px`;
  canvas.style.height = `${height}px`;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  const count = Math.max(24, Math.min(68, Math.floor(width / 26)));
  points = Array.from({ length: count }, (_, i) => ({
    x: Math.random() * width,
    y: Math.random() * height,
    vx: (Math.random() - 0.5) * 0.22,
    vy: (Math.random() - 0.5) * 0.22,
    block: i % 8 === 0,
  }));
}

function drawCanvas() {
  ctx.clearRect(0, 0, width, height);
  ctx.fillStyle = "#11120f";
  ctx.fillRect(0, 0, width, height);

  for (const point of points) {
    point.x += point.vx;
    point.y += point.vy;
    if (point.x < -20) point.x = width + 20;
    if (point.x > width + 20) point.x = -20;
    if (point.y < -20) point.y = height + 20;
    if (point.y > height + 20) point.y = -20;
  }

  for (let i = 0; i < points.length; i += 1) {
    for (let j = i + 1; j < points.length; j += 1) {
      const a = points[i];
      const b = points[j];
      const dx = a.x - b.x;
      const dy = a.y - b.y;
      const distance = Math.sqrt(dx * dx + dy * dy);
      if (distance < 142) {
        ctx.strokeStyle = `rgba(226,185,93,${(1 - distance / 142) * 0.15})`;
        ctx.beginPath();
        ctx.moveTo(a.x, a.y);
        ctx.lineTo(b.x, b.y);
        ctx.stroke();
      }
    }
  }

  for (const point of points) {
    ctx.fillStyle = point.block ? "rgba(228,77,77,0.72)" : "rgba(95,209,140,0.52)";
    if (point.block) {
      ctx.fillRect(point.x - 4, point.y - 4, 8, 8);
    } else {
      ctx.beginPath();
      ctx.arc(point.x, point.y, 2.2, 0, Math.PI * 2);
      ctx.fill();
    }
  }
  requestAnimationFrame(drawCanvas);
}

window.addEventListener("resize", resizeCanvas);
wire();
resizeCanvas();
drawCanvas();
refresh().catch((error) => show(error.message));
setInterval(() => refresh().catch(() => {}), 2500);
