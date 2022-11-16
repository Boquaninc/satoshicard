package main

import (
	"errors"
	"sync"
	"time"
)

type GameServer struct {
	UserSet      map[string]*User
	L            sync.Locker
	RoomSet      map[string]*GameContext
	ContractPath string
}

func NewGameServer(config *Config) *GameServer {
	return &GameServer{
		UserSet:      map[string]*User{},
		L:            &sync.Mutex{},
		RoomSet:      map[string]*GameContext{},
		ContractPath: config.ContractPath,
	}
}

func (gameServer *GameServer) LogInL() string {
	user := NewUser()
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	gameServer.UserSet[user.Id] = user
	return user.Id
}

func (gameServer *GameServer) GetAndTouchUser(id string) (*User, error) {
	user, ok := gameServer.UserSet[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	now := time.Now().Unix()
	user.HeartBeatLastTime = now
	return user, nil
}

func (gameServer *GameServer) CreateRoomL(id string) (string, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	user, err := gameServer.GetAndTouchUser(id)
	if err != nil {
		return "", err
	}
	room := NewGameContext(user.Id, gameServer.ContractPath)
	gameServer.RoomSet[room.Id] = room

	err = user.SetRoomInfo(room.Id, USER_STATE_HOST_GAME)
	if err != nil {
		return "", err
	}

	return room.Id, nil
}

func (gameServer *GameServer) GetRoomListL() ([]string, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	roomList := []string{}
	for id := range gameServer.RoomSet {
		roomList = append(roomList, id)
	}
	return roomList, nil
}

func (gameServer *GameServer) JoinRoomL(uid string, roomId string) error {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	user, err := gameServer.GetAndTouchUser(uid)
	if err != nil {
		return err
	}
	if user.GetState() != USER_STATE_IN_LOBBY {
		return errors.New("user not in lobby")
	}
	room, ok := gameServer.RoomSet[roomId]
	if !ok {
		return errors.New("room not found")
	}
	err = room.AddPlayer(user.Id)
	if err != nil {
		return err
	}
	user.SetRoomInfo(room.Id, USER_STATE_PARTICIPATE_GAME)
	return nil
}

func (gameServer *GameServer) JoinRandomRoomL(uid string) {
	// gameServer.L.Lock()
	// defer gameServer.L.Unlock()
	// user, err := gameServer.GetAndTouchUser(uid)
	// if err != nil {
	// 	return err
	// }
	// for _, room := range gameServer.RoomSet {
	// 	room.AddPlayer(uid)
	// }
}

func (gameServer *GameServer) SetStep1InfoL(uid string, step1Info *Step1Info) (chan string, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	user, err := gameServer.GetAndTouchUser(uid)
	if err != nil {
		return nil, err
	}
	room, ok := gameServer.RoomSet[user.RoomId]
	if !ok {
		return nil, errors.New("not in a room")
	}
	step1ResultChannel, err := room.SetStep1Info(uid, step1Info)
	if err != nil {
		return nil, err
	}
	return step1ResultChannel, nil
}

func (gameServer *GameServer) SetStep1InfoAndReceiveResult(uid string, step1Info *Step1Info) (string, error) {
	step1ResultChannel, err := gameServer.SetStep1InfoL(uid, step1Info)
	if err != nil {
		return "", err
	}
	rawTx := <-step1ResultChannel
	return rawTx, nil
}
