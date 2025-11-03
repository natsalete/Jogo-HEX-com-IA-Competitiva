// ============================================
// CONFIGURAÇÕES E VARIÁVEIS GLOBAIS
// ============================================

const API_URL = "http://localhost:8080/api";
const BOARD_SIZE = 7;
const HEX_WIDTH = 40;
const HEX_HEIGHT = 46;
const HEX_SPACING_X = 35;
const HEX_SPACING_Y = 40;

let board = [];
let currentPlayer = "blue";
let gameActive = true;
let moveCount = 0;
let gameMode = "pvc";
let aiAlgorithm = "alphabeta";
let depth = 3;
let serverConnected = false;
let aiProcessing = false;
let winningPath = []; // NOVO: armazena o caminho vencedor

// ============================================
// VERIFICAÇÃO DE CONEXÃO COM SERVIDOR
// ============================================

async function checkServerConnection() {
  try {
    const response = await fetch(`${API_URL}/check-winner`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        size: BOARD_SIZE,
        cells: Array(BOARD_SIZE)
          .fill(null)
          .map(() => Array(BOARD_SIZE).fill("")),
      }),
    });

    if (response.ok) {
      serverConnected = true;
      updateServerStatus("connected");
      return true;
    }
  } catch (error) {
    serverConnected = false;
    updateServerStatus("error");
    return false;
  }
}

function updateServerStatus(status) {
  const statusEl = document.getElementById("serverStatus");
  const startBtn = document.getElementById("startBtn");

  if (status === "connected") {
    statusEl.className = "server-status connected";
    statusEl.innerHTML =
      '<strong>✅ Servidor Go conectado</strong><div style="margin-top: 5px; font-size: 0.85em;">Backend pronto em http://localhost:8080</div>';
    startBtn.disabled = false;
  } else if (status === "error") {
    statusEl.className = "server-status error";
    statusEl.innerHTML =
      '<strong>❌ Servidor Go não encontrado</strong><div style="margin-top: 5px; font-size: 0.85em;">Execute: cd server && go run main.go</div>';
    startBtn.disabled = true;
  } else {
    statusEl.className = "server-status";
    statusEl.innerHTML =
      '<strong>🔄 Conectando ao servidor...</strong><div style="margin-top: 5px; font-size: 0.85em;">http://localhost:8080</div>';
  }
}

// ============================================
// INICIALIZAÇÃO E RENDERIZAÇÃO
// ============================================

function initBoard() {
  gameMode = document.getElementById("gameMode").value;
  aiAlgorithm = document.getElementById("aiAlgorithm").value;
  depth = parseInt(document.getElementById("depth").value);

  board = Array(BOARD_SIZE)
    .fill(null)
    .map(() => Array(BOARD_SIZE).fill(""));
  currentPlayer = "blue";
  gameActive = true;
  moveCount = 0;
  aiProcessing = false;
  winningPath = []; // NOVO: limpa caminho vencedor

  updateStatus();
  renderBoard();
}

function renderBoard() {
  const boardElement = document.getElementById("board");
  boardElement.innerHTML = "";

  const totalWidth = (BOARD_SIZE + BOARD_SIZE - 1) * HEX_SPACING_X + HEX_WIDTH;
  const totalHeight = BOARD_SIZE * HEX_SPACING_Y + HEX_HEIGHT;

  boardElement.style.width = totalWidth + "px";
  boardElement.style.height = totalHeight + "px";

  for (let row = 0; row < BOARD_SIZE; row++) {
    for (let col = 0; col < BOARD_SIZE; col++) {
      const hex = createHexagon(row, col);
      boardElement.appendChild(hex);
    }
  }
}

function createHexagon(row, col) {
  const hex = document.createElement("div");
  hex.className = "hex";
  hex.dataset.row = row;
  hex.dataset.col = col;

  const x = col * HEX_SPACING_X + row * HEX_SPACING_X;
  const y = row * HEX_SPACING_Y;

  hex.style.left = x + "px";
  hex.style.top = y + "px";

  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("width", HEX_WIDTH);
  svg.setAttribute("height", HEX_HEIGHT);
  svg.setAttribute("viewBox", "0 0 40 46");

  const polygon = document.createElementNS(
    "http://www.w3.org/2000/svg",
    "polygon"
  );
  polygon.setAttribute("points", "20,2 38,12 38,34 20,44 2,34 2,12");
  polygon.setAttribute("class", "hex-polygon");

  svg.appendChild(polygon);
  hex.appendChild(svg);

  if (board[row][col]) {
    hex.classList.add(board[row][col]);
    hex.classList.add("disabled");

    // NOVO: destaca caminho vencedor
    if (isInWinningPath(row, col)) {
      hex.classList.add("winning-path");
    }
  } else if (!aiProcessing) {
    hex.addEventListener("click", () => handleHexClick(row, col));
  }

  return hex;
}

// NOVO: verifica se célula está no caminho vencedor
function isInWinningPath(row, col) {
  return winningPath.some((cell) => cell.row === row && cell.col === col);
}

// ============================================
// LÓGICA DE JOGADAS
// ============================================

function handleHexClick(row, col) {
  if (!gameActive || board[row][col] || aiProcessing) return;
  if (gameMode === "pvc" && currentPlayer === "red") return;
  if (gameMode === "cvc") return;
  if (!serverConnected) {
    alert("Servidor Go não conectado! Execute: cd server && go run main.go");
    return;
  }

  makeMove(row, col);
}

async function makeMove(row, col) {
  board[row][col] = currentPlayer;
  moveCount++;
  updateStatus();
  renderBoard();

  // Verifica vitória usando API do servidor
  const winner = await checkWinnerAPI();
  if (winner) {
    gameActive = false;
    // NOVO: encontra e anima caminho vencedor
    await findAndAnimateWinningPath(winner);
    showWinner(winner);
    return;
  }

  currentPlayer = currentPlayer === "blue" ? "red" : "blue";
  updateStatus();

  if (gameMode === "pvc" && currentPlayer === "red" && gameActive) {
    setTimeout(aiMove, 500);
  } else if (gameMode === "cvc" && gameActive) {
    setTimeout(aiMove, 800);
  }
}

// ============================================
// NOVO: DETECÇÃO E ANIMAÇÃO DO CAMINHO VENCEDOR
// ============================================

async function findAndAnimateWinningPath(winner) {
  winningPath = findWinningPath(winner);

  // Anima cada célula do caminho com delay progressivo
  for (let i = 0; i < winningPath.length; i++) {
    await new Promise((resolve) => setTimeout(resolve, 100)); // 100ms entre cada célula
    renderBoard(); // Re-renderiza para aplicar a classe winning-path
  }
}

function findWinningPath(player) {
  const size = BOARD_SIZE;
  const visited = Array(size)
    .fill(null)
    .map(() => Array(size).fill(false));
  const parent = Array(size)
    .fill(null)
    .map(() => Array(size).fill(null));

  // DFS modificado para rastrear o caminho
  function dfs(row, col, targetCondition) {
    if (targetCondition(row, col)) {
      return reconstructPath(row, col, parent);
    }

    visited[row][col] = true;
    const neighbors = getNeighbors(row, col);

    for (const n of neighbors) {
      if (!visited[n.row][n.col] && board[n.row][n.col] === player) {
        parent[n.row][n.col] = { row, col };
        const path = dfs(n.row, n.col, targetCondition);
        if (path) return path;
      }
    }
    return null;
  }

  // Azul: esquerda para direita
  if (player === "blue") {
    for (let row = 0; row < size; row++) {
      if (board[row][0] === player) {
        visited.forEach((r) => r.fill(false));
        parent.forEach((r) => r.fill(null));
        const path = dfs(row, 0, (r, c) => c === size - 1);
        if (path) return path;
      }
    }
  }
  // Vermelho: cima para baixo
  else {
    for (let col = 0; col < size; col++) {
      if (board[0][col] === player) {
        visited.forEach((r) => r.fill(false));
        parent.forEach((r) => r.fill(null));
        const path = dfs(0, col, (r, c) => r === size - 1);
        if (path) return path;
      }
    }
  }

  return [];
}

function reconstructPath(row, col, parent) {
  const path = [];
  let current = { row, col };

  while (current) {
    path.unshift(current);
    current = parent[current.row][current.col];
  }

  return path;
}

function getNeighbors(row, col) {
  const directions = [
    [-1, 0],
    [-1, 1],
    [0, -1],
    [0, 1],
    [1, -1],
    [1, 0],
  ];

  const neighbors = [];
  for (const [dr, dc] of directions) {
    const newRow = row + dr;
    const newCol = col + dc;
    if (
      newRow >= 0 &&
      newRow < BOARD_SIZE &&
      newCol >= 0 &&
      newCol < BOARD_SIZE
    ) {
      neighbors.push({ row: newRow, col: newCol });
    }
  }
  return neighbors;
}

// ============================================
// INTEGRAÇÃO COM BACKEND GO
// ============================================

async function aiMove() {
  if (!serverConnected || aiProcessing) return;

  aiProcessing = true;
  document.getElementById("aiTime").textContent = "Processando...";
  renderBoard();

  try {
    const response = await fetch(`${API_URL}/ai-move`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        board: board,
        player: currentPlayer,
        algorithm: aiAlgorithm,
        depth: depth,
      }),
    });

    if (!response.ok) {
      throw new Error("Erro na resposta do servidor");
    }

    const data = await response.json();

    document.getElementById("nodesExplored").textContent =
      data.nodesExplored.toLocaleString("pt-BR");
    document.getElementById("aiTime").textContent = data.timeMs + " ms";
    document.getElementById("aiScore").textContent = data.score;

    aiProcessing = false;

    if (data.move && gameActive) {
      await makeMove(data.move.row, data.move.col);
    }
  } catch (error) {
    console.error("Erro ao comunicar com servidor:", error);
    aiProcessing = false;
    document.getElementById("aiTime").textContent = "Erro!";
    alert(
      "Erro ao conectar com servidor Go. Verifique se está rodando em localhost:8080"
    );
    serverConnected = false;
    updateServerStatus("error");
  }
}

async function checkWinnerAPI() {
  try {
    const response = await fetch(`${API_URL}/check-winner`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        size: BOARD_SIZE,
        cells: board,
      }),
    });

    if (response.ok) {
      const data = await response.json();
      return data.winner || null;
    }
  } catch (error) {
    console.error("Erro ao verificar vencedor:", error);
  }
  return null;
}

// ============================================
// INTERFACE E CONTROLES
// ============================================

function showWinner(winner) {
  const winnerText =
    winner === "blue"
      ? "🎉 Jogador Azul Venceu!"
      : "🎉 Jogador Vermelho Venceu!";
  document.getElementById("winnerText").textContent = winnerText;

  const nodesText = document.getElementById("nodesExplored").textContent;
  const pathLength = winningPath.length;
  document.getElementById(
    "winnerStats"
  ).textContent = `Jogadas: ${moveCount} | Caminho: ${pathLength} células | Nós: ${nodesText} | Algoritmo: ${
    aiAlgorithm === "alphabeta" ? "Alfa-Beta" : "Minimax"
  }`;

  document.getElementById("winnerModal").style.display = "flex";
}

function closeModal() {
  document.getElementById("winnerModal").style.display = "none";
  startGame();
}

function updateStatus() {
  document.getElementById("currentPlayer").textContent =
    currentPlayer === "blue" ? "Azul" : "Vermelho";
  document.getElementById("moveCount").textContent = moveCount;
}

async function startGame() {
  if (!serverConnected) {
    const connected = await checkServerConnection();
    if (!connected) {
      alert(
        "Servidor Go não está rodando!\n\nPara iniciar:\n1. Abra terminal na pasta do projeto\n2. Execute: cd server\n3. Execute: go run main.go\n4. Recarregue esta página"
      );
      return;
    }
  }

  initBoard();

  if (gameMode === "cvc") {
    setTimeout(aiMove, 1000);
  }
}

function resetGame() {
  startGame();
}

// ============================================
// INICIALIZAÇÃO
// ============================================

window.onload = async () => {
  console.log("🎮 HEX Game inicializado - Tabuleiro 7x7");
  console.log("✨ Caminho vencedor animado ativado!");
  console.log("🔗 Verificando conexão com servidor Go...");

  initBoard();
  await checkServerConnection();

  if (serverConnected) {
    console.log("✅ Servidor Go conectado!");
  } else {
    console.log(
      "❌ Servidor Go não encontrado. Execute: go run server/main.go"
    );
  }
};
