package conn_tools

import (
	"context"
	"log"
	"net"
	"strings"
	"time"
)

type ConnTools struct {
	Host    string
	Timeout time.Duration
}

func (c *ConnTools) ConnHost(ctx context.Context) (is_ready bool) {
	// Go use HasPrefix or HasSuffix for StartWith or EndWith
	network := "tcp4"
	if strings.HasPrefix(c.Host, "unix://") {
		log.Printf("Unix socket")
		network = "unix"
		// Safe code so it would not be broken if large string and multiple substitue to decrease
		c.Host = strings.Replace(c.Host, "unix://", "", 1)
	}
	conn, err := net.DialTimeout(network, c.Host, c.Timeout)
	if err != nil {
		log.Printf("Connection Failed:" + err.Error())
		is_ready = false
	} else {
		defer conn.Close()
		is_ready = true
	}
	return
}
