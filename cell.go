package  main

type Cell struct {
	Row    int    `json:"row"`
	Col    int    `json:"col"`
	City   bool   `json:"city"`
	Amount int    `json:"amount"`
	Owner  string `json:"owner"`
	Color  string `json:"color"`
}