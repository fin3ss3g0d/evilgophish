package main

import (
	"github.com/gorilla/websocket"
	"encoding/json"
)

type FeedEvent struct {
    Event 	string `json:"event"`
    Time 	string `json:"time"`
    Message string `json:"message"`
}

type FeedEventCaptured struct {
	Event 	string `json:"event"`
    Time 	string `json:"time"`
    Message string `json:"message"`
	Tokens  string `json:"tokens"`
}

func main()  {
    c, _, err := websocket.DefaultDialer.Dial("ws://localhost:1337/ws", nil)
    if err != nil {
        return
    }
    defer c.Close()

    fe := FeedEventCaptured{}
    fe.Event = "Captured Session"
    fe.Message = "Session has been captured for victim: <strong>Victim</strong>"
	fe.Tokens = "Test token"
    fe.Time = "Right Now"
    data, _ := json.Marshal(fe)

    err = c.WriteMessage(websocket.TextMessage, []byte(string(data)))
    if err != nil {
        return 
    }
}