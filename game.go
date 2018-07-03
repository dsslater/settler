package main

type Game struct {
	Id       string
	Password string
	Players  map[string]Player
	Height   int
	Width    int
}

func (g Game) getPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		players = append(players, player)
	}
	return players
}

func (g Game) getReadyPlayers() []Player {
	var players []Player
	for _, player := range g.Players {
		if player.Ready {
			players = append(players, player)
		}
	}
	return players
}

func (g Game) getNumReadyPlayers() int {
	return len(g.getReadyPlayers())
}

func (g Game) getPoints() []int {
	var points []int
	return points
}