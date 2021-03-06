package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/howeyc/gopass"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	DatabaseDriver string
	DatabaseSource string
}

var config Config
var db *sql.DB

func main() {
	str, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(str, &config)
	if err != nil {
		panic(err)
	}

	db, err = sql.Open(config.DatabaseDriver, config.DatabaseSource)
	if err != nil {
		panic(err)
	}

	switch {
	case len(os.Args) == 3 && os.Args[1] == "set":
		SetPassword(os.Args[2], nil)
	case len(os.Args) == 4 && os.Args[1] == "set":
		SetPassword(os.Args[2], []byte(os.Args[3]))
	case len(os.Args) == 3 && os.Args[1] == "check":
		CheckPassword(os.Args[2], nil)
	case len(os.Args) == 4 && os.Args[1] == "check":
		CheckPassword(os.Args[2], []byte(os.Args[3]))
	default:
		fmt.Println("usage: password (set|check) username [password]")
	}
}

func SetPassword(username string, password []byte) {
	if password == nil {
		fmt.Print("Password: ")
		password = gopass.GetPasswdMasked()

		fmt.Print("Retype Password: ")
		retype := gopass.GetPasswdMasked()

		if string(password) != string(retype) {
			fmt.Println("Passwords not match")
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	var dummy string
	err = db.QueryRow("select password from users where name = $1", username).Scan(&dummy)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if err == sql.ErrNoRows {
		_, err = db.Exec("insert into users (name, password) values ($1, $2)", username, string(hash))
	} else {
		_, err = db.Exec("update users set password = $1 where name = $2", string(hash), username)
	}
	if err != nil {
		panic(err)
	}
}

func CheckPassword(username string, password []byte) {
	var hash string
	err := db.QueryRow("select password from users where name = $1", username).Scan(&hash)
	if err == sql.ErrNoRows {
		fmt.Println("User not found:", username)
		return
	} else if err != nil {
		panic(err)
	}

	if password == nil {
		fmt.Print("Password: ")
		password = gopass.GetPasswdMasked()
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), password)
	if err == nil {
		println("OK")
	} else {
		println("Bad password")
	}
}
