package main

import (
	"ccz/utils"
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	utils.LoadEnv(".env")
	db, err := sql.Open("mysql", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
create table if not exists users (
	id int auto_increment primary key,
	email varchar(255) unique,
	password varchar(255),
	full_name varchar(255),
	telephone varchar(50),
	provider varchar(20)
)
`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("migration completed")
}
