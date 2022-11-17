package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"satoshicard/client"
	"satoshicard/conf"
	"satoshicard/core"
	"satoshicard/server"
	"satoshicard/util"
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
	EVENT_JOINED   Event = "joined"
	EVENT_PREIMAGE Event = "preimage"
	EVENT_WIN      Event = "win"
	EVENT_LOSE     Event = "lose"
)

type UIStateCode int

const (
	UI_STATE_WAIT_DECIDE_MODE   UIStateCode = 0
	UI_STATE_WAIT_PLAYER        UIStateCode = 1
	UI_STATE_WAIT_PREIMAGE_UTXO UIStateCode = 2
	UI_STATE_WAIT_RESULT        UIStateCode = 3
)

type UIState struct {
	Code   UIStateCode
	Params []string
}

type UIContext struct {
	Id                 string
	State              *UIState
	EventChannel       chan *UIEvent
	Client             client.Client
	Server             *server.HttpServer
	ParticipantContext *core.ParticipantContext
}

func NewUIContext(config *conf.Config) *UIContext {
	id := util.RandStringBytesMaskImprSrcUnsafe(8)
	ctx := &UIContext{
		Id:           id,
		EventChannel: make(chan *UIEvent),
		State:        &UIState{Code: UI_STATE_WAIT_DECIDE_MODE},
		Client:       nil,
	}
	gameContext := core.NewGameContext(id, config.ContractPath, ctx.OnAddParticipant)
	Server := server.NewHttpServer(gameContext, config.Listen)
	ctx.Server = Server
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

func (uictx *UIContext) CheckStateEqual(code UIStateCode) bool {
	if uictx.State.Code != code {
		return false
	}
	return true
}

func (uictx *UIContext) OnAddParticipant(uid string) {
	event := &UIEvent{
		Event:  EVENT_JOINED,
		Params: uid,
	}
	uictx.EventChannel <- event
}

func (uictx *UIContext) ReadLoop() {
	buf := bufio.NewReader(os.Stdin)
	for {
		lineByte, isPrefix, err := buf.ReadLine()
		if err != nil {
			panic(err)
		}
		if isPrefix {
			panic("too long input")
		}
		line := string(lineByte)
		index := strings.Index(line, ":")
		if index == -1 {
			fmt.Println("wrong fotmat command")
			continue
		}
		UIEvent := &UIEvent{
			Event:  Event(line[0:index]),
			Params: line[index+1:],
		}
		uictx.EventChannel <- UIEvent
	}
}

func (uictx *UIContext) DoEventHost(*UIEvent) error {
	if !uictx.CheckStateEqual(UI_STATE_WAIT_DECIDE_MODE) {
		return errors.New("command not for now")
	}
	uictx.Server.Open()
	uictx.ParticipantContext = core.NewParticipantContext(uictx.Id)
	uictx.SetState(UI_STATE_WAIT_PLAYER, nil)
	return nil
}

func (uictx *UIContext) DoEventJoin(event *UIEvent) error {
	if !uictx.CheckStateEqual(UI_STATE_WAIT_DECIDE_MODE) {
		return errors.New("command not for now")
	}
	client := client.NewHttpClient(uictx.Id, event.Params)
	response, err := client.Join()
	if err != nil {
		return err
	}
	uictx.Client = client
	uictx.ParticipantContext = core.NewParticipantContext(uictx.Id)
	uictx.SetState(UI_STATE_WAIT_PREIMAGE_UTXO, []string{response.Rival})
	return nil
}

func (uictx *UIContext) DoEventJoined(event *UIEvent) error {
	if !uictx.CheckStateEqual(UI_STATE_WAIT_PLAYER) {
		return errors.New("command not for now")
	}
	uictx.SetState(UI_STATE_WAIT_PREIMAGE_UTXO, []string{event.Params})
	return nil
}

func (uictx *UIContext) DoEventPreimage(event *UIEvent) error {
	if !uictx.CheckStateEqual(UI_STATE_WAIT_PREIMAGE_UTXO) {
		return errors.New("command not for now")
	}
	return nil
}

func (uictx *UIContext) DoEvent(event *UIEvent) error {
	switch event.Event {
	case EVENT_HOST:
		return uictx.DoEventHost(event)
	case EVENT_JOIN:
		return uictx.DoEventJoin(event)
	case EVENT_JOINED:
		return uictx.DoEventJoined(event)
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
	case UI_STATE_WAIT_PREIMAGE_UTXO:
		fmt.Printf("> we got a player %s,now game start,please input a really really big number\n", state.Params[0])
	case UI_STATE_WAIT_RESULT:
		fmt.Printf("> you got card %s,anthor player got card %s,please input result\n", state.Params[0], state.Params[1])
	default:
		panic("unknown state")
	}
}

func (uictx *UIContext) ProcessEvent() {
	event := <-uictx.EventChannel
	err := uictx.DoEvent(event)
	if err != nil {
		fmt.Printf("ProcessEvent DoEvent %s %s\n", event.Event, err)
		return
	}
	uictx.HandleState()
}

func (uictx *UIContext) ProcessEventLoop() {
	uictx.HandleState()
	for {
		uictx.ProcessEvent()
	}
}
