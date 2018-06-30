package main

type Room struct {
	Id         string
	Password   string
	Players    map[string]Player
	Dimensions [2]int
}
