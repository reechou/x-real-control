package utils

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

var (
	DefaultOpenConns = 200
	DefaultIdleConns = 100
	DefaultUser      = "root"
)

var (
	ErrMysqlNoHost   = errors.New("Mysql has no host.")
	ErrMysqlNoDBName = errors.New("Mysql has no dbname.")
	ErrMysqlNotInit  = errors.New("Mysql not init.")
)

const (
	SQL_INSERT = "INSERT INTO %s (%s) VALUES (%s)"
	SQL_UPDATE = "UPDATE %s SET %s WHERE %s"
	SQL_DELETE = "DELETE FROM %s WHERE %s"
)

type MysqlInfo struct {
	Host         string
	User         string
	Pass         string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

type MysqlController struct {
	db           *sql.DB
	maxOpenConns int
	maxIdleConns int
}

func NewMysqlController() *MysqlController {
	return &MysqlController{maxOpenConns: DefaultOpenConns, maxIdleConns: DefaultIdleConns}
}

func (mc *MysqlController) InitMysql(info *MysqlInfo) error {
	if info.User == "" {
		info.User = DefaultUser
	}
	if info.Host == "" {
		return ErrMysqlNoHost
	}
	if info.DBName == "" {
		return ErrMysqlNoDBName
	}
	if info.MaxOpenConns == 0 {
		info.MaxOpenConns = DefaultOpenConns
	}
	if info.MaxIdleConns == 0 {
		info.MaxIdleConns = DefaultIdleConns
	}

	mc.maxOpenConns = info.MaxOpenConns
	mc.maxIdleConns = info.MaxIdleConns
	dbSourceName := info.User + ":" + info.Pass + "@tcp(" + info.Host + ")/" + info.DBName + "?charset=utf8&interpolateParams=true"
	mc.db, _ = sql.Open("mysql", dbSourceName)
	mc.db.SetMaxOpenConns(mc.maxOpenConns)
	mc.db.SetMaxIdleConns(mc.maxIdleConns)
	return mc.db.Ping()
}

func (mc *MysqlController) Close() {
	if mc.db != nil {
		mc.db.Close()
	}
}

func (mc *MysqlController) checkDB() bool {
	return mc.db != nil
}

// insert
func (mc *MysqlController) Insert(sqlstr string, args ...interface{}) (int64, error) {
	if !mc.checkDB() {
		return 0, ErrMysqlNotInit
	}

	result, err := mc.db.Exec(sqlstr, args...)
	if err != nil {
		return 0, err
	}

	//stmtIns, err := mc.db.Prepare(sqlstr)
	//if err != nil {
	//	return 0, err
	//}
	//defer stmtIns.Close()
	//
	//result, err := stmtIns.Exec(args...)
	//if err != nil {
	//	return 0, err
	//}

	return result.LastInsertId()
}

// modify or delete
func (mc *MysqlController) Exec(sqlstr string, args ...interface{}) (int64, error) {
	if !mc.checkDB() {
		return 0, ErrMysqlNotInit
	}

	result, err := mc.db.Exec(sqlstr, args...)
	if err != nil {
		return 0, err
	}

	//stmtIns, err := mc.db.Prepare(sqlstr)
	//if err != nil {
	//	return 0, err
	//}
	//defer stmtIns.Close()
	//
	//result, err := stmtIns.Exec(args...)
	//if err != nil {
	//	return 0, err
	//}

	return result.RowsAffected()
}

// query, val type: string
func (mc *MysqlController) FetchRow(sqlstr string, args ...interface{}) (*map[string]string, error) {
	if !mc.checkDB() {
		return nil, ErrMysqlNotInit
	}

	rows, err := mc.db.Query(sqlstr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//stmtOut, err := mc.db.Prepare(sqlstr)
	//if err != nil {
	//	return nil, err
	//}
	//defer stmtOut.Close()
	//
	//rows, err := stmtOut.Query(args...)
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	ret := make(map[string]string, len(scanArgs))

	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		var value string

		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			ret[columns[i]] = value
		}
		break //get the first row only
	}
	return &ret, nil
}

func (mc *MysqlController) FetchRows(sqlstr string, args ...interface{}) (*[]map[string]string, error) {
	if !mc.checkDB() {
		return nil, ErrMysqlNotInit
	}

	rows, err := mc.db.Query(sqlstr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//stmtOut, err := mc.db.Prepare(sqlstr)
	//if err != nil {
	//	return nil, err
	//}
	//defer stmtOut.Close()
	//
	//rows, err := stmtOut.Query(args...)
	//if err != nil {
	//	return nil, err
	//}
	//defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	ret := make([]map[string]string, 0)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		var value string
		vmap := make(map[string]string, len(scanArgs))
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			vmap[columns[i]] = value
		}
		ret = append(ret, vmap)
	}
	return &ret, nil
}
