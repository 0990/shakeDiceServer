package net

import (
	"bytes"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
	"encoding/json"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriond = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"xujialong"},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	//send chan []byte
	sendMap chan map[string]interface{}
}

//func (c *Client) SendMsg(message []byte) {
//	c.send <- message
//}

func(c *Client)SendMsg(sendMap map[string]interface{}){
	c.sendMap <- sendMap
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("err:%v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.message <- ReceiveMessage{
			client:  c,
			message: message,
		}
	}

}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriond)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		//case message, ok := <-c.send:
		//	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		//	if !ok {
		//		c.conn.WriteMessage(websocket.CloseMessage, []byte{})
		//		return
		//	}
		//	w, err := c.conn.NextWriter(websocket.TextMessage)
		//	if err != nil {
		//		return
		//	}
		//	w.Write(message)
		//	n := len(c.send)
		//	for i := 0; i < n; i++ {
		//		w.Write(newline)
		//		w.Write(<-c.send)
		//	}
		//	if err := w.Close(); err != nil {
		//		return
		//	}
		case sendMap,ok:=<-c.sendMap:{
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				data,ok:=json.Marshal(sendMap)
				if ok!=nil{
					w.Write(data)
				}
				n := len(c.send)
				for i := 0; i < n; i++ {
					//w.Write(newline)
					//w.Write(<-c.send)
					data,ok:=json.Marshal(<-c.sendMap)
					if ok!=nil{
						w.Write(data)
					}
				}
				if err := w.Close(); err != nil {
					return
				}
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serverWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:  hub,
		conn: conn,
		//send: make(chan []byte, 256),
		sendMap:make(chan map[string]interface{},10),
	}
	client.hub.register <- client
	go client.writePump()
	go client.readPump()
}
