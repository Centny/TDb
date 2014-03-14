package TDb

import (
	"database/sql"
	"fmt"
	"github.com/Centny/Cny4go/dbutil"
	"strings"
	// "time"
	// "fmt"
	"testing"
)

func TestTDb(t *testing.T) {
	TarErrs = 0
	db, err := sql.Open("TDb", "td@tdata.json")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tx, err := db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	res, err := tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)", 1, 2, 3, 4)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(res.LastInsertId())
	fmt.Println(res.RowsAffected())
	res, err = tx.Exec("INSERT INTO T2 VALUES(?,?,?,?)", 1, 2, 3, 6)
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(res.LastInsertId())
	fmt.Println(res.RowsAffected())
	tx.Commit()
	tx, err = db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	rows, err := tx.Query("SELECT * FROM TESTING WHERE ID=? AND NAME=?", 1, "a1")
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(dbutil.DbRow2Map(rows))
	rows, err = tx.Query("SELECT * FROM T2 WHERE ID=? AND NAME=?", 1, "a8")
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Println(dbutil.DbRow2Map(rows))
	tx.Rollback()
	tx.Commit()

	//
	tx, err = db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	res, err = tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)1", 1, 2, 3, 4)
	if err == nil {
		t.Error("not error")
		return
	}
	res, err = tx.Exec("INSERT INTO TESTING2 VALUES(?,?,?,?)", 1, 2, 3, 4)
	if err == nil {
		t.Error("not error")
		return
	}
	res, err = tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)", 1, 2, 3, 5)
	if err == nil {
		t.Error("not error")
		return
	}
	rows, err = tx.Query("SELECT * FROM TESTING WHERE ID=? AND NAME=s?", 1, "a1")
	if err == nil {
		t.Error("not error")
		return
	}
	rows, err = tx.Query("SELECT * FROM TESTING WHERE ID=? AND NAME=?", 1, "a2")
	if err == nil {
		t.Error("not error")
		return
	}
	//
	td := FindTData("td")
	if td == nil {
		t.Error("td not found")
		return
	}
	td2 := FindTData("td2")
	if td2 != nil {
		t.Error("td found")
		return
	}
	//
	db.Close()

}
func TestConnClose(t *testing.T) {
	//
	TarErrs = 0
	var dr TDbDriver
	c, err := dr.Open("td@tdata.json")
	err = c.Close()
	if err != nil {
		t.Error(err.Error())
	}
	TarErrs = CONN_CLOSE_ERR
	err = c.Close()
	if err == nil {
		t.Error("not error")
	}
}
func TestErr(t *testing.T) {
	TarErrs = (OPEN_ERR | CONN_BEGIN_ERR | CONN_CLOSE_ERR)
	if !TarErrs.Is(OPEN_ERR) {
		t.Error("not LAST_INSERT_ID_ERR")
		return
	}
	if !TarErrs.Is(CONN_BEGIN_ERR) {
		t.Error("not LAST_INSERT_ID_ERR")
		return
	}
}
func TestOpenErr(t *testing.T) {
	TarErrs = OPEN_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	_, err = db.Begin()
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestOpenErr2(t *testing.T) {
	TarErrs = 0
	db, err := sql.Open("TDb", "td:tdata.json")
	_, err = db.Begin()
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestOpenErr3(t *testing.T) {
	TarErrs = 0
	db, err := sql.Open("TDb", "td@tdata2.json")
	_, err = db.Begin()
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestOpenErr4(t *testing.T) {
	TarErrs = 0
	db, err := sql.Open("TDb", "td@tdata_err.json")
	_, err = db.Begin()
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestBeginErr(t *testing.T) {
	TarErrs = CONN_BEGIN_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	_, err = db.Begin()
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestPrepareErr(t *testing.T) {
	TarErrs = PREPARE_ERR | COMMIT_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tx, err := db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)", 1, 2, 3, 4)
	if err == nil {
		t.Error("not error")
		return
	}
	err = tx.Commit()
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestExecQueryErr(t *testing.T) {
	TarErrs = STMT_EXEC_ERR | STMT_QUERY_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tx, err := db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)", 1, 2, 3, 4)
	if err == nil {
		t.Error("not error")
		return
	}
	tx, err = db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = tx.Query("SELECT * FROM TESTING WHERE ID=? AND NAME=?", 1, "a1")
	if err == nil {
		t.Error("not error")
		return
	}
}
func TestIRErr(t *testing.T) {
	TarErrs = LAST_INSERT_ID_ERR | ROWS_AFFECTED_ERR | ROLLBACK_ERR | ROWS_CLOSE_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tx, err := db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	res, err := tx.Exec("INSERT INTO TESTING VALUES(?,?,?,?)", 1, 2, 3, 4)
	if err != nil {
		t.Error(err.Error())
		return
	}
	_, err = res.LastInsertId()
	if err == nil {
		t.Error("not error")
		return
	}
	_, err = res.RowsAffected()
	if err == nil {
		t.Error("not error")
		return
	}
	tx.Commit()
	tx, err = db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	rows, err := tx.Query("SELECT * FROM TESTING WHERE ID=? AND NAME=?", 1, "a1")
	if err != nil {
		t.Error(err.Error())
		return
	}
	// fmt.Println(dbutil.DbRow2Map(rows))
	err = tx.Rollback()
	if err == nil {
		t.Error("not error")
		return
	}
	err = rows.Close()
	if err == nil {
		t.Error("not error")
		return
	}
	tx.Commit()
}
func TestIRErr2(t *testing.T) {
	TarErrs = STMT_CLOSE_ERR
	db, err := sql.Open("TDb", "td@tdata.json")
	if err != nil {
		t.Error(err.Error())
		return
	}
	tx, err := db.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	rows, err := tx.Query("SELECT * FROM TESTING WHERE ID=? AND NAME=?", 1, "a1")
	if err != nil {
		t.Error(err.Error())
		return
	}
	dbutil.DbRow2Map(rows)
	tx.Commit()
}
func TestStrCount(t *testing.T) {
	fmt.Println(strings.Count("aiaagaa", "a"))
}

func TestErrIs(t *testing.T) {
	TarErrs = CONN_BEGIN_ERR
	if !TarErrs.Is(CONN_BEGIN_ERR) {
		t.Error("not valid")
	}
	if TarErrs.Is(CONN_CLOSE_ERR) {
		t.Error("not valid")
	}
	ResetRTarErrsC()
	SetErrC(CONN_BEGIN_ERR, 1)
	if TarErrs.Is(CONN_BEGIN_ERR) {
		t.Error("not valid")
	}
	if !TarErrs.Is(CONN_BEGIN_ERR) {
		t.Error("not valid")
	}
}
