TDb
===

###Description

the golang sql driver implementation for testing.

it is not really implemented the SQL,only mapping to json data file.

###Installation

```
go get github.com/Centny/TDb
````

###Example
```
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

```

json:

```
{
	"INSERT INTO TESTING VALUES(?,?,?,?)": {
		"[1 2 3 4]": {
			"LIID": 1,
			"AROW": 1
		},
		"V_COUNTS":4 //target sql statement count.
	},
	"INSERT INTO TESTING2 VALUES(?,?,?,?)": {
		"[1 2 3 4]": {
			"LIID": 1,
			"AROW": 1
		},
		"V_COUNTS":4
	},
	"SELECT * FROM TESTING WHERE ID=? AND NAME=?": {
		"[1 a1]": [
			{
				"ID": 1,
				"NAME": "testing",
				"V3": "value abc"
			},
			{
				"ID": 1,
				"NAME": "testing"
			}
		],
		"V_COUNTS":2,
		"V_COLUMN":"ID,NAME,V3" //column name
	}
}
```
