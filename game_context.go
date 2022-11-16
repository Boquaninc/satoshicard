package main

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

func (gameContext *GameContext) SetState(gameContextState GameContextState) {
	gameContext.State = gameContextState
}

func (gameContext *GameContext) GetState() GameContextState {
	return gameContext.State
}

func (gameContext *GameContext) CreateRoom() error {
	if gameContext.State != WAIT_DECIDE_MODE {
		return errors.New("already in a room")
	}
	go gameContext.ListenAndServe()
	gameContext.SetState(WAIT_PLAYERS)
	return nil
}

func (gameContext *GameContext) PrintCommand() {
	// 	METHOD_PRINT_METHOD Method = 0
	// METHOD_QUIT         Method = 1
	// METHOD_CREATE_ROOM  Method = 2
	// METHOD_JOIN_ROOM    Method = 3
	fmt.Println(
		`0. print method
1. quit
2. create room
3. join room`,
	)
}

func (gameContext *GameContext) JoinRoom(ip string) {

}

func (gameContext *GameContext) ListenAndServe() {
	server := &http.Server{
		Addr: gConfig.Listen,
	}
	gameContext.Server = server
	fmt.Printf("ListenAndServe %s\n", gameContext.Listen)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func NewGameContext(config *Config) *GameContext {
	l := &sync.Mutex{}
	return &GameContext{
		PlayerContexts: make([]*PlayerContext, 0, 2),
		State:          WAIT_DECIDE_MODE,
		L:              l,
		Listen:         config.Listen,
	}
}
