package orm

/*
 * @abstract Orm客户端
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"

	"gorm.io/gorm"

	"github.com/neo532/gokit/middleware/tracing"
)

type contextTransactionKey struct{}

type Orms struct {
	read        *gorm.DB
	write       *gorm.DB
	shadowRead  *gorm.DB
	shadowWrite *gorm.DB

	cleanupFuncs []func()
	Err          error
}

func News(read, write *Orm) (dbs *Orms) {
	dbs = &Orms{}
	dbs.read = dbs.setDB(read)
	dbs.write = dbs.setDB(write)
	return
}

func (m *Orms) SetShadow(read, write *Orm) *Orms {
	m.shadowRead = m.setDB(read)
	m.shadowWrite = m.setDB(write)
	return m
}

func (m *Orms) Read(c context.Context) (db *gorm.DB) {
	if tx, ok := c.Value(contextTransactionKey{}).(*gorm.DB); ok {
		return tx
	}
	if tracing.IsBenchmark(c) {
		return m.shadowRead
	}
	return m.read
}

func (m *Orms) Write(c context.Context) (db *gorm.DB) {
	if tx, ok := c.Value(contextTransactionKey{}).(*gorm.DB); ok {
		return tx
	}
	if tracing.IsBenchmark(c) {
		return m.shadowWrite
	}
	return m.write
}

func (m *Orms) Transaction(c context.Context, fn func(c context.Context) error) error {
	return m.Write(c).WithContext(c).Transaction(func(tx *gorm.DB) error {
		c = context.WithValue(c, contextTransactionKey{}, tx)
		return fn(c)
	})
}

func (m *Orms) Cleanup() func() {
	return func() {
		for _, fn := range m.cleanupFuncs {
			fn()
		}
	}
}

func (m *Orms) setDB(db *Orm) *gorm.DB {
	m.cleanupFuncs = append(m.cleanupFuncs, db.Cleanup)
	if db.Err != nil {
		m.Err = db.Err
	}
	return db.Orm
}
