package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Tensai75/nntp"
)

var (
	connectionGuard chan struct{}
	connectionOnce  sync.Once
)

func ConnectNNTP() (*nntp.Conn, error) {
	connectionOnce.Do(func() {
		connectionGuard = make(chan struct{}, conf.Server.Connections)
	})
	connectionGuard <- struct{}{} // will block if guard channel is already filled
	var conn *nntp.Conn
	var err error
	if conf.Server.SSL {
		conn, err = nntp.DialTLS("tcp", conf.Server.Host+":"+strconv.Itoa(conf.Server.Port), nil)

	} else {
		conn, err = nntp.Dial("tcp", conf.Server.Host+":"+strconv.Itoa(conf.Server.Port))
	}
	if err != nil {
		fmt.Printf("Connection to usenet server failed: %v\n", err)
		conn.Quit()
		return nil, err
	}
	if err := conn.Authenticate(conf.Server.User, conf.Server.Password); err != nil {
		conn.Quit()
		fmt.Printf("Authentication with usenet server failed: %v\n", err)
		return nil, err

	}
	return conn, nil
}

func DisconnectNNTP(conn *nntp.Conn) {
	if conn != nil {
		conn.Quit()
		select {
		case <-connectionGuard:
			// go on
		default:
			// go on
		}
	}
	conn = nil
}
