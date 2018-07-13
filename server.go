package main


import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
)


const (
	SAFE_BYTES = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	CONNECT_MESSAGE = 0
	READY_MESSAGE = 1
	NUM_NPC_CITIES = 10
	CITY_AMOUNT_BASE = 40
	CITY_AMOUNT_RANGE = 10
	CITY_GROWTH_RATIO = 5
	GROWTH_CYCLE_TIME = 4 * time.Millisecond
)


var SIZES = map[int][2]int {
	0: [2]int{10, 10},
	1: [2]int{15, 15},
	2: [2]int{20, 20},
	3: [2]int{30, 30},
}


type CreateMessage struct {
	Password string `json:"gamePass"`
	Height   int    `json:"height"`
	Width    int    `json:"width"`
}


type JoinMessage struct {
	GameId   string `json:"gameId"`
	Password string `json:"gamePass"`
}


type ReadyMessage struct {
	GameId   string `json:"gameId"`
	PlayerId string `json:"playerId"`
}


type Message struct {
	Event string	  `json:"event"`
	Data  interface{} `json:"data"`
}


type GameInformation struct {
	GameId   string `json:"gameId"`
	Id         string `json:"id"`
	Dimensions [2]int `json:"dimensions"`
	Points     []Cell  `json:"points"`
	NumPlayers int    `json:"numPlayers"`
}


type PlayerInformation struct {
	Players      []Player `json:"players"`
	ReadyPlayers []Player `json:"readyPlayers"`
}


func CheckOriginFunc (r *http.Request) bool {
	return true
}


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOriginFunc,
}


type errorString struct {
	s string
}


func (e *errorString) Error() string {
	return e.s
}


var ActiveGames map[string]*Game
var db *sql.DB


func GameLoop(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print(err)
		return
	}
	var player *Player
	var game *Game
	// Main loop that drives entire game
	for {
		var message Message
		if err := conn.ReadJSON(&message); err != nil {
			// TODO: implement disconnect code
			fmt.Print("Error reading JSON from socket: ", err, "\n")
			return
		}

		if message.Event == "createGame" {
			game, player, err = createGame(conn, message.Data)
			if err != nil {
				fmt.Print("Catastrophic failure creating game. Disconnecting.")
				return
			}
		} else if message.Event == "joinGame" {
			game, player, err = joinGame(conn, message.Data)
			if err != nil {
				fmt.Print("Catastrophic failure joining game. Disconnecting.")
				return
			}
		} else if message.Event == "playerReady" {
			playerReady(conn, game, player)
		} else if message.Event == "moveArmies" {
			moveArmies(conn, message.Data)
		}
	}
}


func createGame(conn *websocket.Conn, data interface{}) (*Game, *Player, error){
	var game *Game
	var player *Player

	fmt.Print("Creating game.\n")
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Print("Error with data in createGame:" + err.Error())
		return game, player, err
	}
	var message CreateMessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		fmt.Print("Unable to unmarshal data to CreateMessage:" + err.Error())
		return game, player, err
	}
	password := message.Password
	height := message.Height
	width := message.Width

	player = createPlayer(conn)
	players := make(map[string]*Player)
	players[player.Id] = player
	game = &Game{
		Id: GenerateRandomId(),
		Password: password,
		Players:  players,
		Height:   height,
		Width:    width,
		Started:  false,
	}
	err = CreateGameTable(game.Id, height, width)
	if err != nil {
		return game, player, err
	}
	addNPCCities(game)
	ActiveGames[game.Id] = game
	sendGameData(conn, player, game)
	sendPlayerData(conn, game)
	setupGrowth(game);
	return game, player, nil
}


func addNPCCities(game *Game) {
	for i := 0; i < NUM_NPC_CITIES; i++ {
		row := rand.Intn(game.Height)
		col := rand.Intn(game.Width)
		index := [2]int{row, col}
		amount := CITY_AMOUNT_BASE + rand.Intn(CITY_AMOUNT_RANGE)
		game.MarkCity(index, "", amount, "white")
	}
}


func joinGame(conn *websocket.Conn, data interface{}) (*Game, *Player, error){
	var game *Game
	var player *Player

	fmt.Print("Joining game.\n")
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Print("Error with data in joinGame:" + err.Error())
		return game, player, err
	}
	var message JoinMessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		fmt.Print("Unable to unmarshal data to JoinMessage:" + err.Error())
		return game, player, err
	}

	gameId := message.GameId
	password := message.Password

	player = createPlayer(conn)

	game, ok := ActiveGames[gameId]
	if !ok {
		fmt.Print("Game not found.\n")
		return game, player, nil // TODO: create custom error
	}

	if password != game.Password {
		fmt.Print("Wrong password!\n")
		return game, player, nil // TODO: create custom error
	}

	game.Players[player.Id] = player
	sendGameData(conn, player, game)
	sendPlayerData(conn, game)
	return game, player, nil
}


func sendGameData(conn *websocket.Conn, player *Player, game *Game) {
	gameInformation := GameInformation{
		GameId: game.Id,
		Id: player.Id,
		Dimensions: [2]int{game.Height, game.Width},
		Points: game.GetCells(),
		NumPlayers: 0,
	};

	emit(conn, "gameReady", gameInformation);
}


func sendPlayerData(conn *websocket.Conn, game *Game) {
	playerInformation := PlayerInformation{
		Players: game.getPlayers(),
		ReadyPlayers: game.getReadyPlayers(),
	};

	emitToGame(game, "playerUpdate", playerInformation);
}


func emitToGame(game *Game, event string, data interface{}) {
	for _, player := range game.Players {
		emit(player.Conn, event, data)
	}
}


func emit(conn *websocket.Conn, event string, data interface{}) {
	wrapper := make(map[string]interface{})
	wrapper["event"] = event
	wrapper["data"] = data
	bytes, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Print("Failure the Marshal in emit: ", err)
		return
	}
	fmt.Print("Emitting: [", event, "] ", string(bytes), "\n")
	if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		fmt.Print("Failure writing to websocket in emit: ", err)
		return
	}
}


func setupGrowth(game *Game) {
	go func () {
		cycle := 1
		for {
			fmt.Print("Growth Cycle!")
			start := time.Now()
			if game.Finished {
				break
			}
			var cells []Cell
			var err error
			if cycle % CITY_GROWTH_RATIO == 0 {
				cells, err = game.GrowAll()
			} else {
				cells, err = game.GrowCities()
			}
			if err != nil {
				return
			}
			emitToGame(game, "update", cells)
			time.Sleep(GROWTH_CYCLE_TIME - time.Since(start))
			cycle++
		}
	}()
}


func playerReady(conn *websocket.Conn, game *Game, player *Player) {
	fmt.Print("Player ready.\n")

	player, ok := game.Players[player.Id]
	if !ok {
		fmt.Print("Player not found.\n")
		return
	}

	player.Ready = true
	if len(game.getPlayers()) == len(game.getReadyPlayers()) {
		// TODO: mark game as started
		game.Started = true
		game.AssignColors()
		startGame(conn, game)
	} else {
		sendPlayerData(conn, game)
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
				cell, err := game.GetCell(index)
				if err != nil {
					fmt.Print("Failure accessing cell at index: ", index, " with error: ", err, "\n")
					return
				}
				if !cell.City {
					break
				}
			}
		}
		playerCities[index] = true
		game.MarkCity(index, player.Id, 1, player.Color)
	}
	sendPlayerCities(game, playerCities)
	sendPlayerData(conn, game)
    emitToGame(game, "startGame", nil)
}


func sendPlayerCities(game *Game, playerCities map[[2]int]bool) {
	var cells []Cell
	for index, _ := range playerCities {
		cell, err := game.GetCell(index)
		if err != nil {
			fmt.Print("Failure accessing player cities cell at index: ", index, " with error: ", err, "\n")
			return
		}
		cells = append(cells, cell)
	}
	emitToGame(game, "update", cells)
}


func moveArmies(conn *websocket.Conn, data interface{}) {
	// TODO
}


func GenerateRandomId() string {
	b := make([]byte, 64)
	for i := range b {
	b[i] = SAFE_BYTES[rand.Intn(len(SAFE_BYTES))]
	}
	return string(b)
}


func CreateGameTable(id string, height int, width int) error {
	creationStmtText := fmt.Sprintf("CREATE TABLE %s (row int, col int, city bool, amount int, owner varchar(255), color varchar(100)) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 ROW_FORMAT=DYNAMIC;", id)
	createStmt, err := db.Prepare(creationStmtText)
	if err != nil {
		fmt.Print("Preparing creation statement failed for CreateGameTable: ", err)
		return err
	}
	defer createStmt.Close()

	 _, err = createStmt.Exec()
	if err != nil {
		fmt.Print("CreateGameTable creation SQL command failed: ", err)
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
		fmt.Print("Preparing insertion statement failed for CreateGameTable: ", err)
		return err
	}
	defer insertionStmt.Close()

	_, err = insertionStmt.Exec(vals...)
	if err != nil {
		fmt.Print("Executing statement failed for insertion in CreateGameTable: ", err)
		return err
	}

	return nil
}


func deleteOldTables() {
	deleteAllTablesText := "SELECT concat('DROP TABLE IF EXISTS `', table_name, '`;') FROM information_schema.tables WHERE table_schema = 'settler';"
	deleteAllTablesStmt, err := db.Prepare(deleteAllTablesText)
	if err != nil {
		fmt.Print("Preparing deleteAllTablesStmt failed: ", err)
		return
	}

	rows, err := deleteAllTablesStmt.Query()
	if err != nil {
		fmt.Print("Query failed on deleteAllTablesStmt call: ", err)
		return
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Print("Rows had an error on deleteAllTablesStmt call: ", err)
		return
	}

	for rows.Next() {
		var deleteCmd string
		err := rows.Scan(&deleteCmd)
		if err != nil {
			fmt.Print("SQL scan failed for deleteCmd: ", err)
			return
		}

		deleteStmt, err := db.Prepare(deleteCmd)

		_, err = deleteStmt.Exec()
		if err != nil {
			fmt.Print("deleteStmt SQL command failed: ", err)
			return
		}
	}
}


func main() {
	ActiveGames = make(map[string]*Game)
	rand.Seed(time.Now().UnixNano())
	// Connect to SQL DB
	data, err := ioutil.ReadFile("./database_login")
	if err != nil {
		fmt.Print("Err reading database login file")
		panic(err)
	}
	databaseLogin := strings.TrimSpace(string(data))
	db, err = sql.Open("mysql", databaseLogin)
	if err != nil {
		fmt.Print("Err connecting to the MYSQL database")
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Print("Err with ping to db")
		panic(err.Error())
	}
	fmt.Print("Connected to SQL\n")
	// Clean up old tables
	deleteOldTables()
	// Set up http server
	http.Handle("/", http.FileServer(http.Dir("./go-public")))
	http.HandleFunc("/game", GameLoop)
	for {
		if err := http.ListenAndServe(":80", nil); err != nil {
			fmt.Print("Catastrophic error serving", err)
			fmt.Print("Restarting server.")
		}
	}
}
