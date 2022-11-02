package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"

	"encoding/base64"

	"github.com/caarlos0/env/v6"
	"github.com/gorilla/websocket"
)

type Config struct {
	ListenPort     string `env:"LISTEN_PORT" envDefault:":80"`
	AllowAddrRegex string `env:"ALLOW_ADDR_REGEX" envDefault:"^[a-zA-Z\\-0-9\\.]*\\.neon\\.tech\\:5432$"`
	AppendPort     string `env:"APPEND_PORT" envDefault:""`
	UseHostHeader  bool   `env:"USE_HOST_HEADER" envDefault:"false"`
	LogTraffic     bool   `env:"LOG_TRAFFIC" envDefault:"false"`
	LogConnInfo    bool   `env:"LOG_CONN_INFO" envDefault:"true"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func IsAddrAllowed(addr string, reg *regexp.Regexp) bool {
	if reg == nil {
		return true
	}

	return reg.MatchString(addr)
}

type ProxyHandler struct {
	cfg       *Config
	addrRegex *regexp.Regexp
}

func NewProxyHandler(config *Config) (*ProxyHandler, error) {
	var (
		addrRegex *regexp.Regexp
		err       error
	)
	if config.AllowAddrRegex != "" {
		addrRegex, err = regexp.Compile(config.AllowAddrRegex)
		if err != nil {
			return nil, err
		}
	}

	return &ProxyHandler{
		cfg:       config,
		addrRegex: addrRegex,
	}, nil
}

func (h *ProxyHandler) ExtractProxyDest(r *http.Request) (string, error) {
	addressArg := r.URL.Query().Get("address")
	hostHeader := r.Host

	addr := addressArg
	if h.cfg.UseHostHeader {
		addr = hostHeader
	}
	if h.cfg.AppendPort != "" {
		addr += h.cfg.AppendPort
	}

	allowed := IsAddrAllowed(addr, h.addrRegex)

	if h.cfg.LogConnInfo {
		log.Printf(
			"Got request from %s, proxying to %s, allowed=%v, addressArg=%v, hostHeader=%v",
			r.RemoteAddr,
			addr, allowed,
			addressArg,
			hostHeader,
		)
	}

	if !allowed {
		return "", fmt.Errorf("proxying to specified address not allowed")
	}

	return addr, nil
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.cfg.LogConnInfo {
		log.Printf("Got request from %s", r.RemoteAddr)
	}

	addr, err := h.ExtractProxyDest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade: %v\n", err)
		return
	}
	defer conn.Close()

	err = h.HandleWS(conn, addr)
	if err != nil {
		log.Printf("failed to handle websocket: %v\n", err)
		return
	}
}

func (h *ProxyHandler) HandleWS(conn *websocket.Conn, addr string) error {
	socket, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer socket.Close()

	go func() {
		const bufferSize = 32 * 1024
		buf := make([]byte, bufferSize)
		for {
			n, err := socket.Read(buf)
			if err != nil {
				log.Printf("failed to read from socket: %v\n", err)
				return
			}

			if h.cfg.LogTraffic {
				log.Printf("Got %d bytes pg->client: %s\n", n, base64.StdEncoding.EncodeToString(buf[:n]))
			}

			err = conn.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Printf("failed to write to websocket: %v\n", err)
				return
			}
		}
	}()

	for {
		_, b, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		if h.cfg.LogTraffic {
			log.Printf("Got %d bytes client->pg: %s\n", len(b), base64.StdEncoding.EncodeToString(b))
		}

		_, err = io.Copy(socket, bytes.NewReader(b))
		if err != nil {
			return err
		}
	}
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.AllowAddrRegex == "" {
		log.Printf("WARN: No regex for allowed addresses, allowing all")
	} else {
		log.Printf("Using regex for allowed addresses: %v", cfg.AllowAddrRegex)
	}

	handler, err := NewProxyHandler(&cfg)
	if err != nil {
		log.Fatalf("Failed to create proxy handler: %v", err)
	}

	http.Handle("/v1", handler)
	log.Printf("Starting server on port %s", cfg.ListenPort)
	err = http.ListenAndServe(cfg.ListenPort, nil) //nolint:gosec
	if err != nil {
		log.Fatalf("HTTP ListenAndServe finished with error: %v", err)
	}
}
