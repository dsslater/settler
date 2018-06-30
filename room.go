package main

type Room struct {
	Id         string
	Password   string
	Players    []Player
	Dimensions [2]int
}
