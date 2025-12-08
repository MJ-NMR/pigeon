package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var debug = false

const (
	address      = ":6969"
	msgMaxLenght = 50
	dbPath       = "./users.sqlit3"
)

var users = make(map[string]*net.Conn)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--debug" {
		fmt.Print("debug are ON")
		debug = true
	}
	db := openDB(dbPath)
	defer db.Close()

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("server start Listen on ", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Print(err)
		}
		go login(conn)
	}
}

func flog(s string) {
	if debug {
		fmt.Print(s)
	}
}

func login(conn net.Conn) {
	// TODO: auth & database
	_, err := conn.Write([]byte("Hello! Welcome to ZA3TER chat server\n"))
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
			continue
		}

		username := string(input[:n-1])
		if n > 0 {
			_, ok := users[username]
			if ok {
				conn.Write([]byte("Username exist tye something else"))
				continue
			}
			users[username] = &conn
			flog("+++ added user " + username)
			go messageReader(username)
			return
		}
	}
}

func messageReader(username string) {
	conn := *users[username]
	buff := make([]byte, msgMaxLenght)
	var msg string
	for {
		n, err := conn.Read(buff)
		if err != nil {
			flog(err.Error())
			delete(users, username)
			return
		}

		msg += string(buff[:n])
		if buff[n-1] != '\n' {
			continue
		}
		err = hub(username, msg)
		if err != nil {
			flog(err.Error())
			return
		}
	}
}

func hub(username, msg string) error {
	if len(msg) > msgMaxLenght {
		return nil
	}
	fullmsg := fmt.Sprintf("\x1b[32m%s:\x1b[0m %s", username, msg)
	for resever := range users {
		if resever == username {
			continue
		}
		_, err := (*users[resever]).Write([]byte(fullmsg))
		if err != nil {
			flog(err.Error())
			delete(users, resever)
			return nil
		}
	}
	return nil
}

func openDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var exist bool
	qres := db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=users);")
	if qres.Scan(&exist); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if exist {
		return db
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (username TEXT PRIMARY KEY, password TEXT)")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return db
}
