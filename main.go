package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Handle(w http.ResponseWriter, r *http.Request) {
	// "dry-mountain-455633.cloud.neon.tech:5432"
	name := r.URL.Query().Get("name")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("failed to upgrade: %v\n", err)
		return
	}
	defer conn.Close()

	err = handleWebsocket(conn, name)
	if err != nil {
		fmt.Printf("failed to handle websocket: %v\n", err)
		return
	}
}

func handleWebsocket(conn *websocket.Conn, name string) error {
	socket, err := net.Dial("tcp", name)
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := socket.Read(buf)
			if err != nil {
				fmt.Printf("failed to read from socket: %v\n", err)
				return
			}

			err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				fmt.Printf("failed to write to websocket: %v\n", err)
				return
			}
		}
	}()

	for {
		_, b, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		_, err = socket.Write(b)
		if err != nil {
			return err
		}
	}
}

func main() {
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		port = ":8080"
	}

	http.HandleFunc("/", Handle)
	fmt.Printf("Starting server on port %s\n", port)
	http.ListenAndServe(port, nil)
}