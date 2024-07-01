package postgres

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"log"
)

var db *sql.DB

func InitDB(sourceName string) error {
	var err error

	db, err = sql.Open("postgres", sourceName)
	if err != nil {
		log.Println("Bad connection!", err)
		return err
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func StartDB() error {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	err := InitDB("host=" + host + " user=" + user + " password=" + pass +
		" dbname=" + name + " sslmode=disable ")
	if err != nil {
		log.Println("Bad connection!", err)
		return err
	}

	fmt.Print("Connected to a database!")
	db.Ping()

	return nil

}
