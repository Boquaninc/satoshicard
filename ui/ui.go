package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"satoshicard/client"
	"satoshicard/core"
	"satoshicard/server"
	"strings"
)

type Event string

type UIEvent struct {
	Event  Event
	Params string
}

const (
	EVENT_HOST     Event = "host"
	EVENT_JOIN     Event = "join"
	EVENT_PREIMAGE Event = "preimage"
	EVENT_WIN      Event = "win"
	EVENT_LOSE     Event = "lose"
)

type UIStateCode int

const (
	UI_STATE_WAIT_DECIDE_MODE UIStateCode = 0
	UI_STATE_WAIT_PLAYER      UIStateCode = 1
	UI_STATE_WAIT_PREIMAGE    UIStateCode = 2
	UI_STATE_WAIT_RESULT      UIStateCode = 3
)

type UIState struct {
	Code   UIStateCode
	Params []string
}

type UIContext struct {
	State        *UIState
	EventChannel chan *UIEvent
	Client       client.Client
	Server       *server.HttpServer
	GameContext  *core.GameContext
}

func NewUIContext() *UIContext {
	ctx := &UIContext{
		EventChannel: make(chan *UIEvent),
		State:        &UIState{Code: UI_STATE_WAIT_DECIDE_MODE},
		Client:       nil,
	}
	go ctx.ProcessEventLoop()
	go ctx.ReadLoop()
	return ctx
}

func (uictx *UIContext) SetState(code UIStateCode, params []string) {
	uictx.State = &UIState{
		Code:   code,
		Params: params,
	}
}

func (uictx *UIContext) ReadLoop() {
	buf := bufio.NewReader(os.Stdin)
	for {
		line, isPrefix, err := buf.ReadLine()
		if err != nil {
			panic(err)
		}
		if isPrefix {
			panic("too long input")
		}
		ss := strings.Split(string(line), ":")
		if len(ss) != 2 {
			fmt.Println("wrong fotmat of command 1")
			continue
		}
		UIEvent := &UIEvent{
			Event:  Event(ss[0]),
			Params: ss[1],
		}
		uictx.EventChannel <- UIEvent
	}
}

func (uictx *UIContext) DoEvent(event *UIEvent) error {
	switch event.Event {
	case EVENT_HOST:
		panic("todo")
	case EVENT_JOIN:
		panic("todo")
	case EVENT_PREIMAGE:
		panic("todo")
	case EVENT_WIN:
		panic("todo")
	case EVENT_LOSE:
		panic("todo")
	default:
		return errors.New("unknown event")
	}
}

func (uictx *UIContext) HandleState() {
	state := uictx.State
	switch state.Code {
	case UI_STATE_WAIT_DECIDE_MODE:
		fmt.Printf("> join a game or host a game\n")
	case UI_STATE_WAIT_PLAYER:
		fmt.Printf("> host a game successful,now please wait for another player\n")
	case UI_STATE_WAIT_PREIMAGE:
		fmt.Printf("> we got a player %s,now game start,please input a really really big number\n", state.Params[0])
	case UI_STATE_WAIT_RESULT:
		fmt.Printf("> you got card %s,anthor player got card %s,please input result\n", state.Params[0], state.Params[1])
	default:
		panic("unknown state")
	}
}

func (uictx *UIContext) ProcessEvent() {
	event := <-uictx.EventChannel
	uictx.DoEvent(event)
	uictx.HandleState()
}

func (uictx *UIContext) ProcessEventLoop() {
	uictx.HandleState()
	for {
		uictx.ProcessEvent()
	}
}
