package handler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

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

func (h *Handler) HandleMsg(Message *tgbotapi.Message) {
	_, err := h.getCfg(Message)
	if err != nil {
		colorlog.Errorf("Get config err [%s] : %v", Message.From.String(), err)
		return
	}
}

func (h *Handler) getCfg(msg *tgbotapi.Message) (*Config, error) {
	info := new(Config)
	u, err := user.LoadAndInitUser(h.bot, msg.From.ID)
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
		photo, err := h.bot.GetFileData(latestPhoto.FileID)
		if err != nil {
			return nil, err
		}
		fi, err := h.cache.Put(photo)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		}
		info.PrePhotoID = fi.FileID
		err = h.getConfig(u, info, msg.MessageID)
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
		if prePhoto, err = h.cache.Get(utils.GetFileNamePrefix(msg.Document.FileName)); err == nil {
			info.PrePhotoID = utils.GetFileNamePrefix(msg.Document.FileName)
		} else {
			prePhoto, err = h.bot.GetFileData(msg.Document.FileID)
			if err != nil {
				colorlog.Error(err)
				return nil, err
			}
			fi, err := h.cache.Put(prePhoto)
			if err != nil {
				colorlog.Errorf("Put file err: %v", err)
			}
			info.PrePhotoID = fi.FileID
		}
		info.Steps = 50
		info.Width, info.Height, err = utils.GetPhotoSize(prePhoto)
		if err != nil {
			return nil, err
		}
		err = h.getConfig(u, info, msg.MessageID)
		if err != nil {
			return nil, err
		}
	} else {
		if yaml.Unmarshal([]byte(msg.Text), info) == nil {
			if len(info.PrePhotoID) == 32 {
				photo, err := h.cache.Get(info.PrePhotoID)
				if err != nil {
					return nil, err
				}
				if info.Width == 0 || info.Height == 0 {
					info.Width, info.Height, err = utils.GetPhotoSize(photo)
					if err != nil {
						return nil, err
					}
				}
				err = h.getConfig(u, info, msg.MessageID)
				if err != nil {
					return nil, err
				}
			} else {
				info.PrePhotoID = ""
				err := h.getConfig(u, info, msg.MessageID)
				if err != nil {
					return nil, err
				}
			}
			return info, nil
		} else {
			info.Tag = msg.Text
			err := h.getConfig(u, info, msg.MessageID)
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

func (h *Handler) drawAndSend(u *user.UserInfo, replyMsgID int, cfg *Config, prePhoto, ControlPhoto []byte, CorrectCfg bool) {
	if u == nil {
		colorlog.Error("user is nil")
		return
	}
	if u.Permissions() == user.T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(h.bot))
		msg.ReplyMarkup = goJoinButton(u)
		msg.ReplyToMessageID = replyMsgID
		h.bot.Send(msg)
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
	m, _ := h.bot.Send(msg)
	cb, err := h.bot.NewCbk(u.ChatMember.User.ID, u.ChatMember.User.ID, m.MessageID)
	if err != nil {
		return
	}
	timer := time.NewTicker(5 * time.Second)
	defer cb.Close()
	defer h.bot.Send(tgbotapi.NewDeleteMessage(u.ChatMember.User.ID, m.MessageID))
	defer timer.Stop()
	for {
		select {
		case <-cb.Chan():
			colorlog.Infof("Cancel the draw task [%s]", u.ChatMember.User.String())
			return
		case <-timer.C:
			h.bot.Send(tgbotapi.NewEditMessageTextAndMarkup(m.Chat.ID, m.MessageID, fmt.Sprintf("%s\n%s", u.LoadLang("generating"), creatBar(nai.Status().Progress, 50)), *cancelButton(u)))
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
			h.sendPhoto(b.Resoult, u, replyMsgID, cfg, true, ControlPhoto != nil)
			return
		}
	}
}

func (h *Handler) superResolutionRun(u *user.UserInfo, replyMsgID int, cfg *Config, prePhoto []byte, resize int) {
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
	m, _ := h.bot.Send(msg)
	cb, err := h.bot.NewCbk(u.ChatMember.User.ID, u.ChatMember.User.ID, m.MessageID)
	if err != nil {
		return
	}
	defer cb.Close()
	defer h.bot.Send(tgbotapi.NewDeleteMessage(u.ChatMember.User.ID, m.MessageID))
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
		h.sendPhoto(b.Resoult, u, replyMsgID, cfg, false, false)
	}
}

func (h *Handler) sendPhoto(b [][]byte, u *user.UserInfo, replyMsgID int, cfg *Config, button, skipCtrlPhotoSaveToDB bool) {
	if b == nil || u == nil || cfg == nil {
		return
	}
	l := len(b) - 1
	for k, v := range b {
		fi, err := h.cache.Put(v)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		} else if !(skipCtrlPhotoSaveToDB && k >= (l-1)) {
			if u.Permissions() != user.T_Subscribe {
				db.DB().Create(&db.PhotoInfo{FileID: fi.FileID, UnShare: false, UserID: u.ChatMember.User.ID, Config: cfg.DrawConfig, PrePhotoID: cfg.PrePhotoID, ControlPhotoID: cfg.ControlPhotoID})
			} else {
				db.DB().Create(&db.PhotoInfo{FileID: fi.FileID, UnShare: !u.UserInfo.SharePhoto, UserID: u.ChatMember.User.ID, Config: cfg.DrawConfig, PrePhotoID: cfg.PrePhotoID, ControlPhotoID: cfg.ControlPhotoID})
			}
		}
		msg := tgbotapi.NewDocument(u.ChatMember.User.ID, tgbotapi.FileBytes{Name: fmt.Sprint(fi.FileID, ".png"), Bytes: v})
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
		if _, err := h.bot.Send(msg); err != nil {
			colorlog.Errorf("Send File err [%s] : %v", u.ChatMember.User.String(), err)
			continue
		}
		colorlog.Infof("Send File ID [%s] : %v", u.ChatMember.User.String(), fi.FileID)
	}
}
