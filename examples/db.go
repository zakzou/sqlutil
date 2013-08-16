package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/zakzou/sqlutil"
	"time"
)

type UserAlbum struct {
	Id         int
	UserId     int64
	AlbumId    int64
	OriginId   int64
	Status     int
	Source     int
	CreateTime time.Time
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	db, err := sql.Open("mysql", "root:root@/tl_album?charset=utf8&parseTime=true")
	checkErr(err)
	defer db.Close()

	// one
	rows, err := db.Query("select * from user_album_0 limit ?", 1)
	var one UserAlbum
	sqlutil.One(&one, rows)

	fmt.Println(one)

	rows, err = db.Query("select * from user_album_0 limit 10")
	var out []UserAlbum
	sqlutil.All(&out, rows)

	fmt.Println(out)
}
