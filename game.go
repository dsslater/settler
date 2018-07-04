package main

import (
	"fmt"
)

type Game struct {
	Id       string
	Password string
	Players  map[string]Player
	Height   int
	Width    int
}

func (g Game) getPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		players = append(players, player)
	}
	return players
}

func (g Game) getReadyPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		if player.Ready {
			players = append(players, player)
		}
	}
	return players
}

func (g Game) getNumReadyPlayers() int {
	return len(g.getReadyPlayers())
}

func (g Game) getCells() []Cell {
	var cells []Cell
	
	getCellsText := fmt.Sprintf("SELECT * FROM %s;", g.Id)
	getCellsStmt, err := db.Prepare(getCellsText)
	if err != nil {
		fmt.Print("Preparing getCells statement failed: ", err)
		return cells
	}
	defer getCellsStmt.Close()
	rows, err := getCellsStmt.Query()
	if err != nil {
		fmt.Print("Query failed on getCellsStmt call: ", err)
		return cells
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on deleteAllTablesStmt call: ", err)
		return cells
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell)
		if err != nil {
			fmt.Print("SQL scan failed for getCells: ", err)
			return cells
		}
		cells = append(cells, cell)
	}
	return cells
}