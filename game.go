package main


import (
	"fmt"
)


// Game controls all logic for a single game.
type Game struct {
	ID       string
	Password string
	Players  map[string]*Player
	Height   int
	Width    int
	Started  bool
	Finished bool
}

var colors = [...]string{"red", "green", "blue", "orange", "purple", "yellow", "grey", "pink"}


// GetPlayers returns a slice of all players.
func (g *Game) GetPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		players = append(players, *player)
	}
	return players
}


// GetReadyPlayers returns a slice of all players who have marked themselves as ready.
func (g *Game) GetReadyPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		if player.Ready {
			players = append(players, *player)
		}
	}
	return players
}


// GetCell returns a Cell object populated with the db information for the row, col provided.
func (g *Game) GetCell(row int, col int) (Cell, error) {
	var cell Cell
	getCellText := fmt.Sprintf("SELECT * FROM %s WHERE row= ? AND col= ?;", g.ID)
	getCellStmt, err := db.Prepare(getCellText)
	if err != nil {
		logError.Println("Preparing getCell statement failed: ", err, "\n")
		return cell, err
	}
	defer getCellStmt.Close()
	rows, err := getCellStmt.Query(row, col)
	if err != nil {
		logError.Println("Query failed on getCellStmt call: ", err, "\n")
		return cell, err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on deleteAllTablesStmt call: ", err, "\n")
		return cell, err
	}

	for rows.Next() {
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			logError.Println("SQL scan failed for getCells: ", err, "\n")
			return cell, err
		}
	}

	return cell, nil
}


// GetCells return a slice of all Cells.
func (g *Game) GetCells() []Cell {
	var cells []Cell
	
	getCellsText := fmt.Sprintf("SELECT * FROM %s;", g.ID)
	getCellsStmt, err := db.Prepare(getCellsText)
	if err != nil {
		logError.Println("Preparing getCells statement failed: ", err, "\n")
		return cells
	}
	defer getCellsStmt.Close()
	rows, err := getCellsStmt.Query()
	if err != nil {
		logError.Println("Query failed on getCellsStmt call: ", err, "\n")
		return cells
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on deleteAllTablesStmt call: ", err, "\n")
		return cells
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			logError.Println("SQL scan failed for getCells: ", err, "\n")
			return cells
		}
		cells = append(cells, cell)
	}

	return cells
}


// MarkCity marks a specified row and col as a city and assign its owner, amount, cand color.
func (g *Game) MarkCity(index [2]int, playerID string, amount int, color string) {
	row := index[0]
	col := index[1]
	markCityText := fmt.Sprintf("UPDATE %s SET city= ?, owner= ?, amount= ?, color= ? WHERE row=? AND col=?;", g.ID)
	markCityStmt, err := db.Prepare(markCityText)
	if err != nil {
		logError.Println("Preparing MarkCity statement failed: ", err, "\n")
		return
	}
	defer markCityStmt.Close()
	_, err = markCityStmt.Exec(true, playerID, amount, color, row, col)
	if err != nil {
		logError.Println("Exec failed on MarkCity call: ", err, "\n")
		return
	}
}


// AssignColors is called to assign all players their respective colors.
func (g *Game) AssignColors() {
	i := 0
	for _, player := range g.Players {
		player.Color = colors[i]
		i++
	}
}


// getGrowthChangedCells, provided with a query for the type of growth, returns a slice of the cells effected.
func getGrowthChangedCells(changedCellsQuery string) ([]Cell, error) {
	var cells []Cell

	changedCellsStmt, err := db.Prepare(changedCellsQuery)
	if err != nil {
		logError.Println("Preparing getGrowthChangedCells statement failed: ", err, "\n")
		return cells, err
	}
	defer changedCellsStmt.Close()

	rows, err := changedCellsStmt.Query()
	if err != nil {
		logError.Println("Query failed on getGrowthChangedCellsStmt call: ", err, "\n")
		return cells, err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on getGrowthChangedCells call: ", err, "\n")
		return cells, err
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			logError.Println("SQL scan failed for getGrowthChangedCells: ", err, "\n")
			return cells, err
		}
		cells = append(cells, cell)
	}
	return cells, nil
}


// GrowAll increments all player owned cells and returns a slice of the effected cells.c
func (g *Game) GrowAll() ([]Cell, error) {
	var cells []Cell
	growAllText := fmt.Sprintf("UPDATE %s SET amount = amount + 1 WHERE owner != 'NPC';", g.ID)
	growAllStmt, err  := db.Prepare(growAllText)
	if err != nil {
		logError.Println("Preparing GrowAll statement failed: ", err, "\n")
		return cells, err
	}
	defer growAllStmt.Close()

	_, err = growAllStmt.Exec()
	if err != nil {
		logError.Println("Exec failed on growAllStmt call: ", err, "\n")
		return cells, err
	}

	changedCellsText := fmt.Sprintf("SELECT * FROM %s WHERE owner != 'NPC';", g.ID)
	return getGrowthChangedCells(changedCellsText)
}


// GrowCities increments all player owned cities amount by one and returns a slice of the effected cells.
func (g *Game) GrowCities() ([]Cell, error) {
	var cells []Cell
	growCitiesText := fmt.Sprintf("UPDATE %s SET amount = amount + 1 WHERE owner != 'NPC' AND city = true;", g.ID)
	growCitiesStmt, err  := db.Prepare(growCitiesText)
	if err != nil {
		logError.Println("Preparing GrowAll statement failed: ", err, "\n")
		return cells, err
	}
	defer growCitiesStmt.Close()

	_, err = growCitiesStmt.Exec()
	if err != nil {
		logError.Println("Exec failed on growAllStmt call: ", err, "\n")
		return cells, err
	}

	changedCellsText := fmt.Sprintf("SELECT * FROM %s WHERE owner != 'NPC' AND city = true;", g.ID)
	return getGrowthChangedCells(changedCellsText)
}


// saveCell, provided with a cell, saves the information to the db.
func (g *Game) saveCell(cell Cell) {
	saveCellText := fmt.Sprintf("UPDATE %s SET owner = ?, color = ?, amount = ? WHERE row = ? AND col = ?;", g.ID)
	saveCellStmt, err := db.Prepare(saveCellText)
	if err != nil {
		logError.Println("Preparing saveCellStmt statement failed: ", err, "\n")
		return
	}
	defer saveCellStmt.Close()

	_, err = saveCellStmt.Exec(cell.Owner, cell.Color, cell.Amount, cell.Row, cell.Col)
	if err != nil {
		logError.Println("Exec failed on saveCell call: ", err, "\n")
		return
	}	
}


// addArmies handles army placement on either a cell owned by the player or an opponent.
func (g *Game) addArmies(player *Player, targetRow int, targetCol int, amount int) {
	cell, err := g.GetCell(targetRow, targetCol)
	if err != nil {
		logError.Println("Couldn't get target cell. Need to rollback: ", err, "\n")
		return
	}

	if (cell.Owner == player.ID) {
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
			cell.Owner = player.ID;
		}
    }
    g.saveCell(cell)
}


/* move handles a players army movement. Checks that the player owns all cells in the move, ensures 
   that there are sufficient armies to move, sets all cells in the move to value 1 and finally
   takes all the armies moved and calls addArmies on the target. */
func (g *Game) move(player *Player, beginRow int, beginCol int, endRow int, endCol int, targetRow int, targetCol int) {
	// Check that the player has exclusive control
	checkControlText := fmt.Sprintf("SELECT DISTINCT owner FROM %s WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.ID)
	checkControlStmt, err := db.Prepare(checkControlText)
	if err != nil {
		logError.Println("Preparing CheckControl statement failed: ", err, "\n")
		return
	}
	defer checkControlStmt.Close()

	rows, err := checkControlStmt.Query(beginRow, endRow, beginCol, endCol)
	if err != nil {
		logError.Println("Query failed on checkControlStmt call: ", err, "\n")
		return
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on checkControlStmt call: ", err, "\n")
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
			logError.Println("SQL scan failed for checkControlStmt: ", err, "\n")
			return
		}
		if owner != player.ID {
			return
		}
		numPlayers++
	}

	// Sum armies in the move
	sumMoveText := fmt.Sprintf("SELECT SUM(amount) FROM %s WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.ID)
	sumMoveStmt, err := db.Prepare(sumMoveText)
	if err != nil {
		logError.Println("Preparing sumMove statement failed: ", err, "\n")
		return
	}
	defer sumMoveStmt.Close()

	rows, err = sumMoveStmt.Query(beginRow, endRow, beginCol, endCol)
	if err != nil {
		logError.Println("Query failed on sumMoveStmt call: ", err, "\n")
		return
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on sumMoveStmt call: ", err, "\n")
		return
	}
	var sum int
	for rows.Next() {
		err := rows.Scan(&sum)
		if err != nil {
			logError.Println("SQL scan failed for sumMoveStmt: ", err, "\n")
			return
		}
	}
	// We need to subtract the 1 army we leave on each cell
	costOfOnes := endRow + endCol - beginRow - beginCol + 1
	armiesToMoveToTarget := sum - costOfOnes
	if armiesToMoveToTarget < 1 {
		return
	}
	// Set all cells in the move to 1
	setCellsToOneText := fmt.Sprintf("UPDATE %s SET amount = 1 WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.ID)
	setCellsToOneStmt, err := db.Prepare(setCellsToOneText)
	if err != nil {
		logError.Println("Preparing setCellsToOne statement failed: ", err, "\n")
		return
	}
	defer setCellsToOneStmt.Close()

	_, err = setCellsToOneStmt.Exec(beginRow, endRow, beginCol, endCol)
	if err != nil {
		logError.Println("Exec failed on setCellsToOne call: ", err, "\n")
		return
	}

	// Apply to target
	g.addArmies(player, targetRow, targetCol, armiesToMoveToTarget)
}


// MoveHorizontal handles horizontal moves.
func (g *Game) MoveHorizontal(player *Player, row int, begin int, end int, target int) {
	g.move(player, row, begin, row, end, row, target)
}


// MoveVertical handles vertical moves.
func (g *Game) MoveVertical(player *Player, col int, begin int, end int, target int) {
	g.move(player, begin, col, end, col, target, col)
}

// GetEffectedCells returns a slice of cells effected by a move.
func (g *Game) GetEffectedCells(beginRow int, beginCol int, endRow int, endCol int) []Cell {
	var cells []Cell
	getEffectedText := fmt.Sprintf("SELECT * FROM %s WHERE row >= ? AND row <= ? AND col >= ? AND col <= ?;", g.ID)
	getEffectedStmt, err := db.Prepare(getEffectedText)
	if err != nil {
		logError.Println("Preparing CheckControl statement failed: ", err, "\n")
		return cells
	}
	defer getEffectedStmt.Close()

	rows, err := getEffectedStmt.Query(beginRow, endRow, beginCol, endCol)
	if err != nil {
		logError.Println("Query failed on getEffectedStmt call: ", err, "\n")
		return cells
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on getEffectedStmt call: ", err, "\n")
		return cells
	}

	for rows.Next() {
		var cell Cell
		err := rows.Scan(&cell.Row, &cell.Col, &cell.City, &cell.Amount, &cell.Owner, &cell.Color)
		if err != nil {
			logError.Println("SQL scan failed for getEffectedStmt: ", err, "\n")
			return cells
		}
		cells = append(cells, cell)
	}
	return cells
}