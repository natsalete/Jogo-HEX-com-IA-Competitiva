 // ============================================
        // VARIÁVEIS GLOBAIS
        // ============================================
        
        const BOARD_SIZE = 7; // Tabuleiro fixo 7x7
        const HEX_WIDTH = 40;
        const HEX_HEIGHT = 46;
        const HEX_SPACING_X = 35;
        const HEX_SPACING_Y = 40;

        let board = [];
        let currentPlayer = 'blue';
        let gameActive = true;
        let moveCount = 0;
        let gameMode = 'pvc';
        let aiAlgorithm = 'alphabeta';
        let depth = 2;
        let nodesExplored = 0;

        // ============================================
        // INICIALIZAÇÃO E RENDERIZAÇÃO
        // ============================================

        /**
         * Inicializa o tabuleiro do jogo
         */
        function initBoard() {
            // Obtém configurações da interface
            gameMode = document.getElementById('gameMode').value;
            aiAlgorithm = document.getElementById('aiAlgorithm').value;
            depth = parseInt(document.getElementById('depth').value);
            
            // Reseta estado do jogo
            board = Array(BOARD_SIZE).fill(null).map(() => Array(BOARD_SIZE).fill(null));
            currentPlayer = 'blue';
            gameActive = true;
            moveCount = 0;
            nodesExplored = 0;
            
            updateStatus();
            renderBoard();
        }

        /**
         * Renderiza o tabuleiro hexagonal na tela
         */
        function renderBoard() {
            const boardElement = document.getElementById('board');
            boardElement.innerHTML = '';
            
            // Calcula dimensões totais do tabuleiro
            const totalWidth = (BOARD_SIZE + BOARD_SIZE - 1) * HEX_SPACING_X + HEX_WIDTH;
            const totalHeight = BOARD_SIZE * HEX_SPACING_Y + HEX_HEIGHT;
            
            boardElement.style.width = totalWidth + 'px';
            boardElement.style.height = totalHeight + 'px';

            // Cria todos os hexágonos
            for (let row = 0; row < BOARD_SIZE; row++) {
                for (let col = 0; col < BOARD_SIZE; col++) {
                    const hex = createHexagon(row, col);
                    boardElement.appendChild(hex);
                }
            }
        }

        /**
         * Cria um hexágono individual
         * @param {number} row - Linha do hexágono
         * @param {number} col - Coluna do hexágono
         * @returns {HTMLElement} Elemento do hexágono
         */
        function createHexagon(row, col) {
            const hex = document.createElement('div');
            hex.className = 'hex';
            hex.dataset.row = row;
            hex.dataset.col = col;
            
            // Calcula posição do hexágono (padrão hexagonal)
            const x = col * HEX_SPACING_X + row * HEX_SPACING_X;
            const y = row * HEX_SPACING_Y;
            
            hex.style.left = x + 'px';
            hex.style.top = y + 'px';

            // Cria SVG do hexágono
            const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
            svg.setAttribute('width', HEX_WIDTH);
            svg.setAttribute('height', HEX_HEIGHT);
            svg.setAttribute('viewBox', '0 0 40 46');

            const polygon = document.createElementNS('http://www.w3.org/2000/svg', 'polygon');
            polygon.setAttribute('points', '20,2 38,12 38,34 20,44 2,34 2,12');
            polygon.setAttribute('class', 'hex-polygon');
            
            svg.appendChild(polygon);
            hex.appendChild(svg);

            // Aplica cor se a célula estiver ocupada
            if (board[row][col]) {
                hex.classList.add(board[row][col]);
                hex.classList.add('disabled');
            } else {
                hex.addEventListener('click', () => handleHexClick(row, col));
            }

            return hex;
        }

        // ============================================
        // LÓGICA DE JOGADAS
        // ============================================

        /**
         * Manipula clique em um hexágono
         */
        function handleHexClick(row, col) {
            // Valida se a jogada é permitida
            if (!gameActive || board[row][col]) return;
            if (gameMode === 'pvc' && currentPlayer === 'red') return;
            if (gameMode === 'cvc') return;

            makeMove(row, col);
        }

        /**
         * Executa uma jogada no tabuleiro
         */
        function makeMove(row, col) {
            board[row][col] = currentPlayer;
            moveCount++;
            updateStatus();
            renderBoard();

            // Verifica vitória
            if (checkWinner(currentPlayer)) {
                gameActive = false;
                showWinner(currentPlayer);
                return;
            }

            // Alterna jogador
            currentPlayer = currentPlayer === 'blue' ? 'red' : 'blue';
            updateStatus();

            // Ativa IA se necessário
            if (gameMode === 'pvc' && currentPlayer === 'red' && gameActive) {
                setTimeout(aiMove, 500);
            } else if (gameMode === 'cvc' && gameActive) {
                setTimeout(aiMove, 800);
            }
        }

        // ============================================
        // INTELIGÊNCIA ARTIFICIAL
        // ============================================

        /**
         * Executa movimento da IA
         */
        function aiMove() {
            const startTime = performance.now();
            nodesExplored = 0;
            
            // Escolhe algoritmo
            let bestMove;
            if (aiAlgorithm === 'minimax') {
                bestMove = findBestMoveMinimax();
            } else {
                bestMove = findBestMoveAlphaBeta();
            }
            
            const endTime = performance.now();
            const aiTime = (endTime - startTime).toFixed(2);
            
            // Atualiza métricas
            document.getElementById('aiTime').textContent = aiTime + ' ms';
            document.getElementById('nodesExplored').textContent = nodesExplored;

            if (bestMove) {
                makeMove(bestMove.row, bestMove.col);
            }
        }

        /**
         * Encontra melhor jogada usando Minimax puro
         */
        function findBestMoveMinimax() {
            let bestScore = -Infinity;
            let bestMove = null;

            const moves = getValidMoves();
            
            for (const move of moves) {
                board[move.row][move.col] = currentPlayer;
                const score = minimax(depth - 1, false);
                board[move.row][move.col] = null;

                if (score > bestScore) {
                    bestScore = score;
                    bestMove = move;
                }
            }

            return bestMove;
        }

        /**
         * Algoritmo Minimax recursivo
         * @param {number} depthLeft - Profundidade restante
         * @param {boolean} isMaximizing - Se é turno do maximizador
         */
        function minimax(depthLeft, isMaximizing) {
            nodesExplored++;

            // Verifica condições de parada
            const winner = getWinner();
            if (winner === currentPlayer) return 1000 + depthLeft;
            if (winner === getOpponent(currentPlayer)) return -1000 - depthLeft;
            if (depthLeft === 0) return evaluateBoard();

            const player = isMaximizing ? currentPlayer : getOpponent(currentPlayer);
            let bestScore = isMaximizing ? -Infinity : Infinity;
            const moves = getValidMoves();

            // Explora todos os movimentos possíveis
            for (const move of moves) {
                board[move.row][move.col] = player;
                const score = minimax(depthLeft - 1, !isMaximizing);
                board[move.row][move.col] = null;

                if (isMaximizing) {
                    bestScore = Math.max(bestScore, score);
                } else {
                    bestScore = Math.min(bestScore, score);
                }
            }

            return bestScore;
        }

        /**
         * Encontra melhor jogada usando Minimax com Poda Alfa-Beta
         */
        function findBestMoveAlphaBeta() {
            let bestScore = -Infinity;
            let bestMove = null;
            let alpha = -Infinity;
            let beta = Infinity;

            const moves = getValidMoves();
            
            for (const move of moves) {
                board[move.row][move.col] = currentPlayer;
                const score = alphaBeta(depth - 1, alpha, beta, false);
                board[move.row][move.col] = null;

                if (score > bestScore) {
                    bestScore = score;
                    bestMove = move;
                }
                alpha = Math.max(alpha, score);
            }

            return bestMove;
        }

        /**
         * Algoritmo Alfa-Beta (Minimax otimizado)
         * @param {number} depthLeft - Profundidade restante
         * @param {number} alpha - Melhor valor para maximizador
         * @param {number} beta - Melhor valor para minimizador
         * @param {boolean} isMaximizing - Se é turno do maximizador
         */
        function alphaBeta(depthLeft, alpha, beta, isMaximizing) {
            nodesExplored++;

            // Verifica condições de parada
            const winner = getWinner();
            if (winner === currentPlayer) return 1000 + depthLeft;
            if (winner === getOpponent(currentPlayer)) return -1000 - depthLeft;
            if (depthLeft === 0) return evaluateBoard();

            const player = isMaximizing ? currentPlayer : getOpponent(currentPlayer);
            const moves = getValidMoves();

            if (isMaximizing) {
                let maxScore = -Infinity;
                for (const move of moves) {
                    board[move.row][move.col] = player;
                    const score = alphaBeta(depthLeft - 1, alpha, beta, false);
                    board[move.row][move.col] = null;
                    
                    maxScore = Math.max(maxScore, score);
                    alpha = Math.max(alpha, score);
                    
                    // Poda Beta: corta ramo se beta <= alpha
                    if (beta <= alpha) break;
                }
                return maxScore;
            } else {
                let minScore = Infinity;
                for (const move of moves) {
                    board[move.row][move.col] = player;
                    const score = alphaBeta(depthLeft - 1, alpha, beta, true);
                    board[move.row][move.col] = null;
                    
                    minScore = Math.min(minScore, score);
                    beta = Math.min(beta, score);
                    
                    // Poda Alpha: corta ramo se beta <= alpha
                    if (beta <= alpha) break;
                }
                return minScore;
            }
        }

        /**
         * Função heurística de avaliação do tabuleiro
         * Avalia a "qualidade" de um estado sem busca completa
         */
        function evaluateBoard() {
            const currentConnectivity = calculateConnectivity(currentPlayer);
            const opponentConnectivity = calculateConnectivity(getOpponent(currentPlayer));
            return currentConnectivity - opponentConnectivity;
        }

        /**
         * Calcula conectividade e posição estratégica de um jogador
         * @param {string} player - Jogador a avaliar
         */
        function calculateConnectivity(player) {
            let score = 0;
            
            for (let row = 0; row < BOARD_SIZE; row++) {
                for (let col = 0; col < BOARD_SIZE; col++) {
                    if (board[row][col] === player) {
                        // Conta vizinhos conectados (importante para formar caminhos)
                        const neighbors = getNeighbors(row, col);
                        score += neighbors.filter(n => board[n.row][n.col] === player).length * 10;
                        
                        // Bonificação por posição central (evita caminhos nas bordas)
                        if (player === 'blue') {
                            score += (BOARD_SIZE - Math.abs(col - BOARD_SIZE/2)) * 2;
                        } else {
                            score += (BOARD_SIZE - Math.abs(row - BOARD_SIZE/2)) * 2;
                        }

                        // Bonificação por progresso em direção ao objetivo
                        if (player === 'blue') {
                            score += col * 3; // Mais à direita = melhor
                        } else {
                            score += row * 3; // Mais abaixo = melhor
                        }
                    }
                }
            }
            
            return score;
        }

        // ============================================
        // FUNÇÕES AUXILIARES
        // ============================================

        /**
         * Retorna todos os movimentos válidos
         */
        function getValidMoves() {
            const moves = [];
            for (let row = 0; row < BOARD_SIZE; row++) {
                for (let col = 0; col < BOARD_SIZE; col++) {
                    if (!board[row][col]) {
                        moves.push({ row, col });
                    }
                }
            }
            return moves;
        }

        /**
         * Retorna vizinhos de uma célula hexagonal
         * No padrão HEX, cada célula tem 6 vizinhos
         */
        function getNeighbors(row, col) {
            const neighbors = [
                { row: row - 1, col: col },     // Superior
                { row: row - 1, col: col + 1 }, // Superior-direita
                { row: row, col: col - 1 },     // Esquerda
                { row: row, col: col + 1 },     // Direita
                { row: row + 1, col: col - 1 }, // Inferior-esquerda
                { row: row + 1, col: col }      // Inferior
            ];
            return neighbors.filter(n => 
                n.row >= 0 && n.row < BOARD_SIZE && 
                n.col >= 0 && n.col < BOARD_SIZE
            );
        }

        /**
         * Verifica se um jogador venceu
         */
        function checkWinner(player) {
            if (player === 'blue') {
                return hasPathLeftToRight(player);
            } else {
                return hasPathTopToBottom(player);
            }
        }

        /**
         * Verifica se existe caminho da esquerda para direita (jogador azul)
         */
        function hasPathLeftToRight(player) {
            const visited = Array(BOARD_SIZE).fill(null).map(() => Array(BOARD_SIZE).fill(false));
            
            // Tenta iniciar DFS de cada célula da borda esquerda
            for (let row = 0; row < BOARD_SIZE; row++) {
                if (board[row][0] === player && dfsLR(row, 0, player, visited)) {
                    return true;
                }
            }
            return false;
        }

        /**
         * DFS (busca em profundidade) para encontrar caminho esquerda-direita
         */
        function dfsLR(row, col, player, visited) {
            if (col === BOARD_SIZE - 1) return true; // Chegou à borda direita
            visited[row][col] = true;

            const neighbors = getNeighbors(row, col);
            for (const n of neighbors) {
                if (!visited[n.row][n.col] && board[n.row][n.col] === player) {
                    if (dfsLR(n.row, n.col, player, visited)) {
                        return true;
                    }
                }
            }
            return false;
        }

        /**
         * Verifica se existe caminho de cima para baixo (jogador vermelho)
         */
        function hasPathTopToBottom(player) {
            const visited = Array(BOARD_SIZE).fill(null).map(() => Array(BOARD_SIZE).fill(false));
            
            // Tenta iniciar DFS de cada célula da borda superior
            for (let col = 0; col < BOARD_SIZE; col++) {
                if (board[0][col] === player && dfsTB(0, col, player, visited)) {
                    return true;
                }
            }
            return false;
        }

        /**
         * DFS (busca em profundidade) para encontrar caminho cima-baixo
         */
        function dfsTB(row, col, player, visited) {
            if (row === BOARD_SIZE - 1) return true; // Chegou à borda inferior
            visited[row][col] = true;

            const neighbors = getNeighbors(row, col);
            for (const n of neighbors) {
                if (!visited[n.row][n.col] && board[n.row][n.col] === player) {
                    if (dfsTB(n.row, n.col, player, visited)) {
                        return true;
                    }
                }
            }
            return false;
        }

        /**
         * Retorna o vencedor atual (ou null se não houver)
         */
        function getWinner() {
            if (hasPathLeftToRight('blue')) return 'blue';
            if (hasPathTopToBottom('red')) return 'red';
            return null;
        }

        /**
         * Retorna o oponente de um jogador
         */
        function getOpponent(player) {
            return player === 'blue' ? 'red' : 'blue';
        }

        // ============================================
        // INTERFACE E CONTROLES
        // ============================================

        /**
         * Mostra modal de vitória
         */
        function showWinner(winner) {
            const winnerText = winner === 'blue' ? '🎉 Jogador Azul Venceu!' : '🎉 Jogador Vermelho Venceu!';
            document.getElementById('winnerText').textContent = winnerText;
            document.getElementById('winnerStats').textContent = 
                `Jogadas: ${moveCount} | Nós Explorados: ${nodesExplored}`;
            document.getElementById('winnerModal').style.display = 'flex';
        }

        /**
         * Fecha modal e inicia novo jogo
         */
        function closeModal() {
            document.getElementById('winnerModal').style.display = 'none';
            startGame();
        }

        /**
         * Atualiza informações de status na interface
         */
        function updateStatus() {
            document.getElementById('currentPlayer').textContent = 
                currentPlayer === 'blue' ? 'Azul' : 'Vermelho';
            document.getElementById('moveCount').textContent = moveCount;
        }

        /**
         * Inicia um novo jogo
         */
        function startGame() {
            initBoard();
            // Se modo IA vs IA, inicia automaticamente
            if (gameMode === 'cvc') {
                setTimeout(aiMove, 1000);
            }
        }

        /**
         * Reinicia o jogo atual
         */
        function resetGame() {
            startGame();
        }

        // ============================================
        // INICIALIZAÇÃO
        // ============================================

        /**
         * Inicializa o jogo quando a página carrega
         */
        window.onload = () => {
            initBoard();
            console.log('🎮 HEX Game inicializado - Tabuleiro 7x7');
            console.log('🧠 Algoritmos disponíveis: Minimax e Alfa-Beta');
        };