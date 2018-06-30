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
	okBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	CONNECT_RESPONSE = 0
)

var SIZES = map[int][]int {
	0: [10, 10],
	1: [15, 15],
	2: [20, 20],
	3: [30, 30],
}

type ConnectMessage struct {
	Room     string `json:"room"`
	Password string `json:"password"`
	Size     int    `json:"size"`
}

func CheckOriginFunc (r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOriginFunc,
}

var ActiveGames map[string]Room
var db *sql.DB

func Connect(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print(err)
		return
	}
	var message ConnectMessage
	if err = conn.ReadJSON(&message); err != nil {
		fmt.Print(err)
		return
	}
	if message.Room != "" {
		fmt.Print("Adding to room: " + message.Room + "\n")
	} else {
		fmt.Print("Creating room!\n")
		room, err := CreateRoom(conn, message.Size)
		if err != nil {
			return
		}
		SendRoomIdToClient(room, conn)
	}
}

func SendRoomIdToClient(room Room, conn *websocket.Conn) {
	wrapper := make(map[string]interface{})
	wrapper["room_id"] = room.Id
	wrapper["type"] = CONNECT_RESPONSE
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
	b[i] = okBytes[rand.Intn(len(okBytes))]
	}
	return string(b)
}


func CreateRoom(conn *websocket.Conn, size int) (Room, error){
	player := Player{
		Id: GenerateRandomId(),
		Conn: conn,
	}
	var players []Player
	players = append(players, player)
	dim := SIZES[size]
	room := Room{
		Id: GenerateRandomId(),
		Players: players,
		Dimensions: dim,
	}
	err := CreateGameTable(room.Id, dim)
	if err != nil {
		return room, err
	}
	ActiveGames[room.Id] = room
	return room, nil
}


func CreateGameTable(id string, dim []int) error {
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
	vals = []interface{}{}
	for {
		for {
			insertionStmtText += "(?, ?, ?, ?),"
			vals = append(vals, r, c, 0, "")
		}
	}
	
	insertionStmtText = insertionStmtText[0:len(sqlStr)-2]

	res, _ := stmt.Exec(vals...)
	insertionStmt, err:= db.Prepare(insertionStmtText)
	if err != nil {
		fmt.Print("Preparing insertion statement failed for CreateGameTable: ", err)
		return err
	}
	defer insertionStmt.Close()

	_, err = stmt.Exec(vals...)
	if err != nil {
		fmt.Print("Executing statement failed for insertion in CreateGameTable: ", err)
		return err
	}

	return nil
}


func main() {
	ActiveGames = make(map[string]Room)
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
	http.HandleFunc("/connect", Connect)
	for {
		if err := http.ListenAndServe(":80", nil); err != nil {
			fmt.Print("Catastrophic error serving", err)
			fmt.Print("Restarting server.")
													}
												}
}



