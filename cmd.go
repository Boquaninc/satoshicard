package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Method int

type Command struct {
	Method Method
	Params string
}

type CmdMethod func(*Command) (string, error)

type CmdMux struct {
	Methods map[Method]CmdMethod
}

var gCmdDefaultMux *CmdMux = &CmdMux{}

type CmdServer struct {
}

func (this *CmdServer) ReadAndServe() {
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
		method, err := strconv.ParseInt(ss[0], 10, 64)
		if err != nil {
			fmt.Println("wrong fotmat of command 2")
			continue
		}
		command := &Command{
			Method: Method(method),
			Params: ss[1],
		}
		cmdMethod, ok := gCmdDefaultMux.Methods[command.Method]
		if !ok {
			fmt.Println("method not found")
			continue
		}
		result, err := cmdMethod(command)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(result)
	}
}

func ReadAndServe() {

}

func CmdAspect() {

}
