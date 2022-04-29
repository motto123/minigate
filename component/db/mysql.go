package db

import (
	"com.minigame.component/log"
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"strings"
)

var mysqlTag = "mysql"

type MysqlDb struct {
	db      *sql.DB
	IsPrint bool
}

func NewMysqlClient(user, password, ip string, port int, dbName string) (db *MysqlDb, err error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", user, password, ip, port, dbName)
	var sqlDb *sql.DB
	sqlDb, err = sql.Open("mysql", dataSourceName)
	err = errors.WithStack(err)
	if err != nil {
		//fmt.Println("数据库链接错误", err)
		return
	}
	err = sqlDb.Ping()
	err = errors.WithStack(err)
	if err != nil {
		return
	}
	//sqlDb.SetMaxOpenConns()
	//sqlDb.SetMaxIdleConns()
	return &MysqlDb{db: sqlDb}, err
}

func (m *MysqlDb) Query(query string, args ...interface{}) (*sql.Rows, error) {
	if m.IsPrint {
		log.Debugf(mysqlTag, transferQuery(query), args...)
	}
	rows, err := m.db.Query(query, args...)
	err = errors.WithStack(err)
	return rows, err
}
func (m *MysqlDb) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.IsPrint {
		log.Debugf(mysqlTag, transferQuery(query), args...)
	}
	rows, err := m.db.QueryContext(ctx, query, args...)
	err = errors.WithStack(err)
	return rows, err
}
func (m *MysqlDb) QueryRow(query string, args ...interface{}) *sql.Row {
	if m.IsPrint {
		log.Debugf(mysqlTag, transferQuery(query), args...)
	}
	return m.db.QueryRow(query, args...)
}
func (m *MysqlDb) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.IsPrint {
		log.Debugf(mysqlTag, transferQuery(query), args...)
	}
	return m.db.QueryRowContext(ctx, query, args...)
}

func (m *MysqlDb) Exec(query string, args ...interface{}) (sql.Result, error) {
	if m.IsPrint {
		log.Debugf(mysqlTag, transferQuery(query), args...)
	}
	result, err := m.db.Exec(query, args...)
	err = errors.WithStack(err)
	return result, err
}

func (m *MysqlDb) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.IsPrint {
		log.Debugf(mysqlTag, transferQuery(query), args...)
	}
	result, err := m.db.ExecContext(ctx, query, args...)
	err = errors.WithStack(err)
	return result, err
}

func (m *MysqlDb) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if m.IsPrint {
		log.Debugf(mysqlTag, query, "")
	}
	stmt, err := m.db.PrepareContext(ctx, query)
	err = errors.WithStack(err)
	return stmt, err
}

//TODO: 事务函数的封装

func transferQuery(query string) string {
	return strings.Replace(query, "?", "%v", -1)
}
