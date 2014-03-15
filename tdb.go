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
func NotFoundErr(msg interface{}) error {
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

//the error will be return
//usage:Errs=OPEN_ERR|CONN_BEGIN_ERR
var TarErrs TDbErr = 0

//the error will occur in execute count.
//for example,TarErrsC[CONN_BEGIN_ERR]=1,it only error occure when begin twice.
var TarErrsC map[TDbErr]int = map[TDbErr]int{}

//the execute count.
var RTarErrsC map[TDbErr]int = map[TDbErr]int{}

//the execute count.
var RSqlQueryC map[string]int = map[string]int{}
var RSqlExecC map[string]map[string]int = map[string]map[string]int{}

//if error contain target error.
func (t TDbErr) Is(e TDbErr) bool {
	defer func() {
		RTarErrsC[e] = RTarErrsC[e] + 1
	}()
	if (t & e) == e {
		if v1, ok := TarErrsC[e]; ok {
			if v2, ok := RTarErrsC[e]; ok {
				return v1 == v2
			} else {
				return false
			}
		} else {
			return true
		}
	} else {
		return false
	}
}

//set the target error occur on execute count.
func SetErrC(e TDbErr, c int) {
	TarErrsC[e] = c
}

//reset the execute count
func ResetErrsC() {
	RTarErrsC = map[TDbErr]int{}
	RSqlQueryC = map[string]int{}
	RSqlExecC = map[string]map[string]int{}
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
		defer func() {
			RSqlQueryC[query] = RSqlQueryC[query] + 1
		}()
		mv := iv.(map[string]interface{})
		if _, ok := RSqlQueryC[query]; !ok {
			RSqlQueryC[query] = 0
		}
		if ec, ok := mv["ERR_C"]; ok {
			if RSqlQueryC[query] == int(ec.(float64)) {
				return nil, Err
			}
		}
		//
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
	fargs := FmtArgs(args)
	mv, ok := s.TData[fargs]
	if !ok {
		mv, ok = s.TData["*"]
	}
	if ok {
		defer func() {
			RSqlExecC[s.Sql][fargs] = RSqlExecC[s.Sql][fargs] + 1
		}()
		if _, ok := RSqlExecC[s.Sql]; !ok {
			RSqlExecC[s.Sql] = map[string]int{}
		}
		if _, ok := RSqlExecC[s.Sql][fargs]; !ok {
			RSqlExecC[s.Sql][fargs] = 0
		}
		if emv, ok := s.TData["ERR_V"]; ok {
			if ec, ok := emv.(map[string]interface{})[fargs]; ok {
				if RSqlExecC[s.Sql][fargs] == int(ec.(float64)) {
					return nil, Err
				}
			}
		}
		//
		trow := TDbRows{}
		trow.Args = args
		trow.Stmt = s
		trow.Rows = mv.([]interface{})
		trow.CIdx = 0
		return &trow, nil
	} else {
		return nil, NotFoundErr(fargs)
	}
}

func (s *TDbStmt) Exec(args []driver.Value) (driver.Result, error) {
	if TarErrs.Is(STMT_EXEC_ERR) {
		return nil, Err
	}
	fargs := FmtArgs(args)
	mv, ok := s.TData[fargs]
	if !ok {
		mv, ok = s.TData["*"]
	}
	if ok {
		defer func() {
			RSqlExecC[s.Sql][fargs] = RSqlExecC[s.Sql][fargs] + 1
		}()
		if _, ok := RSqlExecC[s.Sql]; !ok {
			RSqlExecC[s.Sql] = map[string]int{}
		}
		if _, ok := RSqlExecC[s.Sql][fargs]; !ok {
			RSqlExecC[s.Sql][fargs] = 0
		}
		if emv, ok := s.TData["ERR_V"]; ok {
			if ec, ok := emv.(map[string]interface{})[fargs]; ok {
				if RSqlExecC[s.Sql][fargs] == int(ec.(float64)) {
					return nil, Err
				}
			}
		}
		//
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
		return nil, NotFoundErr(fargs)
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
	Map2Val(rc.Columns(), row, dest)
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

//the function to covert row to value array.
var Map2Val Map2ValFunc = DefaultMap2Val

type Map2ValFunc func(columns []string, row map[string]interface{}, dest []driver.Value)

//default function to covert row to value array.
func DefaultMap2Val(columns []string, row map[string]interface{}, dest []driver.Value) {
	for i, c := range columns {
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
}
