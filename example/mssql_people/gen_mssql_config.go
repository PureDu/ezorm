package test

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/ezbuy/ezorm/db"
	"github.com/jmoiron/sqlx"
)

var (
	_sqlServer     *db.SqlServer
	_db            *sql.DB
	_queryWrappers []db.QueryWrapper
)

func MssqlSetUp(dataSourceName string) {
	// Use commandline args as default app name
	if !strings.Contains(dataSourceName, "app name") {
		var commandArgs []string
		// Add all commandline args until options with "-" prefix
		for _, each := range os.Args {
			if strings.HasPrefix(each, "-") {
				break
			}

			_, filename := path.Split(each)

			commandArgs = append(commandArgs, filename)
		}

		dataSourceName = fmt.Sprintf("%s;app name=%s",
			strings.TrimRight(dataSourceName, " ;"),
			strings.Join(commandArgs, " "))
	}

	conn, err := sqlx.Connect("mssql", dataSourceName)
	if err != nil {
		panic(fmt.Sprintf("[db.GetSqlServer] open sql fail:%s", err.Error()))
	}

	_sqlServer = &db.SqlServer{DB: conn}
	_db = conn.DB
}

func MssqlSetMaxOpenConns(maxOpenConns int) {
	_sqlServer.SetMaxOpenConns(maxOpenConns)
}

func MssqlSetMaxIdleConns(maxIdleConns int) {
	_sqlServer.SetMaxIdleConns(maxIdleConns)
}

func MssqlAddQueryWrapper(r db.QueryWrapper) {
	_sqlServer.AddQueryWrapper(r)
	_queryWrappers = append(_queryWrappers, r)
}

func MssqlClose() error {
	return _sqlServer.Close()
}

func mssqlUnwrappedQuery(query string, args ...interface{}) (interface{}, error) {
	rows, err := _db.Query(query, args...)
	return rows, err
}

func mssqlUnwrappedExec(query string, args ...interface{}) (interface{}, error) {
	result, err := _db.Exec(query, args...)
	return result, err
}

func mssqlQuery(query string, args ...interface{}) (*sql.Rows, error) {
	if len(_queryWrappers) == 0 {
		return _db.Query(query, args...)
	}

	queryer := mssqlUnwrappedQuery

	for _, r := range _queryWrappers {
		queryer = r(queryer, query, args...)
	}

	rowsItf, err := queryer(query, args...)
	if err != nil {
		return nil, err
	}
	rows := rowsItf.(*sql.Rows)
	return rows, err
}

func mssqlExec(query string, args ...interface{}) (sql.Result, error) {
	if len(_queryWrappers) == 0 {
		return _db.Exec(query, args...)
	}

	queryer := mssqlUnwrappedExec

	for _, r := range _queryWrappers {
		queryer = r(queryer, query, args...)
	}

	resultItf, err := queryer(query, args...)
	if err != nil {
		return nil, err
	}
	result := resultItf.(sql.Result)
	return result, err
}
