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
	GameName string `json:"gameName"`
	Password string `json:"gamePass"`
}


type Message struct {
	Event string	  `json:"event"`
	Data  interface{} `json:"data"`
}


type GameInformation struct {
	Room       string `json:"room"`
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


var ActiveGames map[string]Game
var db *sql.DB


func GameLoop(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print(err)
		return
	}
	// Main loop that drives entire game
	for {
		var message Message
		if err := conn.ReadJSON(&message); err != nil {
			// TODO: implement disconnect code
			fmt.Print("Error reading JSON from socket: ", err, "\n")
			return
		}

		if message.Event == "createGame" {
			createGame(conn, message.Data)
		} else if message.Event == "joinGame" {
			joinGame(conn, message.Data)
		} else if message.Event == "iAmReady" {
			playerReady(conn, message.Data)
		} else if message.Event == "moveArmies" {
			moveArmies(conn, message.Data)
		}
	}
}


func createGame(conn *websocket.Conn, data interface{}) {
	fmt.Print("Creating game.\n")
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Print("Error with data in createGame:" + err.Error())
		return
	}
	var message CreateMessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		fmt.Print("Unable to unmarshal data to CreateMessage:" + err.Error())
		return
	}
	password := message.Password
	height := message.Height
	width := message.Width

	player := createPlayer(conn)
	players := make(map[string]Player)
	players[player.Id] = player
	game := Game{
		Id: GenerateRandomId(),
		Password: password,
		Players:  players,
		Height:   height,
		Width:    width,
	}
	err = CreateGameTable(game.Id, height, width)
	if err != nil {
		return
	}
	ActiveGames[game.Id] = game
	sendPlayerData(conn, player, game)
}


func joinGame(conn *websocket.Conn, data interface{}) {
	fmt.Print("Joining game.\n")
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Print("Error with data in joinGame:" + err.Error())
		return
	}
	var message JoinMessage
	err = json.Unmarshal(bytes, &message)
	if err != nil {
		fmt.Print("Unable to unmarshal data to JoinMessage:" + err.Error())
		return
	}
	gameId := message.GameName
	password := message.Password

	player := createPlayer(conn)
	// TODO: not currently working
	game, ok := ActiveGames[gameId]
	if !ok {
		fmt.Print("Game not found.")
		return
	}

	if password != game.Password {
		fmt.Print("Wrong password!")
		return
	}

	game.Players[player.Id] = player
	sendPlayerData(conn, player, game)
}


func sendPlayerData(conn *websocket.Conn, player Player, game Game) {
	gameInformation := GameInformation{
		Room: game.Id,
		Id: player.Id,
		Dimensions: [2]int{game.Height, game.Width},
		Points: game.getCells(),
		NumPlayers: 0,
	};

	emit(conn, "gameReady", gameInformation);

	playerInformation := PlayerInformation{
		Players: game.getPlayers(),
		ReadyPlayers: game.getReadyPlayers(),
	};

	emitToGame(game, "playerUpdate", playerInformation);
	setupGrowth();
}


func emitToGame(game Game, event string, data interface{}) {
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
	fmt.Print("Emitting: ", string(bytes), "\n")
	if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		fmt.Print("Failure writing to websocket in emit: ", err)
		return
	}
}


func setupGrowth() {
	// TODO manage a go routine for growth
}


func playerReady(conn *websocket.Conn, data interface{}) {
	// TODO
}


func moveArmies(conn *websocket.Conn, data interface{}) {
	// TODO
}


func SendGameToClient(conn *websocket.Conn, player Player, game Game, messageType int) {
	wrapper := make(map[string]interface{})
	wrapper["room_id"] = game.Id
	wrapper["player_id"] = player.Id
	wrapper["num_players"] = len(game.Players)
	numReadyPlayers := 0
	for _, player := range game.Players {
		if player.Ready {
			numReadyPlayers++
		}
	}
	wrapper["num_ready_player"] = numReadyPlayers
	wrapper["type"] = messageType
	data, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Print(err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Print(err)
		return
	}
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
			vals = append(vals, r, c, false, 0, "", "")
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
	ActiveGames = make(map[string]Game)
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
