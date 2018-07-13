package  main

import (
	"github.com/gorilla/websocket"
)


// Player maintains all user specfic data.
type Player struct {
	ID   string
	Conn *websocket.Conn
	Ready bool
	Color string
}


// CreatePlayer returns a pointer to a newly created Player object.
func CreatePlayer(conn *websocket.Conn) *Player {
	player := Player{
		ID: GenerateRandomID(),
		Conn: conn,
	}
	return &player
}