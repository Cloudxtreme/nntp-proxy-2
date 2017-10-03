package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
)

const (
	host = "10.1.0.101"
	port = "3333"
	user = "XXX"
	pass = "XXX"
)

type session struct {
	UserConnection    net.Conn
	backendConnection net.Conn
	command           string
}

func main() {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + host + ":" + port)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func (s *session) dispatchCommand() {

	log.Println("Dispatch")
	log.Printf("Command : %v", s.command)

	cmd := strings.Split(s.command, " ")

	args := []string{}

	if len(cmd) > 1 {
		args = cmd[1:]
	}

	if strings.ToLower(cmd[0]) == "authinfo" {
		s.handleAuth(args)
	} else {
		go s.handleRequests()
	}

}

func (s *session) handleRequests() {
	if s.backendConnection != nil {

		for {
			io.CopyN(s.backendConnection, s.UserConnection, 1024)
		}

	}
}

func (s *session) handleAuth(args []string) {
	t := textproto.NewConn(s.UserConnection)

	if len(args) < 2 {
		t.PrintfLine("502 Unknown Syntax!")
		return
	}

	if strings.ToLower(args[0]) != "user" {
		t.PrintfLine("502 Unknown Syntax!")
		return
	}

	t.PrintfLine("350 Continue")

	a, _ := t.ReadLine()
	parts := strings.SplitN(a, " ", 3)

	if strings.ToLower(parts[0]) != "authinfo" || strings.ToLower(parts[1]) != "pass" {
		t.PrintfLine("502 Unknown Syntax!")
		return
	}

	if args[1] == "Test" && parts[2] == "Test" {

		conn, err := net.Dial("tcp", "nntp.ovpn.to:11900")

		conn.Write([]byte("AUTHINFO USER " + user + "\n"))
		conn.Write([]byte("AUTHINFO PASS " + pass + "\n"))
		conn.Write([]byte("ARTICLE <part15of66.maG0joHg6Vj2Hy4TLsBs@powerpost2000AA.local>" + "\n"))

		if err == nil {
			t.PrintfLine("281 Welcome")
			s.backendConnection = conn
			return
		} else {
			t.PrintfLine("502 ERROR!")
			return
		}

	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {

	c := textproto.NewConn(conn)

	sess := &session{
		conn,
		nil,
		"",
	}

	c.PrintfLine("200 Welcome to NNTP Proxy!")
	for {
		l, err := c.ReadLine()
		if err != nil {
			log.Printf("Error reading from client, dropping conn: %v", err)
			conn.Close()
			return

		}

		sess.command = l
		sess.dispatchCommand()
	}

}
