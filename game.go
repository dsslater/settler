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
	Finished bool
}

var COLORS = [...]string{"red", "green", "blue", "orange", "purple", "yellow", "grey", "pink"}


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

func (g *Game) GetCell(row int, col int) (Cell, error) {
	var cell Cell
	getCellText := fmt.Sprintf("SELECT * FROM %s WHERE row= ? AND col= ?;", g.Id)
	getCellStmt, err := db.Prepare(getCellText)
	if err != nil {
		fmt.Print("Preparing getCell statement failed: ", err, "\n")
		return cell, err
	}
	defer getCellStmt.Close()
	rows, err := getCellStmt.Query(row, col)
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

func (g *Game) MarkCity(index [2]int, playerId string, amount int, color string) {
	row := index[0]
	col := index[1]
	markCityText := fmt.Sprintf("UPDATE %s SET city= ?, owner= ?, amount= ?, color= ? WHERE row=? AND col=?;", g.Id)
	markCityStmt, err := db.Prepare(markCityText)
	if err != nil {
		fmt.Print("Preparing MarkCity statement failed: ", err, "\n")
		return
	}
	defer markCityStmt.Close()
	_, err = markCityStmt.Exec(true, playerId, amount, color, row, col)
	if err != nil {
		fmt.Print("Exec failed on MarkCity call: ", err, "\n")
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


func getChangedCells(changedCellsText string) ([]Cell, error) {
	var cells []Cell

	changedCellsStmt, err := db.Prepare(changedCellsText)
	if err != nil {
		fmt.Print("Preparing getChangedCells statement failed: ", err, "\n")
		return cells, err
	}
	defer changedCellsStmt.Close()

	rows, err := changedCellsStmt.Query()
	if err != nil {
		fmt.Print("Query failed on changedCellsStmt call: ", err, "\n")
		return cells, err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on getChangedCells call: ", err, "\n")
		return cells, err
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			fmt.Print("SQL scan failed for getChangedCells: ", err, "\n")
			return cells, err
		}
		cells = append(cells, cell)
	}
	return cells, nil
}


func (g *Game) GrowAll() ([]Cell, error) {
	var cells []Cell
	growAllText := fmt.Sprintf("UPDATE %s SET amount = amount + 1 WHERE owner != 'NPC';", g.Id)
	growAllStmt, err  := db.Prepare(growAllText)
	if err != nil {
		fmt.Print("Preparing GrowAll statement failed: ", err, "\n")
		return cells, err
	}
	defer growAllStmt.Close()

	_, err = growAllStmt.Exec()
	if err != nil {
		fmt.Print("Exec failed on growAllStmt call: ", err, "\n")
		return cells, err
	}

	changedCellsText := fmt.Sprintf("SELECT * FROM %s WHERE owner != 'NPC';", g.Id)
	return getChangedCells(changedCellsText)
}


func (g *Game) GrowCities() ([]Cell, error) {
	var cells []Cell
	growCitiesText := fmt.Sprintf("UPDATE %s SET amount = amount + 1 WHERE owner != 'NPC' AND city = true;", g.Id)
	growCitiesStmt, err  := db.Prepare(growCitiesText)
	if err != nil {
		fmt.Print("Preparing GrowAll statement failed: ", err, "\n")
		return cells, err
	}
	defer growCitiesStmt.Close()

	_, err = growCitiesStmt.Exec()
	if err != nil {
		fmt.Print("Exec failed on growAllStmt call: ", err, "\n")
		return cells, err
	}

	changedCellsText := fmt.Sprintf("SELECT * FROM %s WHERE owner != 'NPC' AND city = true;", g.Id)
	return getChangedCells(changedCellsText)
}


func (g *Game) SaveCell(cell Cell) {
	fmt.Print("Saving Cell\n")
	saveCellText := fmt.Sprintf("UPDATE %s SET owner = ?, color = ?, amount = ? WHERE row = ? AND col = ?;", g.Id)
	saveCellStmt, err := db.Prepare(saveCellText)
	if err != nil {
		fmt.Print("Preparing saveCellStmt statement failed: ", err, "\n")
	}
	defer saveCellStmt.Close()

	_, err = saveCellStmt.Exec(cell.Owner, cell.Color, cell.Amount, cell.Row, cell.Col)
	if err != nil {
		fmt.Print("Exec failed on saveCell call: ", err, "\n")
		return
	}	
}



func (g *Game) AddArmies(player *Player, targetRow int, targetCol int, amount int) {
	fmt.Print("Adding armies\n")
	cell, err := g.GetCell(targetRow, targetCol)
	if err != nil {
		fmt.Print("Couldn't get target cell. Need to rollback: ", err, "\n")
		return
	}

	if (cell.Owner == player.Id) {
      // If player already owns this point add the amount to the point.
      cell.Amount += amount;
    } else {
      /* If the player does not own the point subtract the amoint from the 
         point and, if the amount is greater than the amount in the point, 
         change ownership. */
		remainder := cell.Amount - amount;
		if remainder >= 0 {
			cell.Amount = remainder;
		} else {
			cell.Amount = remainder * -1;
			cell.Color = player.Color;
			cell.Owner = player.Id;
		}
    }
    g.SaveCell(cell)
}


func (g *Game) Move(player *Player, beginRow int, beginCol int, endRow int, endCol int, targetRow int, targetCol int) {
	fmt.Print("MOVE: player: ", player.Id, "    beginRow: ", beginRow, "    beginCol: ", beginCol, "    endRow: ", endRow, "    endCol: ", endCol, "    targetRow: ", targetRow, "    targetCol: ", targetCol, "\n")
	// Check that the player has exclusive control
	fmt.Print("CheckControl\n")
	checkControlText := fmt.Sprintf("SELECT DISTINCT owner FROM %s WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.Id)
	checkControlStmt, err := db.Prepare(checkControlText)
	if err != nil {
		fmt.Print("Preparing CheckControl statement failed: ", err, "\n")
		return
	}
	defer checkControlStmt.Close()

	rows, err := checkControlStmt.Query(beginRow, endRow, beginCol, endCol)
	if err != nil {
		fmt.Print("Query failed on checkControlStmt call: ", err, "\n")
		return
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on checkControlStmt call: ", err, "\n")
		return
	}
	numPlayers := 0
	for rows.Next() {
		if numPlayers > 0 {
			return
		}
		var owner string
		err := rows.Scan(&owner)
		if err != nil {
			fmt.Print("SQL scan failed for checkControlStmt: ", err, "\n")
			return
		}
		if owner != player.Id {
			return
		}
		numPlayers++
	}

	// Sum armies in the move
	fmt.Print("Summing Move\n")
	sumMoveText := fmt.Sprintf("SELECT SUM(amount) FROM %s WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.Id)
	sumMoveStmt, err := db.Prepare(sumMoveText)
	if err != nil {
		fmt.Print("Preparing sumMove statement failed: ", err, "\n")
		return
	}
	defer sumMoveStmt.Close()

	rows, err = sumMoveStmt.Query(beginRow, endRow, beginCol, endCol)
	if err != nil {
		fmt.Print("Query failed on sumMoveStmt call: ", err, "\n")
		return
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on sumMoveStmt call: ", err, "\n")
		return
	}
	var sum int
	for rows.Next() {
		err := rows.Scan(&sum)
		if err != nil {
			fmt.Print("SQL scan failed for sumMoveStmt: ", err, "\n")
			return
		}
	}
	// We ned to subtract the 1 army we leave on each cell
	costOfOnes := endRow + endCol - beginRow - beginCol + 1
	armiesToMoveToTarget := sum - costOfOnes
	if armiesToMoveToTarget < 1 {
		return
	}
	fmt.Print("There are suffcient armies to move: ", armiesToMoveToTarget, "\n")
	// Set all cells in the move to 1
	setCellsToOneText := fmt.Sprintf("UPDATE %s SET amount = 1 WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.Id)
	setCellsToOneStmt, err := db.Prepare(setCellsToOneText)
	if err != nil {
		fmt.Print("Preparing setCellsToOne statement failed: ", err, "\n")
		return
	}
	defer setCellsToOneStmt.Close()

	_, err = setCellsToOneStmt.Exec(beginRow, endRow, beginCol, endCol)
	if err != nil {
		fmt.Print("Exec failed on setCellsToOne call: ", err, "\n")
		return
	}

	// Apply to target
	g.AddArmies(player, targetRow, targetCol, armiesToMoveToTarget)
}


func (g *Game) MoveHorizontal(player *Player, row int, begin int, end int, target int) {
	g.Move(player, row, begin, row, end, row, target)
}


func (g *Game) MoveVertical(player *Player, col int, begin int, end int, target int) {
	g.Move(player, begin, col, end, col, target, col)
}

func (g *Game) getEffectedCells(beginRow int, beginCol int, endRow int, endCol int) []Cell {
	var cells []Cell
	getEffectedText := fmt.Sprintf("SELECT * FROM %s WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.Id)
	getEffectedStmt, err := db.Prepare(getEffectedText)
	if err != nil {
		fmt.Print("Preparing CheckControl statement failed: ", err, "\n")
		return cells
	}
	defer getEffectedStmt.Close()

	rows, err := getEffectedStmt.Query(beginRow, endRow, beginCol, endCol)
	if err != nil {
		fmt.Print("Query failed on getEffectedStmt call: ", err, "\n")
		return cells
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on getEffectedStmt call: ", err, "\n")
		return cells
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			fmt.Print("SQL scan failed for getEffectedStmt: ", err, "\n")
			return cells
		}
		cells = append(cells, cell)
	}
	return cells
}