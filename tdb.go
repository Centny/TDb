//Author:Centny
//Package TDb provide the testing sql driver.
package TDb

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
)

//the testing error.
var Err error = errors.New("TDb testing error")

//create new not found error.
func NotFoundErr(msg ...interface{}) error {
	return errors.New(fmt.Sprintf("data not found %v", msg))
}

//the type of TDbErr
type TDbErr uint32

const (
	OPEN_ERR TDbErr = 1 << iota
	CONN_BEGIN_ERR
	CONN_CLOSE_ERR
	ROLLBACK_ERR
	COMMIT_ERR
	PREPARE_ERR
	STMT_CLOSE_ERR
	STMT_QUERY_ERR
	STMT_EXEC_ERR
	ROWS_CLOSE_ERR
	ROWS_NEXT_ERR
	LAST_INSERT_ID_ERR
	ROWS_AFFECTED_ERR
)

//the error will be panic
//usage:Errs=OPEN_ERR|CONN_BEGIN_ERR
var TarErrs TDbErr = 0

//if error contain target error.
func (t TDbErr) Is(e TDbErr) bool {
	return (t & e) == e
}

//all testing data from json file.
var TData map[string]map[string]interface{} = make(map[string]map[string]interface{})

//add one json file data to TData by name and file page.
func AddTData(n string, path string) (map[string]interface{}, error) {
	bys, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(strings.NewReader(string(bys)))
	var data map[string]interface{}
	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}
	TData[n] = data
	return data, nil
}

//find the testing data by name.
func FindTData(n string) map[string]interface{} {
	if v, ok := TData[n]; ok {
		return v
	} else {
		return nil
	}
}

//
func init() {
	Register("TDb")
}

//register one drive to system by name.
func Register(n string) {
	sql.Register(n, &TDbDriver{N: n})
}

type TDbDriver struct {
	N string //driver name
}

func (d *TDbDriver) Open(dsn string) (driver.Conn, error) {
	if TarErrs.Is(OPEN_ERR) {
		return nil, Err
	}
	dsns := strings.SplitN(dsn, "@", 2)
	if len(dsns) < 2 {
		return nil, errors.New(fmt.Sprintf("dsn error:%s", dsn))
	}
	data, err := AddTData(dsns[0], dsns[1])
	if err != nil {
		return nil, err
	}
	return &TDbConn{
		Name:   dsns[0],
		Path:   dsns[1],
		Driver: d,
		TData:  data,
	}, nil
}

type TDbConn struct {
	Name   string //data key name
	Path   string //the file path.
	Driver *TDbDriver
	TData  map[string]interface{}
}

func (c *TDbConn) Begin() (driver.Tx, error) {
	if TarErrs.Is(CONN_BEGIN_ERR) {
		return nil, Err
	} else {
		return &TDbTx{
			Conn: c,
		}, nil
	}
}

func (c *TDbConn) Prepare(query string) (driver.Stmt, error) {
	if TarErrs.Is(PREPARE_ERR) {
		return nil, Err
	}
	if iv, ok := c.TData[query]; ok {
		mv := iv.(map[string]interface{})
		stmt := TDbStmt{}
		stmt.Sql = query
		stmt.Conn = c
		stmt.TData = mv
		if nv, ok := mv["V_COUNTS"]; ok {
			stmt.Numi = int(nv.(float64))
		} else {
			return nil, NotFoundErr(fmt.Sprintf("%s:V_COUNTS", query))
		}
		if nv, ok := mv["V_COLUMN"]; ok {
			stmt.Clmns = strings.Split(nv.(string), ",")
		}
		return &stmt, nil
	} else {
		return nil, NotFoundErr(query)
	}
}

func (c *TDbConn) Close() error {
	if TarErrs.Is(CONN_CLOSE_ERR) {
		return Err
	} else {
		delete(TData, c.Name)
		return nil
	}
}

type TDbTx struct {
	Conn *TDbConn
}

func (tx *TDbTx) Commit() error {
	if TarErrs.Is(COMMIT_ERR) {
		return Err
	} else {
		return nil
	}
}

func (tx *TDbTx) Rollback() error {
	if TarErrs.Is(ROLLBACK_ERR) {
		return Err
	} else {
		return nil
	}
}

type TDbStmt struct {
	Sql   string
	Conn  *TDbConn
	Numi  int
	Clmns []string
	TData map[string]interface{}
}

func (s *TDbStmt) NumInput() int {
	return s.Numi
}

func (s *TDbStmt) Query(args []driver.Value) (driver.Rows, error) {
	if TarErrs.Is(STMT_QUERY_ERR) {
		return nil, Err
	}
	if mv, ok := s.TData[FmtArgs(args)]; ok {
		trow := TDbRows{}
		trow.Args = args
		trow.Stmt = s
		trow.Rows = mv.([]interface{})
		trow.CIdx = 0
		return &trow, nil
	} else {
		return nil, NotFoundErr(FmtArgs(args))
	}
}

func (s *TDbStmt) Exec(args []driver.Value) (driver.Result, error) {
	if TarErrs.Is(STMT_EXEC_ERR) {
		return nil, Err
	}
	if mv, ok := s.TData[FmtArgs(args)]; ok {
		res := mv.(map[string]interface{})
		tres := TDbResult{}
		tres.Args = args
		tres.Stmt = s
		if liid, ok := res["LIID"]; ok {
			tres.LIid = int64(liid.(float64))
		}
		if arow, ok := res["AROW"]; ok {
			tres.ARow = int64(arow.(float64))
		}
		return &tres, nil
	} else {
		return nil, NotFoundErr(FmtArgs(args))
	}
}

func (s *TDbStmt) Close() error {
	if TarErrs.Is(STMT_CLOSE_ERR) {
		return Err
	} else {
		return nil
	}
}

type TDbResult struct {
	Stmt *TDbStmt
	Args []driver.Value
	LIid int64
	ARow int64
}

func (r *TDbResult) LastInsertId() (int64, error) {
	if TarErrs.Is(LAST_INSERT_ID_ERR) {
		return 0, Err
	} else {
		return r.LIid, nil
	}
}

func (r *TDbResult) RowsAffected() (int64, error) {
	if TarErrs.Is(ROWS_AFFECTED_ERR) {
		return 0, Err
	} else {
		return r.ARow, nil
	}
}

type TDbRows struct {
	Stmt *TDbStmt
	Args []driver.Value
	Rows []interface{}
	CIdx int
}

func (rc *TDbRows) Columns() []string {
	return rc.Stmt.Clmns
}

func (rc *TDbRows) Next(dest []driver.Value) error {
	if rc.CIdx >= len(rc.Rows) {
		return io.EOF
	}
	row := rc.Rows[rc.CIdx].(map[string]interface{})
	for i, c := range rc.Columns() {
		if v, ok := row[c]; ok {
			aa := reflect.TypeOf(v)
			switch aa.Kind() {
			case reflect.String:
				dest[i] = []byte(v.(string))
			default:
				dest[i] = v
			}
		} else {
			dest[i] = nil
		}
	}
	rc.CIdx = rc.CIdx + 1
	return nil
}

func (rc *TDbRows) Close() error {
	if TarErrs.Is(ROWS_CLOSE_ERR) {
		return Err
	} else {
		return nil
	}
}

//
func FmtArgs(args []driver.Value) string {
	return fmt.Sprintf("%v", args)
}
