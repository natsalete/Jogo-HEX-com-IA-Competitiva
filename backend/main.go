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

// ============================================
// SERVIDOR HTTP
// ============================================

func main() {
	http.HandleFunc("/api/ai-move", corsMiddleware(handleAIMove))
	http.HandleFunc("/api/validate-board", corsMiddleware(handleValidateBoard))
	http.HandleFunc("/api/check-winner", corsMiddleware(handleCheckWinner))
	http.HandleFunc("/", corsMiddleware(serveIndex))

	fmt.Println("🎮 Servidor HEX Game OTIMIZADO rodando em http://localhost:8080")
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
// ALGORITMO MINIMAX - OTIMIZADO
// ============================================

func findBestMoveMinimax(state *GameState, depth int) ScoredMove {
	bestScore := math.MinInt32
	var bestMove ScoredMove

	// OTIMIZAÇÃO 1: Limita movimentos candidatos
	moves := getBestCandidateMoves(state, 15)

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
		return 10000 + depth
	}
	if winner == state.OpponentPlayer {
		return -10000 - depth
	}
	if depth == 0 {
		return evaluateBoardFast(state)
	}

	// OTIMIZAÇÃO 2: Limita movimentos em profundidade
	moves := getBestCandidateMoves(state, 12)
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
// ALGORITMO ALFA-BETA - OTIMIZADO
// ============================================

func findBestMoveAlphaBeta(state *GameState, depth int) ScoredMove {
	bestScore := math.MinInt32
	var bestMove ScoredMove
	alpha := math.MinInt32
	beta := math.MaxInt32

	// OTIMIZAÇÃO 3: Pega apenas os melhores candidatos
	moves := getBestCandidateMoves(state, 15)

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
		return 10000 + depth
	}
	if winner == state.OpponentPlayer {
		return -10000 - depth
	}
	if depth == 0 {
		return evaluateBoardFast(state)
	}

	moves := getBestCandidateMoves(state, 12)
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

// ============================================
// OTIMIZAÇÕES - SELEÇÃO INTELIGENTE DE MOVIMENTOS
// ============================================

// getBestCandidateMoves retorna apenas os movimentos mais promissores
// REDUZ drasticamente o fator de ramificação
func getBestCandidateMoves(state *GameState, maxMoves int) []Cell {
	allMoves := getValidMoves(state.Board)
	
	// Se há poucos movimentos, retorna todos
	if len(allMoves) <= maxMoves {
		return allMoves
	}

	// Avalia cada movimento UMA ÚNICA VEZ
	scoredMoves := make([]ScoredMove, len(allMoves))
	for i, move := range allMoves {
		state.Board[move.Row][move.Col] = state.CurrentPlayer
		score := quickEval(state, move.Row, move.Col) // Avaliação rápida
		state.Board[move.Row][move.Col] = ""
		scoredMoves[i] = ScoredMove{Row: move.Row, Col: move.Col, Score: score}
	}

	// Ordena e retorna os melhores
	sort.Slice(scoredMoves, func(i, j int) bool {
		return scoredMoves[i].Score > scoredMoves[j].Score
	})

	result := make([]Cell, maxMoves)
	for i := 0; i < maxMoves && i < len(scoredMoves); i++ {
		result[i] = Cell{Row: scoredMoves[i].Row, Col: scoredMoves[i].Col}
	}
	return result
}

// quickEval: avaliação ULTRA-RÁPIDA focada apenas na jogada atual
func quickEval(state *GameState, row, col int) int {
	score := 0
	player := state.CurrentPlayer
	size := state.Size

	// 1. VIZINHOS CONECTADOS (peso alto)
	neighbors := getNeighbors(row, col, size)
	for _, n := range neighbors {
		if state.Board[n.Row][n.Col] == player {
			score += 50 // Alta pontuação por conexão
		}
	}

	// 2. PROGRESSO (peso médio)
	if player == "blue" {
		score += col * 10 // Avança para direita
	} else {
		score += row * 10 // Avança para baixo
	}

	// 3. POSIÇÃO CENTRAL (peso baixo)
	if player == "blue" {
		centerDist := abs(col - size/2)
		score += (size - centerDist) * 2
	} else {
		centerDist := abs(row - size/2)
		score += (size - centerDist) * 2
	}

	// 4. BLOQUEIA OPONENTE
	opponent := state.OpponentPlayer
	for _, n := range neighbors {
		if state.Board[n.Row][n.Col] == opponent {
			score += 30 // Bonifica bloqueio
		}
	}

	return score
}

// ============================================
// AVALIAÇÃO COMPLETA (mais cara, só no depth 0)
// ============================================

func evaluateBoardFast(state *GameState) int {
	currentScore := 0
	opponentScore := 0

	for row := 0; row < state.Size; row++ {
		for col := 0; col < state.Size; col++ {
			if state.Board[row][col] == state.CurrentPlayer {
				currentScore += quickEval(state, row, col)
			} else if state.Board[row][col] == state.OpponentPlayer {
				opponentScore += quickEval(state, row, col)
			}
		}
	}

	return currentScore - opponentScore
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