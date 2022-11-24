package main

import (
	"flag"
	"fmt"
	"satoshicard/conf"
	"satoshicard/ui"
	"time"
)

func WaitInput() {
	waitinput := ""
	fmt.Scanf("%s", &waitinput)
}

func DoMode1() {
	config := conf.GetConfig()
	uictx := ui.NewUIContext(config, 1)
	hostEvent := &ui.UIEvent{
		Event:  ui.EVENT_HOST,
		Params: "",
	}
	uictx.EventChannel <- hostEvent

	WaitInput()

	preimageEvent := &ui.UIEvent{
		Event:  ui.EVENT_PREIMAGE,
		Params: "22",
	}
	uictx.EventChannel <- preimageEvent

	WaitInput()

	signEvent := &ui.UIEvent{
		Event:  ui.EVENT_SIGN,
		Params: "",
	}
	uictx.EventChannel <- signEvent

	WaitInput()
	publishEvent := &ui.UIEvent{
		Event:  ui.EVENT_PUBLISH,
		Params: "",
	}
	uictx.EventChannel <- publishEvent

	WaitInput()
	openEvent := &ui.UIEvent{
		Event:  ui.EVENT_OPEN,
		Params: "",
	}
	uictx.EventChannel <- openEvent

	WaitInput()
	checkEvent := &ui.UIEvent{
		Event:  ui.EVENT_CHEKC,
		Params: "",
	}
	uictx.EventChannel <- checkEvent

	WaitInput()
	loseEvent := &ui.UIEvent{
		Event:  ui.EVENT_LOSE,
		Params: "",
	}
	uictx.EventChannel <- loseEvent
}

func DoMode2() {
	config := conf.GetConfig()
	uictx := ui.NewUIContext(config, 2)
	joinEvent := &ui.UIEvent{
		Event:  ui.EVENT_JOIN,
		Params: "127.0.0.1:10001",
	}
	uictx.EventChannel <- joinEvent

	WaitInput()

	preimageEvent := &ui.UIEvent{
		Event:  ui.EVENT_PREIMAGE,
		Params: "27",
	}
	uictx.EventChannel <- preimageEvent

	WaitInput()
	signEvent := &ui.UIEvent{
		Event:  ui.EVENT_SIGN,
		Params: "",
	}
	uictx.EventChannel <- signEvent

	WaitInput()
	openEvent := &ui.UIEvent{
		Event:  ui.EVENT_OPEN,
		Params: "",
	}
	uictx.EventChannel <- openEvent

	// WaitInput()
	// takedepositEvent := &ui.UIEvent{
	// 	Event:  ui.EVENT_TAKEDEPOSIT,
	// 	Params: "",
	// }
	// uictx.EventChannel <- takedepositEvent

	WaitInput()
	checkEvent := &ui.UIEvent{
		Event:  ui.EVENT_CHEKC,
		Params: "",
	}
	uictx.EventChannel <- checkEvent

	WaitInput()
	winEvent := &ui.UIEvent{
		Event:  ui.EVENT_WIN,
		Params: "2",
	}
	uictx.EventChannel <- winEvent
}

func DoMode0() {
	config := conf.GetConfig()
	ui.NewUIContext(config, 0)
}

func main() {
	conf.Init()
	config := conf.GetConfig()
	if config.Help {
		flag.Usage()
		return
	}
	switch config.Mode {
	case 0:
		DoMode0()
	case 1:
		DoMode1()
	case 2:
		DoMode2()
	default:
		panic("unknown mode")
	}
	for {
		time.Sleep(time.Minute)
	}
}
