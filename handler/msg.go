package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	tgbotapi "github.com/zijiren233/tg-bot-api/v6"

	"github.com/zijiren233/go-colorlog"
	"gopkg.in/yaml.v3"
)

var avilableDocumentType = []string{"image/jpeg", "image/png"}

func (h *Handler) sendNewConfig(u *UserInfo, cfg *db.Config, replyMsgId int) (err error) {
	if cfg == nil {
		return errors.New("cfg is nil")
	}
	panel := panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
	h.CorrectCfg(cfg, u, WithTag(), WithUc(), WithSeed())
	mc := tgbotapi.NewMessage(u.ChatMember.User.ID, string(cfg.Fomate2TgHTML()))
	mc.ReplyMarkup = panel
	mc.ReplyToMessageID = replyMsgId
	mc.ParseMode = "HTML"
	mc.DisableWebPagePreview = false
	_, err = h.bot.Send(mc)
	return err
}

func (h *Handler) HandleMsg(Message *tgbotapi.Message) {
	u, err := h.UserHandler.LoadAndInitUser(h.bot, Message.From.ID)
	if err != nil {
		colorlog.Errorf("Load And Init User Err: %v", err)
		return
	}
	info := u.DefaultConfig()
	info, err = h.getCfg(info, Message)
	if err != nil {
		colorlog.Errorf("Get config err [%s] : %v", Message.From.String(), err)
		return
	}
	err = h.sendNewConfig(u, info, Message.MessageID)
	if err != nil {
		colorlog.Errorf("Get config err [%s] : %v", Message.From.String(), err)
	}
}

func (h *Handler) getCfg(info *db.Config, msg *tgbotapi.Message) (*db.Config, error) {
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
	} else if msg.Document != nil {
		if k := utils.In(avilableDocumentType, func(s string) bool {
			return msg.Document.MimeType == s
		}); k == -1 {
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
	} else {
		if yaml.Unmarshal([]byte(msg.Text), info) == nil {
			if info.PrePhotoID != "" {
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
			} else {
				info.PrePhotoID = ""
			}
			return info, nil
		} else {
			info.Tag = msg.Text
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

func (h *Handler) NewDrawConfig(cfg *db.Config, initPhoto, ControlPhoto []byte) *api.Config {
	config := &api.Config{}

	if k := utils.In[gconfig.Model](h.Models, func(m gconfig.Model) bool {
		return m.Name == cfg.Model
	}); k != -1 {
		config.Model = h.Models[k].File
		config.Vae = h.Models[k].Vae
		config.ClipSkip = h.Models[k].ClipSkip
	}

	config.Prompt = cfg.Tag
	config.Seed = cfg.Seed
	config.SamplerName = cfg.Mode
	config.SamplerIndex = cfg.Mode
	config.Width = cfg.Width
	config.Height = cfg.Height
	config.CfgScale = cfg.CfgScale
	config.Steps = cfg.Steps
	config.NegativePrompt = cfg.Uc
	config.Num = 1
	config.Count = cfg.Num

	if len(initPhoto) != 0 {
		config.ResizeMode = 2
		config.InitImages = []string{base64.StdEncoding.EncodeToString(initPhoto)}
		config.DenoisingStrength = cfg.Strength
	} else {
		config.Width /= 2
		config.Height /= 2
		config.EnableHr = true
		config.DenoisingStrength = 0.55
		config.HrScale = 2
		config.HrUpscaler = "R-ESRGAN 4x+ Anime6B"
		if config.Steps < 20 {
			config.HrSecondPassSteps = config.Steps
		} else {
			config.HrSecondPassSteps = 20
		}
	}
	if len(ControlPhoto) != 0 {
		if config.AlwaysonScripts.Controlnet == nil {
			config.AlwaysonScripts.Controlnet = &struct {
				Args []api.ControlnetUnits "json:\"args,omitempty\""
			}{}
		}
		var max int
		if config.Width > config.Height {
			max = config.Width
		} else {
			max = config.Height
		}
		ctrl := api.ControlnetUnits{
			Lowvram:      false,
			InputImage:   base64.StdEncoding.EncodeToString(ControlPhoto),
			Module:       cfg.ControlPreprocess,
			Model:        cfg.ControlProcess,
			ProcessorRes: max,
		}
		config.AlwaysonScripts.Controlnet.Args = append(config.AlwaysonScripts.Controlnet.Args, ctrl)
	}
	return config
}

func (h *Handler) drawAndSend(u *UserInfo, replyMsgID int, cfg *db.Config, prePhoto, ControlPhoto []byte) {
	if u == nil {
		colorlog.Error("user is nil")
		return
	}
	if u.Permissions() == T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(h.bot))
		msg.ReplyMarkup = h.goJoinButton(u)
		msg.ReplyToMessageID = replyMsgID
		h.bot.Send(msg)
		return
	}
	if cfg == nil {
		colorlog.Error("cfg is nil")
		return
	}

	nai, err := h.Api.New(h.NewDrawConfig(cfg, prePhoto, ControlPhoto), prePhoto, ControlPhoto)
	if err != nil {
		colorlog.Errorf("New config err [%s] : %v", u.ChatMember.User.String(), err)
		return
	}
	colorlog.Debugf("Draw [%s] : %v", u.ChatMember.User.String(), cfg)
	ctx, cf := context.WithCancel(context.Background())
	defer cf()
	resoult := nai.Draw(ctx, u.Permissions() != T_Subscribe)
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
			if u.Permissions() == T_Guest {
				u.UseFree(1)
			}
			h.sendPhoto(b.Resoult, u, replyMsgID, cfg, true, ControlPhoto != nil)
			return
		}
	}
}

func (h *Handler) superResolutionRun(u *UserInfo, replyMsgID int, cfg *db.Config, prePhoto []byte, resize int) {
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
	spcfg, err := h.Api.NewSuperResolution([][]byte{prePhoto}, resize)
	if err != nil {
		return
	}
	resoult := spcfg.SuperResolution(ctx, u.Permissions() != T_Subscribe)
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
		if u.Permissions() == T_Guest {
			u.UseFree(1)
		}
		cfg.Width *= resize
		cfg.Height *= resize
		h.sendPhoto(b.Resoult, u, replyMsgID, cfg, false, false)
	}
}

func (h *Handler) sendPhoto(b [][]byte, u *UserInfo, replyMsgID int, cfg *db.Config, button, skipCtrlPhotoSaveToDB bool) {
	if b == nil || u == nil || cfg == nil {
		return
	}
	l := len(b) - 1
	for k, v := range b {
		fi, err := h.cache.Put(v)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		} else if !(skipCtrlPhotoSaveToDB && k >= (l-1)) {
			if u.Permissions() != T_Subscribe {
				h.DB.DB().Create(&db.PhotoInfo{FileID: fi.FileID, UnShare: false, UserID: u.ChatMember.User.ID, Config: *cfg})
			} else {
				h.DB.DB().Create(&db.PhotoInfo{FileID: fi.FileID, UnShare: !u.UserInfo.SharePhoto, UserID: u.ChatMember.User.ID, Config: *cfg})
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
