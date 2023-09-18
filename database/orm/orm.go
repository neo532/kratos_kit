package orm

/*
 * @abstract Orm客户端
 * @mail neo532@126.com
 * @date 2023-08-13
 */

import (
	"context"
	"database/sql"
	"sync"
	"time"

	klog "github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	gLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/neo532/gokit/log"
)

var (
	instanceLock sync.Mutex
	ormMap       = make(map[string]*Orm, 2)
)

// ========== Option ==========
type gormOpt struct {
	schema schema.NamingStrategy
}

type Opt func(*Orm)

func WithMaxIdleConns(i int) Opt {
	return func(o *Orm) {
		o.db = append(o.db, func(db *sql.DB) {
			db.SetMaxIdleConns(i)
		})
	}
}
func WithMaxOpenConns(i int) Opt {
	return func(o *Orm) {
		o.db = append(o.db, func(db *sql.DB) {
			db.SetMaxOpenConns(i)
		})
	}
}
func WithConnMaxLifetime(t time.Duration) Opt {
	return func(o *Orm) {
		o.db = append(o.db, func(db *sql.DB) {
			db.SetConnMaxLifetime(t)
		})
	}
}
func WithSlowLog(t time.Duration) Opt {
	return func(o *Orm) {
		o.slowTime = t
	}
}
func WithTablePrefix(s string) Opt {
	return func(o *Orm) {
		o.gormOpt.schema.TablePrefix = s
	}
}
func WithLogger(l klog.Logger) Opt {
	return func(o *Orm) {
		o.logger = log.NewHelper(l)
	}
}

// 使用单数表名，启用该选项后，`User` 表将是`user`
func WithSingularTable() Opt {
	return func(o *Orm) {
		o.gormOpt.schema.SingularTable = true
	}
}
func WithContext(c context.Context) Opt {
	return func(o *Orm) {
		o.bootstrapContext = c
	}
}

// ========== /Option ==========
type Orm struct {
	Orm              *gorm.DB
	Cleanup          func()
	Err              error
	bootstrapContext context.Context

	gormOpt  *gormOpt
	db       []func(db *sql.DB)
	slowTime time.Duration
	logger   *log.Helper
}

func New(name string, dsn gorm.Dialector, opts ...Opt) (db *Orm) {
	instanceLock.Lock()
	defer instanceLock.Unlock()

	var ok bool
	if db, ok = ormMap[name]; ok {
		return
	}

	db = &Orm{
		bootstrapContext: context.Background(),
		gormOpt: &gormOpt{
			schema: schema.NamingStrategy{},
		},
		db: make([]func(db *sql.DB), 0),
	}
	for _, o := range opts {
		o(db)
	}

	db.Orm, db.Err = gorm.Open(
		dsn,
		&gorm.Config{
			Logger:         NewGormLogger(name, db.slowTime, db.logger),
			NamingStrategy: db.gormOpt.schema,
		},
	)
	if db.Err != nil {
		db.logger.
			WithContext(db.bootstrapContext).
			Errorf("Gorm open client[%s] error: %+v",
				name,
				db.Err,
			)
		return
	}

	var sqlDB *sql.DB
	if sqlDB, db.Err = db.Orm.DB(); db.Err != nil {
		db.logger.
			WithContext(db.bootstrapContext).
			Errorf("Orm DB[%s] has error: %+v",
				name,
				db.Err,
			)
		return
	}
	for _, o := range db.db {
		o(sqlDB)
	}

	db.Cleanup = func() {
		if sqlDB == nil {
			db.logger.
				WithContext(db.bootstrapContext).
				Errorf("Close db[%s] is nil!", name)
			return
		}
		if db.Err = sqlDB.Close(); db.Err != nil {
			db.logger.
				WithContext(db.bootstrapContext).
				Errorf("Close db[%s] has error: %+v", name, db.Err)
			return
		}
	}
	ormMap[name] = db
	return
}

type GormLogger struct {
	gorm.Config

	db          string
	slowLogTime time.Duration
	logger      *log.Helper

	LogLevel gLogger.LogLevel
}

func NewGormLogger(db string, slowLogTime time.Duration, logger *log.Helper) *GormLogger {
	return &GormLogger{
		db:          db,
		slowLogTime: slowLogTime,
		logger:      logger,
	}
}

func (g *GormLogger) LogMode(level gLogger.LogLevel) gLogger.Interface {
	g.LogLevel = level
	return g
}

func (g *GormLogger) Info(c context.Context, s string, i ...interface{}) {
	g.logger.WithContext(c).Infof(s, i...)
}

func (g *GormLogger) Warn(c context.Context, s string, i ...interface{}) {
	g.logger.WithContext(c).Warnf(s, i...)
}

func (g *GormLogger) Error(c context.Context, s string, i ...interface{}) {
	g.logger.WithContext(c).Errorf(s, i...)
}

func (g *GormLogger) Trace(c context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rows := fc()
	cost := time.Since(begin)

	if err == gorm.ErrRecordNotFound {
		err = nil
	}

	if err != nil {
		g.logger.
			WithContext(c).
			Errorf("err:[%+v], name:%s, limit:%v, cost:%v, rows:%d, sql:[%s]",
				err,
				g.db,
				g.slowLogTime,
				cost,
				rows,
				sql,
			)
		return
	}

	if cost > g.slowLogTime {
		g.logger.
			WithContext(c).
			Warnf("slowlog, name:%s, limit:%v, cost:%v, rows:%d, sql:[%s]",
				g.db,
				g.slowLogTime,
				cost,
				rows,
				sql,
			)
		return
	}

	g.logger.
		WithContext(c).
		Infof("name:%s, limit:%v, cost:%s, rows:%d, sql:[%s]",
			g.db,
			g.slowLogTime,
			cost,
			rows,
			sql,
		)
}
