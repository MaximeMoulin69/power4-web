package main

import (
	"html/template"
	"net/http"
	"strconv"
)

type Game struct {
	Board         [][]int
	CurrentPlayer int
	Winner        int
	IsDraw        bool
	Player1Name   string
	Player2Name   string
	Rows          int
	Cols          int
	TurnCount     int
}

var game *Game

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/reset", resetHandler)
	
	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		player1 := r.FormValue("player1")
		player2 := r.FormValue("player2")
		difficulty := r.FormValue("difficulty")

		var rows, cols int
		switch difficulty {
		case "easy":
			rows, cols = 6, 7
		case "normal":
			rows, cols = 6, 9
		case "hard":
			rows, cols = 7, 8
		default:
			rows, cols = 6, 7
		}

		game = &Game{
			Board:         makeBoard(rows, cols),
			CurrentPlayer: 1,
			Winner:        0,
			IsDraw:        false,
			Player1Name:   player1,
			Player2Name:   player2,
			Rows:          rows,
			Cols:          cols,
			TurnCount:     0,
		}

		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/start.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	if game == nil {
		http.Redirect(w, r, "/start", http.StatusSeeOther)
		return
	}

	funcMap := template.FuncMap{
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
	}

	tmpl, err := template.New("game.html").Funcs(funcMap).ParseFiles("templates/game.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || game == nil {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	colStr := r.FormValue("column")
	col, err := strconv.Atoi(colStr)
	if err != nil || col < 0 || col >= game.Cols {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	if game.Winner != 0 || game.IsDraw {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	row := findRow(game.Board, col)
	if row == -1 {
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	game.Board[row][col] = game.CurrentPlayer
	game.TurnCount++

	if checkWin(game.Board, row, col, game.CurrentPlayer) {
		game.Winner = game.CurrentPlayer
	} else if checkDraw(game.Board) {
		game.IsDraw = true
	} else {
		game.CurrentPlayer = 3 - game.CurrentPlayer
	}

	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	if game != nil {
		game.Board = makeBoard(game.Rows, game.Cols)
		game.CurrentPlayer = 1
		game.Winner = 0
		game.IsDraw = false
		game.TurnCount = 0
	}
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func makeBoard(rows, cols int) [][]int {
	board := make([][]int, rows)
	for i := range board {
		board[i] = make([]int, cols)
	}
	return board
}

func findRow(board [][]int, col int) int {
	for row := len(board) - 1; row >= 0; row-- {
		if board[row][col] == 0 {
			return row
		}
	}
	return -1
}

func checkWin(board [][]int, row, col, player int) bool {
	return checkDirection(board, row, col, player, 0, 1) ||
		checkDirection(board, row, col, player, 1, 0) ||
		checkDirection(board, row, col, player, 1, 1) ||
		checkDirection(board, row, col, player, 1, -1)
}

func checkDirection(board [][]int, row, col, player, dRow, dCol int) bool {
	count := 1
	count += countInDirection(board, row, col, player, dRow, dCol)
	count += countInDirection(board, row, col, player, -dRow, -dCol)
	return count >= 4
}

func countInDirection(board [][]int, row, col, player, dRow, dCol int) int {
	count := 0
	r, c := row+dRow, col+dCol
	for r >= 0 && r < len(board) && c >= 0 && c < len(board[0]) {
		if board[r][c] == player {
			count++
			r += dRow
			c += dCol
		} else {
			break
		}
	}
	return count
}

func checkDraw(board [][]int) bool {
	for _, row := range board {
		for _, cell := range row {
			if cell == 0 {
				return false
			}
		}
	}
	return true
}
