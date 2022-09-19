package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

func Check(err error) {
	if err != nil {
		log.Fatal("Problem Connecting DB:",err)
	}
}

type dbParam struct {
	Host string
	Port int
	User string
	Pass string
	DBN string
}

func InitDBConn() *sqlx.DB {
	dbjs := dbParam{Host: "127.0.0.1",Port: 5432,User: "postgres",Pass: "postgres",DBN: "postgres"}
	psqlInfo:=fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbjs.Host,dbjs.Port,dbjs.User,dbjs.Pass,dbjs.DBN)
	db, err := sqlx.Open("postgres", psqlInfo)
	Check(err)
	//defer db.Close()
	return db
}