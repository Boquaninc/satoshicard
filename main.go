package main

import (
	"errors"
	"flag"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

const (
	JOIN_GAME_URI   = "/join_game"
	SUBMIT_HASH_URI = "/submit_hash"
)

type GameContextState int

const (
	WAIT_DECIDE_MODE GameContextState = 0
	WAIT_PALYERS     GameContextState = 1
	WAIT_HASH        GameContextState = 2
)

type PlayerContext struct {
	Id       string
	Preimage *big.Int
	Hash     *big.Int
}

type GameContext struct {
	PlayerContexts []*PlayerContext
	State          GameContextState
	L              sync.Locker
}

func NewGameContext(config *Config) *GameContext {
	return &GameContext{
		PlayerContexts: make([]*PlayerContext, 0, 2),
		L:              &sync.Mutex{},
		State:          WAIT_PALYERS,
	}
}

func (gameContext *GameContext) SetStateNL(state GameContextState) {
	gameContext.State = state
}

func (gameContext *GameContext) SetStateL(state GameContextState) {
	gameContext.L.Lock()
	defer gameContext.L.Unlock()
	gameContext.SetStateNL(state)
}

// func (gameContext *GameContext) AddPlayerL(playerContext *PlayerContext) error {
// 	gameContext.L.Lock()
// 	defer gameContext.L.Unlock()
// 	if gameContext.State != WAIT_PALYERS {
// 		return errors.New("cant not add user right now")
// 	}
// 	if len(gameContext.PlayerContexts) >= 2 {
// 		return errors.New("already got 2 players")
// 	}

// 	gameContext.PlayerContexts = append(gameContext.PlayerContexts, playerContext)
// 	if len(gameContext.PlayerContexts) == 2 {
// 		gameContext.SetStateNL(WAIT_HASH)
// 	}
// 	return nil
// }

// func (gameContext *GameContext) AcceptPlayerL(joinGameRequest *JoinGameRequest) error {
// 	id := RandStringBytesMaskImprSrcUnsafe(8)
// 	answer := ""

// 	gameContext.L.Lock()
// 	defer gameContext.L.Unlock()

// 	if gameContext.State != WAIT_PALYERS {
// 		return errors.New("not host a game for now")
// 	}
// 	for {
// 		fmt.Printf("there is a user %s want to start a game,do you accept?[y/n]\n", id)
// 		_, err := fmt.Scanf("%s", &answer)
// 		if err != nil {
// 			return err
// 		}
// 		if answer == "y" {
// 			return &JoinGameResponse{
// 				Id: id,
// 			}
// 		} else if answer == "n" {
// 			return nil, errors.New("request refuse")
// 		}
// 	}

// }

func (GameContext *GameContext) StartGame() {

}

type JoinGameRequest struct {
}

type JoinGameResponse struct {
	Id string `json:"id"`
}

func (gameContext *GameContext) JoinGame(rsp http.ResponseWriter, req *http.Request, request *JoinGameRequest) (*JoinGameResponse, error) {
	id := RandStringBytesMaskImprSrcUnsafe(8)
	answer := ""
	for {
		fmt.Printf("there is a user %s want to start a game,do you accept?[y/n]\n", id)
		_, err := fmt.Scanf("%s", &answer)
		if err != nil {
			return nil, err
		}
		if answer == "y" {
			return &JoinGameResponse{
				Id: id,
			}, nil
		} else if answer == "n" {
			return nil, errors.New("request refuse")
		}
	}
}

type SubmitHashRequest struct {
	Hash string `json:"hash"`
}

type SubmitHashResponse struct {
}

func (gameContext *GameContext) SubmitHash(rsp http.ResponseWriter, req *http.Request, request *SubmitHashRequest) (*SubmitHashResponse, error) {
	panic("todo")
}

func (gameContext *GameContext) StartServer(config *Config) {
	http.HandleFunc(JOIN_GAME_URI, Aspect(gameContext.JoinGame))
	http.HandleFunc(SUBMIT_HASH_URI, Aspect(gameContext.SubmitHash))
	err := http.ListenAndServe(config.Listen, nil)
	if err != nil {
		panic(err)
	}
}

func (gameContext *GameContext) SelectModeL() {
	gameContext.L.Lock()
	defer gameContext.L.Unlock()
	if gameContext.State != WAIT_DECIDE_MODE {
		return
	}
	for {
		fmt.Println("press 0 to host a game,or press 1 to join a game")
		// command := -1
	}
}

func (gameContext *GameContext) EventLoop() {
	for {
		gameContext.SelectModeL()
	}
}

func main() {
	env := ""
	flag.StringVar(&env, "env", "", "env")
	flag.Parse()

	InitConfig(env)
	config := gConfig
	gameContext := NewGameContext(config)
	go gameContext.StartServer(config)
	for {
		time.Sleep(time.Hour)
	}
}
