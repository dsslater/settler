package  main

type Cell struct {
	/* Represents a single space on a game board. */
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	City   bool   `json:"city"`
	Amount int    `json:"amount"`
	Owner  string `json:"owner"`
	Color  string `json:"color"`
}