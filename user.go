package main

import "errors"

type UserState int

const (
	USER_STATE_IN_LOBBY         UserState = 0
	USER_STATE_HOST_GAME        UserState = 1
	USER_STATE_PARTICIPATE_GAME UserState = 2
)

type User struct {
	Id                string
	State             UserState
	HeartBeatLastTime int64
	RoomId            string
}

func (user *User) SetState(state UserState) {
	user.State = state
}

func (user *User) GetState() UserState {
	return user.State
}

func (user *User) SetRoomInfo(id string, state UserState) error {
	if user.GetState() != USER_STATE_IN_LOBBY {
		return errors.New("user not in lobby")
	}
	user.RoomId = id
	user.SetState(state)
	return nil
}
