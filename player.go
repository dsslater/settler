package  main

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Id   string
	Conn *websocket.Conn
	Ready bool
	Color string
}

func createPlayer(conn *websocket.Conn) Player {
	player := Player{
		Id: GenerateRandomId(),
		Conn: conn,
	}
	return player
}