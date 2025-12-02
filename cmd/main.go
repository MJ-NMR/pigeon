package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

var debug = false

const address = ":6969"

type user struct {
	name string
	conn net.Conn
}

var users = make(map[string]*user)

func (u *user) read() (msg string, err error) {
	reader := bufio.NewReader(u.conn)
	msg, err = reader.ReadString('\n')
	if err != nil {
		flog(err.Error())
		delete(users, u.name)
		return "", err
	}
	return msg , nil
}

func (u *user) write(s string) {
	_, err := u.conn.Write([]byte(s))
	if err != nil {
		flog(err.Error())
		delete(users, u.name)
		return
	}
}

func flog(s string) {
	if debug {
		log.Print(s)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		log.Print("debug are ON")
		debug = true
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("server start Listen on ", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err)
		}
		go login(conn)
	}
}

func login(conn net.Conn) {
	_, err := conn.Write([]byte("Hello! Wilcome to ZA3TER chat server\n"))
	if err != nil {
		flog(err.Error())
		conn.Close()
	}
	input := make([]byte, 10)

	for {

		_, err = conn.Write([]byte("Username: "))
		if err != nil {
			flog(err.Error())
			conn.Close()
		}

		n, err := conn.Read(input)
		if err != nil {
			conn.Write([]byte("Sorry...Something Wrong Happen"))
			flog(err.Error())
			conn.Close()
		}

		username := string(input[:n-1])
		if n > 0 {
			_, ok := users[username]
			if ok {
				conn.Write([]byte("Username exist tye something else"))
				continue
			}
			users[username] = &user{username, conn}
			flog("+++ added user " + username)
			go messageReader(username)
			return
		}
	}
}

func messageReader(username string) {
	u := users[username]
	for {
		msg, err := u.read()
		if err != nil {
			return
		}
		err = hub(username, msg)
		if err != nil {
			return
		}
	}
}

func hub(username, msg string) error {
	fullmsg := "\x1b[32m"+username+":\x1b[0m "+msg
	for v := range users {
		if v == username { continue }
		users[v].write(fullmsg)
	}
	return nil
}
