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
	Event int		  `json:"event"`
	Data  interface{} `json:"data"`
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

var ActiveRooms map[string]Room
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
			fmt.Print(err)
			return
		}

		if message.Event == "createGame" {
			createGame(conn, message.Payload)
		} else if message.Event == "joinGame" {
			joinGame(conn, message.Payload)
		} else if message.Event == "iAmReady" {
			playerReady(conn, message.Payload)
		} else if message.Event == "moveArmies" {
			moveArmies(conn, message.Payload)
		}
	}
}


func createGame(conn *websocket.Conn, data interface{}) (Player, Room, error){
	message, ok := data.(CreateMessage)
	password := message.Password
	height := message.Height
	width := message.Width

	player := Player{
		Id: GenerateRandomId(),
		Conn: conn,
	}
	players := make(map[string]Player)
	players[player.Id] = player
	room := Room{
		Id: GenerateRandomId(),
		Password: password,
		Players:  players,
		Height:   height,
		Width:    width
	}
	err := CreateGameTable(room.Id, height, width)
	if err != nil {
		return player, room, err
	}
	ActiveRooms[room.Id] = room
	return player, room, nil
}


func joinGame(conn *websocket.Conn, data interface{}) (Player, Room, error){
	message, ok := data.(JoinMessage)
	room := message.Room
	password := message.Password

	player := Player{
		Id: GenerateRandomId(),
		Conn: conn,
	}

	room, ok := ActiveRooms[roomId]
	if !ok {
		return player, room, &errorString{"Room not found."}
	}

	room.Players[player.Id] = player
	return player, room, nil
}


func playerReady(conn *websocket.Conn, data interface{}) {
	// TODO
}


func moveArmies(conn *websocket.Conn, data interface{}) {
	// TODO
}


func SendRoomToClient(conn *websocket.Conn, player Player, room Room, messageType int) {
	wrapper := make(map[string]interface{})
	wrapper["room_id"] = room.Id
	wrapper["player_id"] = player.Id
	wrapper["num_players"] = len(room.Players)
	numReadyPlayers := 0
	for _, player := range room.Players {
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
	creationStmtText := fmt.Sprintf("CREATE TABLE %s (row int, col int, value int, owner varchar(255)) ENGINE=InnoDB DEFAULT CHARSET=utf8 AUTO_INCREMENT=1 ROW_FORMAT=DYNAMIC;", id)
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

	insertionStmtText := fmt.Sprintf("INSERT INTO %s (row, col, value, owner) VALUES ", id)
	var vals = []interface{}{}
	for r := 0; r < height; r++{
		for c := 0; c < width; c++{
			insertionStmtText += "(?, ?, ?, ?),"
			vals = append(vals, r, c, 0, "")
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


func main() {
	ActiveRooms = make(map[string]Room)
	rand.Seed(time.Now().UnixNano())
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
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/game", GameLoop)
	for {
		if err := http.ListenAndServe(":80", nil); err != nil {
			fmt.Print("Catastrophic error serving", err)
			fmt.Print("Restarting server.")
		}
	}
}



