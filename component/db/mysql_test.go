package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransferQuery(t *testing.T) {
	insert_sql := "insert into depart (departName,memo) VALUES(?,?)"
	reuslt_sql := "insert into depart (departName,memo) VALUES(%v,%v)"
	reuslt := transferQuery(insert_sql)
	assert.Equal(t, reuslt_sql, reuslt)
	//t.Logf(reuslt, 1, "a")
}
