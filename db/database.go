package db

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	db     *gorm.DB
	dbType dbTYPE
}

type dbTYPE int

const (
	T_SQLITE dbTYPE = iota
	T_POSTGRESQL
)

func (db *DB) DB() *gorm.DB {
	return db.db
}

func (db *DB) DBType() dbTYPE {
	return db.dbType
}

func New(dsn string) *DB {
	tmpDB := &DB{}
	var dialector gorm.Dialector
	if regexp.MustCompile(`^postgres(ql)?://`).MatchString(dsn) ||
		len(strings.Fields(dsn)) >= 3 {
		dialector = postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: false,
		})
		tmpDB.dbType = T_POSTGRESQL
	} else {
		dialector = sqlite.Open(dsn)
		tmpDB.dbType = T_SQLITE
	}
	var err error
	tmpDB.db, err = gorm.Open(dialector, &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Error)},
	)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	d, err := tmpDB.db.DB()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	d.SetMaxIdleConns(8)
	d.SetMaxOpenConns(32)
	d.SetConnMaxLifetime(time.Hour)
	err = tmpDB.db.AutoMigrate(&Subscribe{}, &UserInfo{}, &PhotoInfo{}, &Token{})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return tmpDB
}
