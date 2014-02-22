package main

import (
	"database/sql"
	"fmt"
	"github.com/Centny/TDb"
)

func main() {
	TDb.TarErrs = TDb.LAST_INSERT_ID_ERR | TDb.ROWS_AFFECTED_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	res, err := tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)", 1, 2, 3, 4)
	if err != nil {
		panic(err)
	}
	_, err = res.LastInsertId()
	if err == nil {
		panic("not error")
	}
	fmt.Println(err)
	_, err = res.RowsAffected()
	if err == nil {
		panic("not error")
	}
	fmt.Println(err)
	fmt.Println("all end...")
}
