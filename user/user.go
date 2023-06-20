package user

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	"github.com/zijiren233/stable-diffusion-webui-bot/i18n"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"github.com/zijiren233/go-colorlog"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	tgbotapi "github.com/zijiren233/tg-bot-api/v6"

	"github.com/im7mortal/kmutex"
)

type permissions uint

const (
	T_Prohibit permissions = iota
	T_Guest
	T_Subscribe
)

type UserInfo struct {
	UserInfo   *db.UserInfo
	Subscribe  *db.Subscribe
	ChatMember *tgbotapi.ChatMember
	LastUpdate time.Time
}

func (u *UserInfo) Permissions() permissions {
	now := time.Now()
	if u.Subscribe.Deadline.After(now) {
		return T_Subscribe
	}
	addTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if u.Subscribe.UpdatedAt.Before(addTime) && u.Subscribe.FreeAmount < parseflag.MaxFree {
		u.Subscribe.FreeAmount = parseflag.MaxFree
	}
	if u.Subscribe.FreeAmount > 0 {
		return T_Guest
	} else {
		return T_Prohibit
	}
}

func (u *UserInfo) LoadLang(tag string) string {
	return i18n.LoadLang(u.UserInfo.Language, tag)
}

func (u *UserInfo) LoadExtraLang(tag string) string {
	return i18n.LoadExtraLang(u.UserInfo.Language, tag)
}

func (u *UserInfo) SetLang(langType string) error {
	u.UserInfo.Language = langType
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("language", langType).Error
}

// func (u *userInfo) chatMemberRemove() {
// 	userCache.Delete(u.UserInfo.UserID)
// }

func (u *UserInfo) UseFree(n int) {
	now := time.Now()
	addTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if u.Subscribe.UpdatedAt.Before(addTime) && u.Subscribe.FreeAmount < parseflag.MaxFree {
		u.Subscribe.FreeAmount = parseflag.MaxFree
	}
	if u.Subscribe.FreeAmount-n >= 0 {
		u.Subscribe.FreeAmount -= n
	} else {
		u.Subscribe.FreeAmount = 0
	}
	u.Subscribe.UpdatedAt = time.Now()
	db.DB().Model(db.Subscribe{}).Where(`user_id = ?`, u.Subscribe.UserID).Update("free_amount", u.Subscribe.FreeAmount)
}

var userCache = &sync.Map{}

var userL = kmutex.New()

func LoadAndInitUser(bot *tgbotapi.BotAPI, userID int64) (u *UserInfo, err error) {
	userL.Lock(userID)
	defer userL.Unlock(userID)
	i, ok := userCache.Load(userID)
	now := time.Now()
	if !ok {
		u = new(UserInfo)
		cm, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: userID, UserID: userID}})
		if err != nil {
			return nil, err
		}
		u.ChatMember = &cm
		u.UserInfo = &db.UserInfo{}
		db.DB().Where("user_id = ?", userID).Attrs(db.UserInfo{UserID: userID, Language: "en_us", Passwd: utils.RandomString(10), SharePhoto: true}).FirstOrCreate(u.UserInfo)
		u.Subscribe = &db.Subscribe{}
		db.DB().Where("user_id = ?", userID).Attrs(db.Subscribe{UserID: userID, FreeAmount: parseflag.MaxFree, Deadline: now}).FirstOrCreate(u.Subscribe)
		userCache.Store(userID, u)
	} else {
		u = i.(*UserInfo)
	}
	return
}

func LoadUser(bot *tgbotapi.BotAPI, userID int64) (u *UserInfo, err error) {
	userL.Lock(userID)
	defer userL.Unlock(userID)
	i, ok := userCache.Load(userID)
	now := time.Now()
	if !ok {
		u = new(UserInfo)
		u.UserInfo, err = findUser(userID)
		if err != nil {
			return nil, err
		}
		cm, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{ChatConfigWithUser: tgbotapi.ChatConfigWithUser{ChatID: userID, UserID: userID}})
		if err != nil {
			return nil, err
		}
		u.ChatMember = &cm
		db.DB().Where("user_id = ?", userID).Attrs(db.UserInfo{UserID: userID, Language: "en_us", Passwd: utils.RandomString(10), SharePhoto: true}).FirstOrCreate(u.UserInfo)
		u.Subscribe = &db.Subscribe{}
		db.DB().Where("user_id = ?", userID).Attrs(db.Subscribe{UserID: userID, FreeAmount: parseflag.MaxFree, Deadline: now}).FirstOrCreate(u.Subscribe)
		userCache.Store(userID, u)
	} else {
		u = i.(*UserInfo)
	}
	return
}

var errUserNotFind = errors.New("user not found")

// func FindUser(userID int64) (*db.UserInfo, error) {
// 	userL.Lock(userID)
// 	defer userL.Unlock(userID)
// 	return findUser(userID)
// }

func findUser(userID int64) (*db.UserInfo, error) {
	i, ok := userCache.Load(userID)
	if !ok {
		u := new(db.UserInfo)
		db.DB().Where("user_id = ?", userID).Find(u)
		if u.UserID != 0 {
			userCache.Store(userID, u)
			return u, nil
		} else {
			return u, errUserNotFind
		}
	}
	return i.(*db.UserInfo), nil
}

// userID -> passwd
func (u *UserInfo) Passwd() string {
	return u.UserInfo.Passwd
}

func (u *UserInfo) ChangeShare(share bool) error {
	if u.UserInfo.SharePhoto == share {
		return nil
	}
	u.UserInfo.SharePhoto = share
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("share_photo", share).Error
}

func (u *UserInfo) ChangeDefaultUC(uc string) error {
	if u.UserInfo.UserDefaultUC == uc {
		return nil
	}
	uc = api.ReplaceString(uc)
	if len(uc) > 2048 {
		return errors.New("us is very long")
	}
	u.UserInfo.UserDefaultUC = uc
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("user_default_uc", uc).Error
}

func (u *UserInfo) ChangeDefaultMODE(mode string) error {
	if u.UserInfo.UserDefaultMODE == mode {
		return nil
	}
	m := api.AllMode()
	if _, ok := utils.InString(mode, m[:]); !ok {
		return errors.New("mode not support")
	}
	u.UserInfo.UserDefaultMODE = mode
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("user_default_mode", mode).Error
}

func (u *UserInfo) ChangeDefaultNumber(num int) error {
	if u.UserInfo.UserDefaultNumber == num {
		return nil
	}
	u.UserInfo.UserDefaultNumber = api.ParseNum(num)
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("user_default_number", num).Error
}

func (u *UserInfo) ChangeDefaultScale(scale int) error {
	if u.UserInfo.UserDefaultScale == scale {
		return nil
	}
	u.UserInfo.UserDefaultScale = api.ParseCfgScalse(scale)
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("user_default_scale", u.UserInfo.UserDefaultScale).Error
}

func (u *UserInfo) ChangeDefaultSteps(steps int) error {
	if u.UserInfo.UserDefaultSteps == steps {
		return nil
	}
	u.UserInfo.UserDefaultSteps = api.ParseSteps(steps)
	return db.DB().Model(db.UserInfo{}).Where("user_id = ?", u.ChatMember.User.ID).Update("user_default_steps", u.UserInfo.UserDefaultSteps).Error
}

func (u *UserInfo) DefaultConfig() *api.DrawConfig {
	cfg := api.DefaultConfig()
	if u.UserInfo.UserDefaultMODE != "" {
		cfg.Mode = u.UserInfo.UserDefaultMODE
	}
	if u.UserInfo.UserDefaultUC != "" {
		cfg.Uc = u.UserInfo.UserDefaultUC
	}
	if u.UserInfo.UserDefaultNumber != 0 {
		cfg.Num = u.UserInfo.UserDefaultNumber
	}
	if u.UserInfo.UserDefaultScale != 0 {
		cfg.CfgScale = u.UserInfo.UserDefaultScale
	}
	if u.UserInfo.UserDefaultSteps != 0 {
		cfg.Steps = u.UserInfo.UserDefaultSteps
	}
	return cfg
}

func (u *UserInfo) ProhibitString(bot *tgbotapi.BotAPI) string {
	t := time.Now()
	buffer := bytes.NewBufferString(strings.ReplaceAll(u.LoadLang("prohibit"), "{{.time}}", utils.TimeFomate(time.Until(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Add(time.Hour*24)))))
	rawUrl, err := url.Parse(fmt.Sprintf("https://t.me/%s", bot.Self.UserName))
	if err != nil {
		colorlog.Error(err)
		return buffer.String()
	}
	query := rawUrl.Query()
	query.Set("start", fmt.Sprintf("invite-%d", u.UserInfo.UserID))
	rawUrl.RawQuery = query.Encode()
	buffer.WriteString(fmt.Sprintf("\n%s\n%s", u.LoadLang("invite"), rawUrl.String()))
	return buffer.String()
}
