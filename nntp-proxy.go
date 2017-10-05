package main

import (
	"crypto/tls"
	"fmt"
	"go-nntp-proxy/config"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
)

var (
	cfg                config.Configuration
	backendConnections map[string]int
)

type session struct {
	UserConnection    net.Conn
	backendConnection net.Conn
	command           string
}

func main() {

	cfg = config.LoadConfig("config.json")

	backendConnections = make(map[string]int)

	for _, elem := range cfg.Backend {
		backendConnections[elem.BackendName] = 0
	}

	var l net.Listener

	if cfg.Frontend.FrontendTLS {

		// New var for error
		var err error

		// try to load cert pair
		cer, err := tls.LoadX509KeyPair(cfg.Frontend.FrontendTLSCert, cfg.Frontend.FrontendTLSKey)

		if err != nil {
			log.Printf("%v", err)
			return
		}

		// Set certs
		tlsConf := &tls.Config{Certificates: []tls.Certificate{cer}}

		// Listen for incoming TLS connections.
		l, err = tls.Listen("tcp", cfg.Frontend.FrontendAddr+":"+cfg.Frontend.FrontendPort, tlsConf)

		if err != nil {
			log.Printf("%v", err)
			os.Exit(1)
		}

		log.Printf("[TLS] Listening on %v:%v", cfg.Frontend.FrontendAddr, cfg.Frontend.FrontendPort)

	} else {

		// New var for error
		var err error

		// Listen for incoming connections.
		l, err = net.Listen("tcp", cfg.Frontend.FrontendAddr+":"+cfg.Frontend.FrontendPort)

		if err != nil {
			log.Printf("%v", err)
			os.Exit(1)
		}

		log.Printf("[PLAIN - DO NOT USE PROD!] Listening on %v:%v", cfg.Frontend.FrontendAddr, cfg.Frontend.FrontendPort)
	}

	// Close the listener when the application closes.
	defer l.Close()

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

		s.backendConnection.Write([]byte(s.command + "\n"))

		go io.Copy(s.backendConnection, s.UserConnection)
		io.Copy(s.UserConnection, s.backendConnection)

		//defer s.backendConnection.Close()
	}
}

func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes)
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *session) handleAuthorization(user string, password string) bool {

	for _, elem := range cfg.Users {
		if elem.Username == user && CheckPasswordHash(password, elem.Password) {
			return true
		}
	}
	return false

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

	t.PrintfLine("381 Continue")

	a, _ := t.ReadLine()
	parts := strings.SplitN(a, " ", 3)

	if strings.ToLower(parts[0]) != "authinfo" || strings.ToLower(parts[1]) != "pass" {
		t.PrintfLine("502 Unknown Syntax!")
		return
	}

	if s.handleAuthorization(args[1], parts[2]) {

		for _, elem := range cfg.Backend {
			if elem.BackendConns < backendConnections[elem.BackendName] {

			}
			backendConnections[elem.BackendName] = 0
		}

		// New backend connection to upstream NNTP
		conn, err := net.Dial("tcp", "nntp.ovpn.to:11900")

		c := textproto.NewConn(conn)

		_, _, err = c.ReadCodeLine(200)
		if err != nil {
			return
		}

		err = c.PrintfLine("authinfo user %s", "")

		if err != nil {
			return
		}

		_, _, err = c.ReadCodeLine(381)
		if err != nil {
			return
		}

		err = c.PrintfLine("authinfo pass %s", "")
		if err != nil {
			return
		}
		_, _, err = c.ReadCodeLine(281)

		if err == nil {
			t.PrintfLine("281 Welcome")
			s.backendConnection = conn
			return
		} else {
			t.PrintfLine("502 ERROR!")
			return
		}

	} else {
		t.PrintfLine("502 AUTH FAILED!")
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
