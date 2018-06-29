package  main

import (
	"github.com/gorilla/websocket"
)

type Player struct {
	Id string
	Conn *websocket.Conn
}
