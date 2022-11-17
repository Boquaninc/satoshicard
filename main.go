package main

import (
	"flag"
	"satoshicard/conf"
	"satoshicard/ui"
	"time"
)

type Flags struct {
	Env string
}

func main() {
	// uictx := &ui.UIContext{}
	flags := &Flags{}
	flag.StringVar(&flags.Env, "env", "", "")
	flag.Parse()

	conf.Init(flags.Env)
	config := conf.GetConfig()
	ui.NewUIContext(config)
	for {
		time.Sleep(time.Minute)
	}
}
