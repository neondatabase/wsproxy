package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"encoding/base64"

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
	fmt.Printf("handling connection to %s\n", name)

	socket, err := net.Dial("tcp", name)
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, 10024)
		for {
			n, err := socket.Read(buf)
			if err != nil {
				fmt.Printf("failed to read from socket: %v\n", err)
				return
			}

			fmt.Printf("Got %d bytes pg->client: %s\n", n, base64.StdEncoding.EncodeToString(buf[:n]))

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

		fmt.Printf("Got %d bytes client->pg: %s\n", len(b), base64.StdEncoding.EncodeToString(b))

		_, err = io.Copy(socket, bytes.NewReader(b))
		if err != nil {
			return err
		}
	}
}

func main() {
	port := os.Getenv("LISTEN_PORT")
	if port == "" {
		port = ":80"
	}

	http.HandleFunc("/", Handle)
	fmt.Printf("Starting server on port %s\n", port)
	http.ListenAndServe(port, nil)
}
