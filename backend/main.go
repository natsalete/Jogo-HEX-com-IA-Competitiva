package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"time"
)

// ============================================
// ESTRUTURAS DE DADOS
// ============================================

type Cell struct {
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Color string `json:"color,omitempty"`
}

type Board struct {
	Size  int        `json:"size"`
	Cells [][]string `json:"cells"`
}

type AIRequest struct {
	Board     [][]string `json:"board"`
	Player    string     `json:"player"`
	Algorithm string     `json:"algorithm"`
	Depth     int        `json:"depth"`
}

type AIResponse struct {
	Move          Cell   `json:"move"`
	NodesExplored int    `json:"nodesExplored"`
	TimeMs        int64  `json:"timeMs"`
	Score         int    `json:"score"`
}

type GameState struct {
	Board          [][]string
	Size           int
	NodesExplored  int
	CurrentPlayer  string
	OpponentPlayer string
}

type ScoredMove struct {
	Row   int
	Col   int
	Score int
}

// Cache para distâncias (otimização)
var distanceCache map[string]int

func init() {
	distanceCache = make(map[string]int)
}

// ============================================
// SERVIDOR HTTP
// ============================================

func main() {
	http.HandleFunc("/api/ai-move", corsMiddleware(handleAIMove))
	http.HandleFunc("/api/validate-board", corsMiddleware(handleValidateBoard))
	http.HandleFunc("/api/check-winner", corsMiddleware(handleCheckWinner))
	http.HandleFunc("/", corsMiddleware(serveIndex))

	fmt.Println("🎮 Servidor HEX Game INTELIGENTE rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// ============================================
// HANDLERS DA API
// ============================================

func handleAIMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	gameState := &GameState{
		Board:          req.Board,
		Size:           len(req.Board),
		NodesExplored:  0,
		CurrentPlayer:  req.Player,
		OpponentPlayer: getOpponent(req.Player),
	}

	var bestMove ScoredMove
	if req.Algorithm == "alphabeta" {
		bestMove = findBestMoveAlphaBeta(gameState, req.Depth)
	} else {
		bestMove = findBestMoveMinimax(gameState, req.Depth)
	}

	elapsed := time.Since(startTime).Milliseconds()

	response := AIResponse{
		Move: Cell{
			Row: bestMove.Row,
			Col: bestMove.Col,
		},
		NodesExplored: gameState.NodesExplored,
		TimeMs:        elapsed,
		Score:         bestMove.Score,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleValidateBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var board Board
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	valid := validateBoard(board.Cells)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"valid": valid})
}

func handleCheckWinner(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var board Board
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	winner := checkWinner(board.Cells)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"winner": winner})
}

// ============================================
// ALGORITMO MINIMAX COM ALFA-BETA
// ============================================

func findBestMoveAlphaBeta(state *GameState, depth int) ScoredMove {
	bestScore := math.MinInt32
	var bestMove ScoredMove
	alpha := math.MinInt32
	beta := math.MaxInt32

	moves := getSmartMoves(state, 20) // Mais movimentos para melhor escolha

	for _, move := range moves {
		state.Board[move.Row][move.Col] = state.CurrentPlayer
		score := alphaBeta(state, depth-1, alpha, beta, false)
		state.Board[move.Row][move.Col] = ""

		if score > bestScore {
			bestScore = score
			bestMove = ScoredMove{Row: move.Row, Col: move.Col, Score: score}
		}
		if score > alpha {
			alpha = score
		}
		if beta <= alpha {
			break
		}
	}

	return bestMove
}

func alphaBeta(state *GameState, depth, alpha, beta int, isMaximizing bool) int {
	state.NodesExplored++

	winner := getWinner(state.Board)
	if winner == state.CurrentPlayer {
		return 100000 + depth*100
	}
	if winner == state.OpponentPlayer {
		return -100000 - depth*100
	}
	if depth == 0 {
		return evaluateBoardSmart(state)
	}

	moves := getSmartMoves(state, 15)
	if len(moves) == 0 {
		return 0
	}

	var player string
	if isMaximizing {
		player = state.CurrentPlayer
	} else {
		player = state.OpponentPlayer
	}

	if isMaximizing {
		maxScore := math.MinInt32
		for _, move := range moves {
			state.Board[move.Row][move.Col] = player
			score := alphaBeta(state, depth-1, alpha, beta, false)
			state.Board[move.Row][move.Col] = ""

			if score > maxScore {
				maxScore = score
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
		return maxScore
	} else {
		minScore := math.MaxInt32
		for _, move := range moves {
			state.Board[move.Row][move.Col] = player
			score := alphaBeta(state, depth-1, alpha, beta, true)
			state.Board[move.Row][move.Col] = ""

			if score < minScore {
				minScore = score
			}
			if score < beta {
				beta = score
			}
			if beta <= alpha {
				break
			}
		}
		return minScore
	}
}

func findBestMoveMinimax(state *GameState, depth int) ScoredMove {
	bestScore := math.MinInt32
	var bestMove ScoredMove

	moves := getSmartMoves(state, 20)

	for _, move := range moves {
		state.Board[move.Row][move.Col] = state.CurrentPlayer
		score := minimax(state, depth-1, false)
		state.Board[move.Row][move.Col] = ""

		if score > bestScore {
			bestScore = score
			bestMove = ScoredMove{Row: move.Row, Col: move.Col, Score: score}
		}
	}

	return bestMove
}

func minimax(state *GameState, depth int, isMaximizing bool) int {
	state.NodesExplored++

	winner := getWinner(state.Board)
	if winner == state.CurrentPlayer {
		return 100000 + depth*100
	}
	if winner == state.OpponentPlayer {
		return -100000 - depth*100
	}
	if depth == 0 {
		return evaluateBoardSmart(state)
	}

	moves := getSmartMoves(state, 15)
	if len(moves) == 0 {
		return 0
	}

	var player string
	if isMaximizing {
		player = state.CurrentPlayer
	} else {
		player = state.OpponentPlayer
	}

	if isMaximizing {
		maxScore := math.MinInt32
		for _, move := range moves {
			state.Board[move.Row][move.Col] = player
			score := minimax(state, depth-1, false)
			state.Board[move.Row][move.Col] = ""
			if score > maxScore {
				maxScore = score
			}
		}
		return maxScore
	} else {
		minScore := math.MaxInt32
		for _, move := range moves {
			state.Board[move.Row][move.Col] = player
			score := minimax(state, depth-1, true)
			state.Board[move.Row][move.Col] = ""
			if score < minScore {
				minScore = score
			}
		}
		return minScore
	}
}

// ============================================
// HEURÍSTICA INTELIGENTE - CONCEITOS DO HEX
// ============================================

func evaluateBoardSmart(state *GameState) int {
	myScore := evaluatePlayerPosition(state, state.CurrentPlayer)
	oppScore := evaluatePlayerPosition(state, state.OpponentPlayer)
	return myScore - oppScore
}

func evaluatePlayerPosition(state *GameState, player string) int {
	score := 0
	
	// 1. DISTÂNCIA MÍNIMA ATÉ O OBJETIVO (peso altíssimo)
	minDist := calculateMinDistance(state.Board, player, state.Size)
	score += (1000 - minDist*100) // Quanto menor distância, melhor
	
	// 2. COMPONENTES CONECTADOS (força da posição)
	components := countConnectedComponents(state.Board, player, state.Size)
	score -= components * 200 // Menos componentes = mais conectado = melhor
	
	// 3. MAIOR CADEIA CONECTADA
	longestChain := findLongestChain(state.Board, player, state.Size)
	score += longestChain * 150
	
	// 4. CONTROLE DO CENTRO
	centerControl := evaluateCenterControl(state.Board, player, state.Size)
	score += centerControl * 50
	
	// 5. PONTES VIRTUAIS (conexões de 2 espaços)
	bridges := countVirtualBridges(state.Board, player, state.Size)
	score += bridges * 80
	
	return score
}

// Calcula distância mínima até completar o objetivo
func calculateMinDistance(board [][]string, player string, size int) int {
	minDist := math.MaxInt32
	
	if player == "blue" {
		// Azul: menor distância da direita até a esquerda
		for row := 0; row < size; row++ {
			for col := 0; col < size; col++ {
				if board[row][col] == player {
					dist := bfsMinDistance(board, row, col, player, size, true)
					if dist < minDist {
						minDist = dist
					}
				}
			}
		}
	} else {
		// Vermelho: menor distância de baixo até cima
		for row := 0; row < size; row++ {
			for col := 0; col < size; col++ {
				if board[row][col] == player {
					dist := bfsMinDistance(board, row, col, player, size, false)
					if dist < minDist {
						minDist = dist
					}
				}
			}
		}
	}
	
	if minDist == math.MaxInt32 {
		return size * 2 // Sem peças no tabuleiro
	}
	return minDist
}

// BFS para encontrar menor distância até o objetivo
func bfsMinDistance(board [][]string, startRow, startCol int, player string, size int, isBlue bool) int {
	visited := make([][]bool, size)
	for i := range visited {
		visited[i] = make([]bool, size)
	}
	
	queue := []struct{ row, col, dist int }{{startRow, startCol, 0}}
	visited[startRow][startCol] = true
	
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		
		// Chegou ao objetivo?
		if isBlue && curr.col == size-1 {
			return curr.dist
		}
		if !isBlue && curr.row == size-1 {
			return curr.dist
		}
		
		// Explora vizinhos
		neighbors := getNeighbors(curr.row, curr.col, size)
		for _, n := range neighbors {
			if !visited[n.Row][n.Col] {
				visited[n.Row][n.Col] = true
				// Conta apenas células vazias ou do jogador
				if board[n.Row][n.Col] == player {
					queue = append(queue, struct{ row, col, dist int }{n.Row, n.Col, curr.dist})
				} else if board[n.Row][n.Col] == "" {
					queue = append(queue, struct{ row, col, dist int }{n.Row, n.Col, curr.dist + 1})
				}
			}
		}
	}
	
	return math.MaxInt32
}

// Conta componentes desconectados
func countConnectedComponents(board [][]string, player string, size int) int {
	visited := make([][]bool, size)
	for i := range visited {
		visited[i] = make([]bool, size)
	}
	
	components := 0
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if board[row][col] == player && !visited[row][col] {
				dfsMarkComponent(board, visited, row, col, player, size)
				components++
			}
		}
	}
	return components
}

func dfsMarkComponent(board [][]string, visited [][]bool, row, col int, player string, size int) {
	visited[row][col] = true
	neighbors := getNeighbors(row, col, size)
	for _, n := range neighbors {
		if !visited[n.Row][n.Col] && board[n.Row][n.Col] == player {
			dfsMarkComponent(board, visited, n.Row, n.Col, player, size)
		}
	}
}

// Encontra maior cadeia conectada
func findLongestChain(board [][]string, player string, size int) int {
	visited := make([][]bool, size)
	for i := range visited {
		visited[i] = make([]bool, size)
	}
	
	maxChain := 0
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if board[row][col] == player && !visited[row][col] {
				chainSize := dfsCountChain(board, visited, row, col, player, size)
				if chainSize > maxChain {
					maxChain = chainSize
				}
			}
		}
	}
	return maxChain
}

func dfsCountChain(board [][]string, visited [][]bool, row, col int, player string, size int) int {
	visited[row][col] = true
	count := 1
	neighbors := getNeighbors(row, col, size)
	for _, n := range neighbors {
		if !visited[n.Row][n.Col] && board[n.Row][n.Col] == player {
			count += dfsCountChain(board, visited, n.Row, n.Col, player, size)
		}
	}
	return count
}

// Avalia controle do centro
func evaluateCenterControl(board [][]string, player string, size int) int {
	score := 0
	center := size / 2
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if board[row][col] == player {
				distToCenter := abs(row-center) + abs(col-center)
				score += (size - distToCenter)
			}
		}
	}
	return score
}

// Conta pontes virtuais (duas peças com célula vazia entre elas)
func countVirtualBridges(board [][]string, player string, size int) int {
	bridges := 0
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if board[row][col] == player {
				neighbors := getNeighbors(row, col, size)
				for _, n := range neighbors {
					if board[n.Row][n.Col] == "" {
						// Verifica se há outra peça do outro lado
						nextNeighbors := getNeighbors(n.Row, n.Col, size)
						for _, nn := range nextNeighbors {
							if board[nn.Row][nn.Col] == player && (nn.Row != row || nn.Col != col) {
								bridges++
							}
						}
					}
				}
			}
		}
	}
	return bridges / 2 // Divide por 2 pois conta cada ponte duas vezes
}

// ============================================
// SELEÇÃO INTELIGENTE DE MOVIMENTOS
// ============================================

func getSmartMoves(state *GameState, maxMoves int) []Cell {
	allMoves := getValidMoves(state.Board)
	
	if len(allMoves) <= maxMoves {
		return allMoves
	}

	scoredMoves := make([]ScoredMove, len(allMoves))
	for i, move := range allMoves {
		state.Board[move.Row][move.Col] = state.CurrentPlayer
		score := evaluateMoveQuality(state, move.Row, move.Col)
		state.Board[move.Row][move.Col] = ""
		scoredMoves[i] = ScoredMove{Row: move.Row, Col: move.Col, Score: score}
	}

	sort.Slice(scoredMoves, func(i, j int) bool {
		return scoredMoves[i].Score > scoredMoves[j].Score
	})

	result := make([]Cell, maxMoves)
	for i := 0; i < maxMoves && i < len(scoredMoves); i++ {
		result[i] = Cell{Row: scoredMoves[i].Row, Col: scoredMoves[i].Col}
	}
	return result
}

func evaluateMoveQuality(state *GameState, row, col int) int {
	score := 0
	player := state.CurrentPlayer
	opponent := state.OpponentPlayer
	size := state.Size
	
	// 1. Conectividade com peças existentes
	neighbors := getNeighbors(row, col, size)
	myNeighbors := 0
	oppNeighbors := 0
	for _, n := range neighbors {
		if state.Board[n.Row][n.Col] == player {
			myNeighbors++
		} else if state.Board[n.Row][n.Col] == opponent {
			oppNeighbors++
		}
	}
	score += myNeighbors * 300
	score += oppNeighbors * 200 // Bloquear também é importante
	
	// 2. Progresso em direção ao objetivo
	if player == "blue" {
		score += col * 50
	} else {
		score += row * 50
	}
	
	// 3. Proximidade ao centro
	center := size / 2
	distToCenter := abs(row-center) + abs(col-center)
	score += (size - distToCenter) * 20
	
	return score
}

// ============================================
// FUNÇÕES AUXILIARES
// ============================================

func getNeighbors(row, col, size int) []Cell {
	directions := [][]int{
		{-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0},
	}

	var neighbors []Cell
	for _, dir := range directions {
		newRow := row + dir[0]
		newCol := col + dir[1]
		if newRow >= 0 && newRow < size && newCol >= 0 && newCol < size {
			neighbors = append(neighbors, Cell{Row: newRow, Col: newCol})
		}
	}
	return neighbors
}

func getValidMoves(board [][]string) []Cell {
	var moves []Cell
	size := len(board)
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if board[row][col] == "" {
				moves = append(moves, Cell{Row: row, Col: col})
			}
		}
	}
	return moves
}

// ============================================
// DETECÇÃO DE VITÓRIA
// ============================================

func checkWinner(board [][]string) string {
	if hasPathLeftToRight(board, "blue") {
		return "blue"
	}
	if hasPathTopToBottom(board, "red") {
		return "red"
	}
	return ""
}

func hasPathLeftToRight(board [][]string, player string) bool {
	size := len(board)
	visited := make([][]bool, size)
	for i := range visited {
		visited[i] = make([]bool, size)
	}

	for row := 0; row < size; row++ {
		if board[row][0] == player {
			if dfsLR(board, visited, row, 0, player, size) {
				return true
			}
		}
	}
	return false
}

func dfsLR(board [][]string, visited [][]bool, row, col int, player string, size int) bool {
	if col == size-1 {
		return true
	}
	visited[row][col] = true

	neighbors := getNeighbors(row, col, size)
	for _, n := range neighbors {
		if !visited[n.Row][n.Col] && board[n.Row][n.Col] == player {
			if dfsLR(board, visited, n.Row, n.Col, player, size) {
				return true
			}
		}
	}
	return false
}

func hasPathTopToBottom(board [][]string, player string) bool {
	size := len(board)
	visited := make([][]bool, size)
	for i := range visited {
		visited[i] = make([]bool, size)
	}

	for col := 0; col < size; col++ {
		if board[0][col] == player {
			if dfsTB(board, visited, 0, col, player, size) {
				return true
			}
		}
	}
	return false
}

func dfsTB(board [][]string, visited [][]bool, row, col int, player string, size int) bool {
	if row == size-1 {
		return true
	}
	visited[row][col] = true

	neighbors := getNeighbors(row, col, size)
	for _, n := range neighbors {
		if !visited[n.Row][n.Col] && board[n.Row][n.Col] == player {
			if dfsTB(board, visited, n.Row, n.Col, player, size) {
				return true
			}
		}
	}
	return false
}

func getWinner(board [][]string) string {
	return checkWinner(board)
}

func validateBoard(board [][]string) bool {
	if len(board) == 0 {
		return false
	}
	size := len(board)
	for _, row := range board {
		if len(row) != size {
			return false
		}
	}
	return true
}

func getOpponent(player string) string {
	if player == "blue" {
		return "red"
	}
	return "blue"
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}