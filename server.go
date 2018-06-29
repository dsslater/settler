package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"fmt"
	"math/rand"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
)

type ConnectMessage struct {
	Room string `json:"room"`
	Password string `json:"password"`
}

func CheckOriginFunc (r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     CheckOriginFunc,
}

const okBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
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
		CreateRoom(conn)
	}
	if err := conn.WriteMessage(websocket.TextMessage, []byte("TEST")); err != nil {
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

func CreateRoom(conn *websocket.Conn) {
	player := Player{
		Id: GenerateRandomId(),
		Conn: conn,
	}
	var players []Player
	players = append(players, player)
	room := Room{
		Id: GenerateRandomId(),
		Players: players,
	}
	fmt.Print(room)
}


func main() {
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



