package main


import (
	"cloud.google.com/go/logging"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)


const (
	safeBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	numNPCCities = 10
	cityAmountBase = 40
	cityAmountRange = 10
	cityGrowthRatio = 5
	growthCycleTime = 4.0 * time.Second
)


var logInfo *log.Logger

var logError *log.Logger


type createmessage struct {
	Password string `json:"gamePass"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
}


type joinmessage struct {
	GameID   string `json:"gameId"`
	Password string `json:"gamePass"`
}


type moveArmiesmessage struct {
	StartRow int `json:"start_row"`
	EndRow   int `json:"end_row"`
	StartCol int `json:"start_col"`
	EndCol   int `json:"end_col"`
}


type message struct {
	Event string	  `json:"event"`
	Data  interface{} `json:"data"`
}


type gameInformation struct {
	GameID   string `json:"gameId"`
	ID         string `json:"id"`
	Dimensions [2]int `json:"dimensions"`
	Points     []Cell  `json:"points"`
	NumPlayers int    `json:"numPlayers"`
}


type playerInformation struct {
	Players      []Player `json:"players"`
	ReadyPlayers []Player `json:"readyPlayers"`
}


func checkOriginFunc (r *http.Request) bool {
	return true
}


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOriginFunc,
}


var activeGames map[string]*Game
var db *sql.DB

// GameLoop establishes a websocket connection and then handles all messages from the client.
func GameLoop(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logError.Println(err)
		return
	}
	var player *Player
	var game *Game
	// Main loop that drives entire game
	for {
		var message message
		if err := conn.ReadJSON(&message); err != nil {
			if game != nil {
				game.RemovePlayer(player)
			}
			return
		}

		if message.Event == "createGame" {
			game, player, err = createGame(conn, message.Data)
			if err != nil {
				logError.Println("Catastrophic failure creating game. Disconnecting.")
				return
			}
		} else if message.Event == "joinGame" {
			game, player, err = joinGame(conn, message.Data)
			if err != nil {
				logError.Println("Catastrophic failure joining game. Disconnecting.")
				return
			}
		} else if message.Event == "playerReady" {
			playerReady(conn, game, player)
		} else if message.Event == "moveArmies" {
			moveArmies(conn, game, player, message.Data)
		}
	}
}


func createGame(conn *websocket.Conn, data interface{}) (*Game, *Player, error){
	var game *Game
	var player *Player

	logInfo.Println("Creating game.\n")
	bytes, err := json.Marshal(data)
	if err != nil {
		logError.Println("Error with data in createGame:" + err.Error())
		_ := emit(conn, "unkownGameCreationError", nil)
		return game, player, err
	}
	var message createmessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		logError.Println("Unable to unmarshal data to createmessage:" + err.Error())
		_ := emit(conn, "unkownGameCreationError", nil)
		return game, player, err
	}
	password := message.Password
	height := message.Height
	width := message.Width

	player = CreatePlayer(conn)
	players := make(map[string]*Player)
	players[player.ID] = player
	game = &Game{
		ID: GenerateRandomID(),
		Password: password,
		Players:  players,
		Height:   height,
		Width:    width,
		Started:  false,
	}
	err = createGameTable(game.ID, height, width)
	if err != nil {
		_ := emit(conn, "unkownGameCreationError", nil)
		return game, player, err
	}
	addNPCCities(game)
	activeGames[game.ID] = game
	sendGameData(conn, player, game)
	sendPlayerData(game)
	return game, player, nil
}


func addNPCCities(game *Game) {
	for i := 0; i < numNPCCities; i++ {
		row := rand.Intn(game.Height)
		col := rand.Intn(game.Width)
		index := [2]int{row, col}
		amount := cityAmountBase + rand.Intn(cityAmountRange)
		game.MarkCity(index, "NPC", amount, "white")
	}
}


func joinGame(conn *websocket.Conn, data interface{}) (*Game, *Player, error){
	var game *Game
	var player *Player

	logInfo.Println("Joining game.")
	bytes, err := json.Marshal(data)
	if err != nil {
		logError.Println("Error with data in joinGame: ", err)
		_ := emit(conn, "unknownGameJoinError", nil)
		return game, player, err
	}
	var message joinmessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		logError.Println("Unable to unmarshal data to joinmessage: ", err)
		_ := emit(conn, "unknownGameJoinError", nil)
		return game, player, err
	}

	gameID := message.GameID
	password := message.Password

	player = CreatePlayer(conn)

	game, ok := activeGames[gameID]
	if !ok {
		logError.Println("Game not found.")
		_ := emit(conn, "gameNotFound", nil)
		return game, player, errors.New("Game not found")
	}
	if game.Started {
		logError.Println("Game has already started.")
		_ := emit(conn, "gameStarted", nil)
		return game, player, errors.New("Game has already started")
	}

	if password != game.Password {
		logError.Println("Wrong password!")
		_ := emit(conn, "wrongPassword", nil)
		return game, player, errors.New("Wrong password")
	}

	game.Players[player.ID] = player
	sendGameData(conn, player, game)
	sendPlayerData(game)
	return game, player, nil
}


func sendGameData(conn *websocket.Conn, player *Player, game *Game) {
	gameInformation := gameInformation{
		GameID: game.ID,
		ID: player.ID,
		Dimensions: [2]int{game.Height, game.Width},
		Points: game.GetCells(),
		NumPlayers: 0,
	}

	if err := emit(conn, "gameReady", gameInformation); err != nil {
		game.RemovePlayer(player)
	}
}


func sendPlayerData(game *Game) {
	playerInformation := playerInformation{
		Players: game.GetPlayers(),
		ReadyPlayers: game.GetReadyPlayers(),
	}

	emitToGame(game, "playerUpdate", playerInformation)
}


func emitToGame(game *Game, event string, data interface{}) {
	for _, player := range game.Players {
		if err := emit(player.Conn, event, data); err != nil {
			game.RemovePlayer(player)
		}
	}
}


func emit(conn *websocket.Conn, event string, data interface{}) error {
	// TODO: handle failure especially on websocket writing.
	wrapper := make(map[string]interface{})
	wrapper["event"] = event
	wrapper["data"] = data
	bytes, err := json.Marshal(wrapper)
	if err != nil {
		logError.Println("Failure the Marshal in emit: ", err)
		return err
	}
	if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		logError.Println("Failure writing to websocket in emit: ", err)
		return err
	}
	return nil
}


func setupGrowth(game *Game) {
	go func () {
		cycle := 1
		for {
			start := time.Now()
			if game.Finished {
				break
			}
			var cells []Cell
			var err error
			if cycle % cityGrowthRatio == 0 {
				cells, err = game.GrowAll()
			} else {
				cells, err = game.GrowCities()
			}
			if err != nil {
				return
			}
			emitToGame(game, "update", cells)
			time.Sleep(growthCycleTime - time.Since(start))
			cycle++
		}
	}()
}


func playerReady(conn *websocket.Conn, game *Game, player *Player) {
	logInfo.Println("Player ready.\n")

	player, ok := game.Players[player.ID]
	if !ok {
		logError.Println("Player not found.\n")
		return
	}

	player.Ready = true
	if len(game.GetPlayers()) == len(game.GetReadyPlayers()) {
		game.Started = true
		game.AssignColors()
		startGame(conn, game)
	} else {
		sendPlayerData(game)
	}
}


func startGame(conn *websocket.Conn, game *Game) {
	// add cities for each player marked as being owned by them
	playerCities := make(map[[2]int]bool)
	for _, player := range game.Players {
		var row int
		var col int
		var index [2]int
		for {
			row = rand.Intn(game.Height)
			col = rand.Intn(game.Width)
			index = [2]int{row, col}
			_, ok := playerCities[index]
			if !ok {
				cellRow := index[0]
				cellCol := index[1]
				cell, err := game.GetCell(cellRow, cellCol)
				if err != nil {
					logError.Println("Failure accessing cell at index: ", index, " with error: ", err, "\n")
					return
				}
				if !cell.City {
					break
				}
			}
		}
		playerCities[index] = true
		game.MarkCity(index, player.ID, 1, player.Color)
	}
	setupGrowth(game);
	sendPlayerCities(game, playerCities)
	sendPlayerData(game)
    emitToGame(game, "startGame", nil)
}


func sendPlayerCities(game *Game, playerCities map[[2]int]bool) {
	var cells []Cell
	for index := range playerCities {
		row := index[0]
		col := index[1]
		cell, err := game.GetCell(row, col)
		if err != nil {
			logError.Println("Failure accessing player cities cell at index: ", index, " with error: ", err, "\n")
			return
		}
		cells = append(cells, cell)
	}
	emitToGame(game, "update", cells)
}


func moveArmies(conn *websocket.Conn, game *Game, player *Player, data interface{}) {
	logInfo.Println("Move armies.")
	bytes, err := json.Marshal(data)
	if err != nil {
		logError.Println("Error with data in moveArmies: ", err)
		return
	}

	var message moveArmiesmessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		logError.Println("Unable to unmarshal data to joinmessage: ", err)
		return
	}

	var effectedCells []Cell
	startRow := message.StartRow
	endRow := message.EndRow
	startCol := message.StartCol
	endCol := message.EndCol

	if startRow == endRow {
		// horizontal drag
		target := endCol
		var begin int
		var end int
		if endCol > startCol {
			// right drag
			begin = startCol
			end = endCol - 1
		} else {
			// left drag
			begin = endCol + 1
			end = startCol
		}
		game.MoveHorizontal(player, startRow, begin, end, target)
		if startCol < endCol {
			effectedCells = game.GetEffectedCells(startRow, startCol, startRow, endCol)
		} else {
			effectedCells = game.GetEffectedCells(startRow, endCol, startRow, startCol)
		}
	} else {
		// vertical drag
		target := endRow
		var begin int
		var end int
		if endRow > startRow {
			// down drag
			begin = startRow
			end = endRow - 1
		} else {
			// up drag
			begin = endRow + 1
			end = startRow
		}
		game.MoveVertical(player, startCol, begin, end, target)
		if startRow < endRow {
			effectedCells = game.GetEffectedCells(startRow, startCol, endRow, startCol)
		} else {
			effectedCells = game.GetEffectedCells(endRow, startCol, startRow, startCol)
		}
	}
	emitToGame(game, "update", effectedCells)
}


// GenerateRandomID creates a string of random characters to be used as a unique id
func GenerateRandomID() string {
	b := make([]byte, 32)
	for i := range b {
	b[i] = safeBytes[rand.Intn(len(safeBytes))]
	}
	return string(b)
}


func createGameTable(id string, height int, width int) error {
	creationStmtText := fmt.Sprintf("CREATE TABLE %s (row int, col int, city bool, amount int, owner varchar(255), color varchar(100)) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 ROW_FORMAT=DYNAMIC;", id)
	createStmt, err := db.Prepare(creationStmtText)
	if err != nil {
		logError.Println("Preparing creation statement failed for createGameTable: ", err)
		return err
	}
	defer createStmt.Close()

	 _, err = createStmt.Exec()
	if err != nil {
		logError.Println("createGameTable creation SQL command failed: ", err)
		return err
	}

	insertionStmtText := fmt.Sprintf("INSERT INTO %s (row, col, city, amount, owner, color) VALUES ", id)
	var vals = []interface{}{}
	for r := 0; r < height; r++{
		for c := 0; c < width; c++{
			insertionStmtText += "(?, ?, ?, ?, ?, ?),"
			vals = append(vals, r, c, false, 0, "NPC", "white")
		}
	}
	
	insertionStmtText = insertionStmtText[0:len(insertionStmtText)-1]
	insertionStmt, err:= db.Prepare(insertionStmtText)
	if err != nil {
		logError.Println("Preparing insertion statement failed for createGameTable: ", err)
		return err
	}
	defer insertionStmt.Close()

	_, err = insertionStmt.Exec(vals...)
	if err != nil {
		logError.Println("Executing statement failed for insertion in createGameTable: ", err)
		return err
	}

	indexText := fmt.Sprintf("CREATE INDEX row_col ON %s (row, col);", id)
	indexStmt, err := db.Prepare(indexText)
	if err != nil {
		logError.Println("Failed to build index error: ", err)
		return err
	}
	defer indexStmt.Close()

	_, err = indexStmt.Exec()
	if err != nil {
		logError.Println("Executing statement failed for indexing in createGameTable: ", err)
		return err
	}

	return nil
}


func deleteOldTables() {
	deleteAllTablesText := "SELECT concat('DROP TABLE IF EXISTS `', table_name, '`;') FROM information_schema.tables WHERE table_schema = 'settler';"
	deleteAllTablesStmt, err := db.Prepare(deleteAllTablesText)
	if err != nil {
		logError.Println("Preparing deleteAllTablesStmt failed: ", err)
		return
	}

	rows, err := deleteAllTablesStmt.Query()
	if err != nil {
		logError.Println("Query failed on deleteAllTablesStmt call: ", err)
		return
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		logError.Println("Rows had an error on deleteAllTablesStmt call: ", err)
		return
	}

	for rows.Next() {
		var deleteCmd string
		err := rows.Scan(&deleteCmd)
		if err != nil {
			logError.Println("SQL scan failed for deleteCmd: ", err)
			return
		}

		deleteStmt, err := db.Prepare(deleteCmd)

		_, err = deleteStmt.Exec()
		if err != nil {
			logError.Println("deleteStmt SQL command failed: ", err)
			return
		}
	}
}


func main() {
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "settler-208704"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defer client.Close()

	// Sets the name of the log to write to.
	logName := "settler-log"

	logInfo = client.Logger(logName).StandardLogger(logging.Info)
	logError = client.Logger(logName).StandardLogger(logging.Error)

	logInfo.Println("Logging Initialized")
	activeGames = make(map[string]*Game)
	rand.Seed(time.Now().UnixNano())
	// Connect to SQL DB
	data, err := ioutil.ReadFile("./database_login")
	if err != nil {
		logError.Println("Err reading database login file")
		panic(err)
	}
	databaseLogin := strings.TrimSpace(string(data))
	db, err = sql.Open("mysql", databaseLogin)
	if err != nil {
		logError.Println("Err connecting to the MYSQL database")
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logError.Println("Err with ping to db")
		panic(err.Error())
	}
	logInfo.Println("Connected to SQL")
	// Clean up old tables
	deleteOldTables()
	// Set up http server
	http.Handle("/", http.FileServer(http.Dir("./go-public")))
	http.HandleFunc("/game", GameLoop)
	for {
		if err := http.ListenAndServe(":80", nil); err != nil {
			logError.Println("Catastrophic error serving", err)
			logError.Println("Restarting server.")
		}
	}
}
