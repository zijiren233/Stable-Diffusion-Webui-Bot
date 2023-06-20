package db

import (
	"fmt"
	"strings"
	"time"
)

var selectKey = [...]string{"file_id", "pre_photo_id", "tag", "mode", "steps", "seed", "cfg_scale", "width", "height", "model", "uc", "strength", "control_photo_id", "control_preprocess", "control_process"}

type FindConfig struct {
	Deadline      time.Time
	Order         string
	Limit, Offset int64
	User_id       any
	Keywords      []string
	KeywordsRe    []string // Not support sqlite
	SelectKey     []string
}

type MaxCountCfg struct {
	Deadline   time.Time
	User_id    any
	Keywords   []string
	KeywordsRe []string // Not support sqlite
}

func FindImg(cfg FindConfig) ([]PhotoInfo, error) {
	var db = db
	db = db.Where("updated_at <= ?", cfg.Deadline.Format("2006-01-02 15:04:05"))
	if len(cfg.SelectKey) == 0 {
		db = db.Select(selectKey[:])
	} else {
		db = db.Select(cfg.SelectKey)
	}
	db = db.Limit(int(cfg.Limit))
	db = db.Order(cfg.Order)
	if cfg.Offset != 0 {
		db = db.Offset(int(cfg.Offset))
	}
	if cfg.User_id != nil {
		db = db.Where("user_id = ?", cfg.User_id)
	} else {
		db = db.Where("un_share = ?", false)
	}
	for _, v := range cfg.Keywords {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if dbType == T_POSTGRESQL {
			db = db.Where("tag ILIKE ?", fmt.Sprint(`%`, v, `%`))
		} else {
			db = db.Where("tag LIKE ?", fmt.Sprint(`%`, v, `%`))
		}
	}
	if dbType == T_POSTGRESQL {
		for _, v := range cfg.KeywordsRe {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			db = db.Where("tag ~ ?", v)
		}
	}
	var photo = []PhotoInfo{}
	d := db.Find(&photo)
	return photo, d.Error
}

func GetMaxCount(cfg MaxCountCfg) int64 {
	var db = db
	db = db.Where("updated_at <= ?", cfg.Deadline.Format("2006-01-02 15:04:05"))
	db = db.Select("count(id)")
	db = db.Limit(1)
	if cfg.User_id != nil {
		db = db.Where("user_id = ?", cfg.User_id)
	} else {
		db = db.Where("un_share = ?", false)
	}
	db = db.Model(PhotoInfo{})
	for _, v := range cfg.Keywords {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if dbType == T_POSTGRESQL {
			db = db.Where("tag ILIKE ?", fmt.Sprint(`%`, v, `%`))
		} else {
			db = db.Where("tag LIKE ?", fmt.Sprint(`%`, v, `%`))
		}
	}
	if dbType == T_POSTGRESQL {
		for _, v := range cfg.KeywordsRe {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			db = db.Where("tag ~ ?", v)
		}
	}
	var count int64 = 0
	db.Count(&count)
	return count
}
