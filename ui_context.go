package main

type UIState int

const (
	UI_LOBBY            UIState = 0
	UI_HOST_GAME        UIState = 1
	UI_PARTICIPATE_GAME UIState = 2
)

type UIContext struct {
	State      UIState
	GameServer *GameServer
}

func NewUIContext(GameServer *GameServer) *UIContext {
	return &UIContext{
		GameServer: GameServer,
	}
}

func (uictx *UIContext) Read() {
	// buf := bufio.NewReader(os.Stdin)
	// for {
	// 	line, isPrefix, err := buf.ReadLine()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	if isPrefix {
	// 		panic("too long input")
	// 	}
	// 	ss := strings.Split(string(line), ":")
	// 	if len(ss) != 2 {
	// 		fmt.Println("wrong fotmat of command 1")
	// 		continue
	// 	}
	// 	method, err := strconv.ParseInt(ss[0], 10, 64)
	// 	if err != nil {
	// 		fmt.Println("wrong fotmat of command 2")
	// 		continue
	// 	}
	// 	command := &Command{
	// 		Method: Method(method),
	// 		Params: ss[1],
	// 	}
	// 	cmdMethod, ok := gCmdDefaultMux.Methods[command.Method]
	// 	if !ok {
	// 		fmt.Println("method not found")
	// 		continue
	// 	}
	// 	result, err := cmdMethod(command)
	// 	if err != nil {
	// 		fmt.Println(err.Error())
	// 		continue
	// 	}
	// 	fmt.Println(result)
	// }
}
