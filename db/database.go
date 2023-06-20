package db

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

var dbType dbTYPE

type dbTYPE int

const (
	T_SQLITE dbTYPE = iota
	T_POSTGRESQL
)

func DB() *gorm.DB {
	return db
}

func DBType() dbTYPE {
	return dbType
}

func Init() {
	var dialector gorm.Dialector
	if regexp.MustCompile(`^postgres(ql)?://`).MatchString(parseflag.DSN) ||
		len(strings.Fields(parseflag.DSN)) >= 3 {
		dialector = postgres.New(postgres.Config{
			DSN:                  parseflag.DSN,
			PreferSimpleProtocol: false,
		})
		dbType = T_POSTGRESQL
	} else {
		dialector = sqlite.Open(parseflag.DSN)
		dbType = T_SQLITE
	}
	var err error
	db, err = gorm.Open(dialector, &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Error)},
	)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	d, err := db.DB()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	d.SetMaxIdleConns(8)
	d.SetMaxOpenConns(32)
	d.SetConnMaxLifetime(time.Hour)
	err = db.AutoMigrate(&Subscribe{}, &UserInfo{}, &PhotoInfo{}, &Token{})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
