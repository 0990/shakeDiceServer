package net

import (
	"flag"
	"log"
	"net/http"
)

var connectFunc func(*Client)
var disconnecFunc func(*Client)
var messageFunc func(*Client, []byte)

func RegisterConnectFun(handle func(*Client)) {
	connectFunc = handle
}

func RegisterDisconnectFun(handle func(*Client)) {
	disconnecFunc = handle
}

func RegisterMessageFun(handle func(*Client, []byte)) {
	messageFunc = handle
}

type Hub struct {
	clients    map[*Client]bool
	message    chan ReceiveMessage
	register   chan *Client
	unregister chan *Client
}

type ReceiveMessage struct {
	client  *Client
	message []byte
}

func newHub() *Hub {
	return &Hub{
		message:    make(chan ReceiveMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run(workerChan chan func()) {
	for {
		select {
		case client := <-h.register:
			workerChan <- func() {
				connectFunc(client)
				//h.clients[client] = true
			}
		case client := <-h.unregister:
			workerChan <- func() {
				disconnecFunc(client)
				close(client.send)
				//if _, ok := h.clients[client]; ok {
				//	delete(h.clients, client)
				//	close(client.send)
				//}
			}
		case message := <-h.message:
			workerChan <- func() {
				messageFunc(message.client, message.message)
				//for client := range h.clients {
				//	select {
				//	case client.send <- message:
				//	default:
				//		close(client.send)
				//		delete(h.clients, client)
				//	}
				//}
			}
		}
	}
}

var addr = flag.String("addr", ":8080", "http service address")

func Run(workerChan chan func()) {
	hub := newHub()
	go hub.run(workerChan)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServer:", err)
	}
}
