package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

const (
	JOIN_GAME_URI   = "/join_game"
	SUBMIT_HASH_URI = "/submit_hash"
)

func PrintJson(i interface{}) {
	ib, ok := i.([]byte)
	if ok {
		fmt.Println(string(ib))
		return
	}
	is, ok := i.(string)
	if ok {
		fmt.Println(is)
		return
	}
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

const (
	METHOD_PRINT_METHOD Method = 0
	METHOD_QUIT         Method = 1
	METHOD_CREATE_ROOM  Method = 2
	METHOD_JOIN_ROOM    Method = 3
)

func main() {
	env := ""
	flag.StringVar(&env, "env", "", "env")
	flag.Parse()
	InitConfig(env)
}
