package main

type Room struct {
	Id       string
	Password string
	Players  map[string]Player
	Height   int
	Width    int
}
