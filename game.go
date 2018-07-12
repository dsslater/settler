package main

import (
	"fmt"
)

type Game struct {
	Id       string
	Password string
	Players  map[string]*Player
	Height   int
	Width    int
	Started  bool
}

const COLORS = [...]string{"red", "green", "blue", "orange", "purple", "yellow", "grey", "pink"}

func (g *Game) getPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		players = append(players, *player)
	}
	return players
}

func (g *Game) getReadyPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		if player.Ready {
			players = append(players, *player)
		}
	}
	return players
}

func (g *Game) getNumReadyPlayers() int {
	return len(g.getReadyPlayers())
}

func (g *Game) GetCell(index [2]int) (Cell, error) {
	var cell Cell
	row := index[0]
	col := index[1]
	getCellText := fmt.Sprintf("SELECT * FROM %s WHERE row=%d AND col=%d;", g.Id, row, col)
	getCellStmt, err := db.Prepare(getCellText)
	if err != nil {
		fmt.Print("Preparing getCell statement failed: ", err, "\n")
		return cell, err
	}
	defer getCellStmt.Close()
	rows, err := getCellStmt.Query()
	if err != nil {
		fmt.Print("Query failed on getCellStmt call: ", err, "\n")
		return cell, err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on deleteAllTablesStmt call: ", err, "\n")
		return cell, err
	}

	for rows.Next() {
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			fmt.Print("SQL scan failed for getCells: ", err, "\n")
			return cell, err
		}
	}
	fmt.Print("CELL: ", cell, "\n")
	return cell, nil
}

func (g *Game) GetCells() []Cell {
	var cells []Cell
	
	getCellsText := fmt.Sprintf("SELECT * FROM %s;", g.Id)
	getCellsStmt, err := db.Prepare(getCellsText)
	if err != nil {
		fmt.Print("Preparing getCells statement failed: ", err, "\n")
		return cells
	}
	defer getCellsStmt.Close()
	rows, err := getCellsStmt.Query()
	if err != nil {
		fmt.Print("Query failed on getCellsStmt call: ", err, "\n")
		return cells
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on deleteAllTablesStmt call: ", err, "\n")
		return cells
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			fmt.Print("SQL scan failed for getCells: ", err, "\n")
			return cells
		}
		cells = append(cells, cell)
	}
	fmt.Print("CELLS: ", cells, "\n")
	return cells
}

func (g *Game) MarkCity(index [2]int, playerId string) {
	row := index[0]
	col := index[1]
	markCityText := fmt.Sprintf("UPDATE %s SET city=true, owner=%s WHERE row=%d AND col=%d;", g.Id, playerId, row, col)
	markCityStmt, err := db.Prepare(markCityText)
	if err != nil {
		fmt.Print("Preparing MarkCity statement failed: ", err, "\n")
		return
	}
	defer markCityStmt.Close()
	_, err = markCityStmt.Exec()
	if err != nil {
		fmt.Print("Query failed on MarkCity call: ", err, "\n")
		return
	}
}

func (g *Game) AssignColors() {
	i := 0
	for _, player := range g.Players {
		player.Color = COLORS[i]
		i++
	}
}