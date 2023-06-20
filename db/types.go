package db

import (
	"time"

	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"

	"gorm.io/gorm"
)

type UserInfo struct {
	UserID     int64  `gorm:"unique;not null;primary_key"`
	Passwd     string `gorm:"not null"`
	Language   string `gorm:"default:en_us"`
	SharePhoto bool   `gorm:"default:true"`
	UserDefaultCfg
}

type UserDefaultCfg struct {
	UserDefaultUC     string
	UserDefaultMODE   string
	UserDefaultNumber int
	UserDefaultScale  int
	UserDefaultSteps  int
}

type PhotoInfo struct {
	gorm.Model
	FileID         string         `gorm:"unique;not null"`
	UserID         int64          `gorm:"not null"`
	UnShare        bool           `gorm:"default:false"`
	Config         api.DrawConfig `gorm:"embedded"`
	PrePhotoID     string
	ControlPhotoID string
}

type Subscribe struct {
	gorm.Model
	UserID     int64 `gorm:"unique;not null"`
	FreeAmount int   `gorm:"not null"`
	Deadline   time.Time
}

type Token struct {
	Token     string `gorm:"unique;not null"`
	ValidDate uint64 `gorm:"not null"`
}
