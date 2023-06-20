package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/cache"
	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/user"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	tgbotapi "github.com/zijiren233/tg-bot-api/v6"

	"github.com/zijiren233/go-colorlog"
	"gopkg.in/yaml.v3"
)

var avilableDocumentType = []string{"image/jpeg", "image/png"}

func HandleMsg(bot *tgbotapi.BotAPI, Message *tgbotapi.Message) {
	_, err := getCfg(bot, Message)
	if err != nil {
		colorlog.Errorf("Get config err [%s] : %v", Message.From.String(), err)
		return
	}
}

func getCfg(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) (*Config, error) {
	info := new(Config)
	u, err := user.LoadAndInitUser(bot, msg.From.ID)
	if err != nil {
		colorlog.Errorf("Load And Init User Err: %v", err)
		return nil, err
	}
	info.DrawConfig = *u.DefaultConfig()
	if len(msg.Photo) > 0 {
		info.Tag = msg.Caption
		latestPhoto := msg.Photo[len(msg.Photo)-1]
		info.Width = latestPhoto.Width
		info.Height = latestPhoto.Height
		info.Steps = 50
		photo, err := bot.GetFileData(latestPhoto.FileID)
		if err != nil {
			return nil, err
		}
		fi, err := cache.Put(photo)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		}
		info.PrePhotoID = fi.Md5
		err = getConfig(bot, u, info, msg.MessageID)
		if err != nil {
			return nil, err
		}
	} else if msg.Document != nil {
		if _, ok := utils.InString(msg.Document.MimeType, avilableDocumentType); !ok {
			return nil, errors.New("document type is not avilable")
		}
		info.Tag = msg.Caption
		var err error
		var prePhoto []byte
		if prePhoto, err = cache.GetFile(utils.GetFileNamePrefix(msg.Document.FileName)); err == nil {
			info.PrePhotoID = utils.GetFileNamePrefix(msg.Document.FileName)
		} else {
			prePhoto, err = bot.GetFileData(msg.Document.FileID)
			if err != nil {
				colorlog.Error(err)
				return nil, err
			}
			fi, err := cache.Put(prePhoto)
			if err != nil {
				colorlog.Errorf("Put file err: %v", err)
			}
			info.PrePhotoID = fi.Md5
		}
		info.Steps = 50
		info.Width, info.Height, err = utils.GetPhotoSize(prePhoto)
		if err != nil {
			return nil, err
		}
		err = getConfig(bot, u, info, msg.MessageID)
		if err != nil {
			return nil, err
		}
	} else {
		if yaml.Unmarshal([]byte(msg.Text), info) == nil {
			if len(info.PrePhotoID) == 32 {
				photo, err := cache.GetFile(info.PrePhotoID)
				if err != nil {
					return nil, err
				}
				if info.Width == 0 || info.Height == 0 {
					info.Width, info.Height, err = utils.GetPhotoSize(photo)
					if err != nil {
						return nil, err
					}
				}
				err = getConfig(bot, u, info, msg.MessageID)
				if err != nil {
					return nil, err
				}
			} else {
				info.PrePhotoID = ""
				err := getConfig(bot, u, info, msg.MessageID)
				if err != nil {
					return nil, err
				}
			}
			return info, nil
		} else {
			info.Tag = msg.Text
			err := getConfig(bot, u, info, msg.MessageID)
			if err != nil {
				return nil, err
			}
		}
	}
	return info, nil
}

func creatBar(p float64, maxCount int) string {
	if p < 0 {
		p = 0
	} else if p > 1 {
		p = 1
	}
	if maxCount <= 0 {
		maxCount = 50
	}
	s := math.Floor(float64(maxCount) * p)
	return fmt.Sprintf("[%s>%s]", strings.Repeat("=", int(s)), strings.Repeat(" ", maxCount-int(s)))
}

func drawAndSend(bot *tgbotapi.BotAPI, u *user.UserInfo, replyMsgID int, cfg *Config, prePhoto, ControlPhoto []byte, CorrectCfg bool) {
	if u == nil {
		colorlog.Error("user is nil")
		return
	}
	if u.Permissions() == user.T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(bot))
		msg.ReplyMarkup = goJoinButton(u)
		msg.ReplyToMessageID = replyMsgID
		bot.Send(msg)
		return
	}
	if cfg == nil {
		colorlog.Error("cfg is nil")
		return
	}
	nai, err := api.New(&cfg.DrawConfig, prePhoto, ControlPhoto, CorrectCfg)
	if err != nil {
		colorlog.Errorf("New config err [%s] : %v", u.ChatMember.User.String(), err)
		return
	}
	colorlog.Debugf("Draw [%s] : %v", u.ChatMember.User.String(), cfg.DrawConfig)
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	resoult := nai.Draw(ctx, u.Permissions() != user.T_Subscribe)
	msg := tgbotapi.NewMessage(u.ChatMember.User.ID, fmt.Sprintf("%s\n%s", u.LoadLang("generating"), creatBar(nai.Status().Progress, 50)))
	msg.ReplyMarkup = cancelButton(u)
	msg.ReplyToMessageID = replyMsgID
	m, _ := bot.Send(msg)
	cb, err := bot.NewCbk(u.ChatMember.User.ID, u.ChatMember.User.ID, m.MessageID)
	if err != nil {
		return
	}
	timer := time.NewTicker(5 * time.Second)
	defer cb.Close()
	defer bot.Send(tgbotapi.NewDeleteMessage(u.ChatMember.User.ID, m.MessageID))
	defer timer.Stop()
	for {
		select {
		case <-cb.Chan():
			colorlog.Infof("Cancel the draw task [%s]", u.ChatMember.User.String())
			return
		case <-timer.C:
			bot.Send(tgbotapi.NewEditMessageTextAndMarkup(m.Chat.ID, m.MessageID, fmt.Sprintf("%s\n%s", u.LoadLang("generating"), creatBar(nai.Status().Progress, 50)), *cancelButton(u)))
		case b, ok := <-resoult:
			if !ok {
				colorlog.Debug("resoult closed...")
				return
			}
			if b.Err != nil {
				colorlog.Errorf("Draw Err [%s] : %v", u.ChatMember.User.String(), b.Err)
				return
			}
			if u.Permissions() == user.T_Guest {
				u.UseFree(1)
			}
			sendPhoto(b.Resoult, bot, u, replyMsgID, cfg, true, ControlPhoto != nil)
			return
		}
	}
}

func superResolutionRun(bot *tgbotapi.BotAPI, u *user.UserInfo, replyMsgID int, cfg *Config, prePhoto []byte, resize int) {
	if u == nil {
		colorlog.Error("user is nil")
		return
	}
	if prePhoto == nil {
		colorlog.Error("prePhoto is nil")
		return
	}
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	spcfg, err := api.NewSuperResolution([][]byte{prePhoto}, resize)
	if err != nil {
		return
	}
	resoult := spcfg.SuperResolution(ctx, u.Permissions() != user.T_Subscribe)
	colorlog.Debugf("superResolution [%s]", u.ChatMember.User.String())
	msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.LoadLang("generating"))
	msg.ReplyMarkup = cancelButton(u)
	msg.ReplyToMessageID = replyMsgID
	m, _ := bot.Send(msg)
	cb, err := bot.NewCbk(u.ChatMember.User.ID, u.ChatMember.User.ID, m.MessageID)
	if err != nil {
		return
	}
	defer cb.Close()
	defer bot.Send(tgbotapi.NewDeleteMessage(u.ChatMember.User.ID, m.MessageID))
	select {
	case <-cb.Chan():
		colorlog.Infof("Cancel the draw task [%s]", u.ChatMember.User.String())
		return
	case b, ok := <-resoult:
		if !ok {
			return
		}
		if b.Err != nil {
			colorlog.Errorf("Draw Err [%s] : %v", u.ChatMember.User.String(), b.Err)
			return
		}
		if u.Permissions() == user.T_Guest {
			u.UseFree(1)
		}
		cfg.Width *= resize
		cfg.Height *= resize
		sendPhoto(b.Resoult, bot, u, replyMsgID, cfg, false, false)
	}
}

func sendPhoto(b [][]byte, bot *tgbotapi.BotAPI, u *user.UserInfo, replyMsgID int, cfg *Config, button, skipCtrlPhotoSaveToDB bool) {
	if b == nil || u == nil || cfg == nil {
		return
	}
	l := len(b) - 1
	for k, v := range b {
		fi, err := cache.Put(v)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		} else if !(skipCtrlPhotoSaveToDB && k >= (l-1)) {
			if u.Permissions() != user.T_Subscribe {
				db.DB().Create(&db.PhotoInfo{FileID: fi.Md5, UnShare: false, UserID: u.ChatMember.User.ID, Config: cfg.DrawConfig, PrePhotoID: cfg.PrePhotoID, ControlPhotoID: cfg.ControlPhotoID})
			} else {
				db.DB().Create(&db.PhotoInfo{FileID: fi.Md5, UnShare: !u.UserInfo.SharePhoto, UserID: u.ChatMember.User.ID, Config: cfg.DrawConfig, PrePhotoID: cfg.PrePhotoID, ControlPhotoID: cfg.ControlPhotoID})
			}
		}
		msg := tgbotapi.NewDocument(u.ChatMember.User.ID, tgbotapi.FileBytes{Name: fmt.Sprint(fi.Md5, ".png"), Bytes: v})
		msg.ReplyToMessageID = replyMsgID
		msg.AllowSendingWithoutReply = true
		if button {
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("▼", fmt.Sprintf("openImageButton:%v", cfg.Width*cfg.Height <= parseflag.ImgMaxSize)),
			))
			// msg.ReplyMarkup = u.imgButton(cfg.Width*cfg.Height <= parseflag.ImgMaxSize)
		}
		if k == l {
			msg.DisableNotification = false
		} else {
			msg.DisableNotification = true
		}
		if _, err := bot.Send(msg); err != nil {
			colorlog.Errorf("Send File err [%s] : %v", u.ChatMember.User.String(), err)
			continue
		}
		colorlog.Infof("Send File ID [%s] : %v", u.ChatMember.User.String(), fi.Md5)
	}
}
