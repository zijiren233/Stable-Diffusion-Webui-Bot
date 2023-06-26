package handler

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/cache"
	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	tgbotapi "github.com/zijiren233/tg-bot-api/v6"

	"github.com/zijiren233/go-colorlog"
	"gopkg.in/yaml.v3"
)

func (h *Handler) returnCallback(CallbackQuery *tgbotapi.CallbackQuery) {
	h.bot.Request(tgbotapi.NewCallback(CallbackQuery.ID, ""))
}

func (h *Handler) HandleCallback(CallbackQuery *tgbotapi.CallbackQuery) {
	defer h.returnCallback(CallbackQuery)
	data := tgbotapi.ParseCbkData(CallbackQuery)
	if ch, ok := h.bot.FindCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID, CallbackQuery.Message.MessageID); ok {
		select {
		case ch.Chan() <- data:
		default:
		}
		return
	}
	u, err := h.UserHandler.LoadAndInitUser(h.bot, CallbackQuery.From.ID)
	if err != nil {
		colorlog.Errorf("Load And Init User Err: %v", err)
		return
	}
	switch data.Key {
	default:
		return
	case "panel":
		h.panel(CallbackQuery, data, u)
	case "default":
		h.defaultPanel(CallbackQuery, data, u)
	case "setDft":
		h.setDft(CallbackQuery, data, u)
	case "share":
		h.shareCbk(CallbackQuery, data, u)
	case "setCfg":
		h.setCfg(CallbackQuery, data, u)
	case "reDraw":
		h.reDraw(CallbackQuery, u)
	case "editCfg":
		h.editCfg(CallbackQuery, data, u)
	case "editImg":
		h.editImg(CallbackQuery, u)
	case "fineTune":
		h.fineTune(CallbackQuery, data.Value, u)
	case "openImageButton":
		h.openImageButton(CallbackQuery, data, u)
	case "spr":
		h.superResolution(CallbackQuery, data, u)
	case "delete":
		h.bot.Send(tgbotapi.NewDeleteMessage(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID))
	case "cmd-spr":
		h.bot.Send(tgbotapi.NewDeleteMessage(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID))
	case "lang":
		h.lang(CallbackQuery, data, u)
	case "helpLang":
		h.helpLang(CallbackQuery, data, u)
	case "pool":
		h.pool(CallbackQuery, data, u)
	}
}

func (h *Handler) shareCbk(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	if u.Permissions() != T_Subscribe {
		msg := tgbotapi.NewMessage(CallbackQuery.Message.Chat.ID, fmt.Sprintf("%s\n%s", u.LoadLang("shareInfo"), u.LoadLang("mustShare")))
		msg.ReplyToMessageID = CallbackQuery.Message.MessageID
		msg.ReplyMarkup = h.goJoinButton(u)
		h.bot.Send(msg)
		return
	}
	i, err := strconv.ParseInt(data.Value, 10, 64)
	if err != nil {
		return
	}
	if i == 0 {
		u.ChangeShare(false)
	} else {
		u.ChangeShare(true)
	}
	var option string
	if u.UserInfo.SharePhoto {
		option = u.LoadLang("enable")
	} else {
		option = u.LoadLang("disable")
	}
	msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s: %s", u.LoadLang("shareInfo"), option))
	msg.ReplyMarkup = gShareButton(u)
	h.bot.Send(msg)
}

func (h *Handler) editCfg(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(CallbackQuery.Message.Text), cfg)
	if err != nil {
		colorlog.Error(err)
		return
	}
	h.getConfig(u, cfg, CallbackQuery.Message.MessageID)
}

func (h *Handler) pool(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	if u.UserInfo.UserID != parseflag.OwnerID {
		return
	}
	msg := tgbotapi.NewEditMessageTextAndMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("```\npool: %d\nfree: %d\nwait: %d\ntime: %v\n```", h.Api.DrawPoolCap(), h.Api.DrawFree(), h.Api.DrawWait(), time.Now().Format("01-02 15:04:05")), poolButton)
	msg.ParseMode = "Markdown"
	h.bot.Send(msg)
}

func (h *Handler) helpLang(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	if err := u.SetLang(data.Value); err != nil {
		colorlog.Errorf("Set Language err [%s] : %v", CallbackQuery.From.String(), err)
	} else {
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n\nYou Can Use Website\nUser ID: `%d`\nPassword: `%s`", u.LoadLang("help"), u.UserInfo.UserID, u.Passwd()))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: append(helpLangButton.InlineKeyboard, clictUrlButton(u, fmt.Sprintf("%s://%s/login", parseflag.ApiScheme, parseflag.ApiHost)).InlineKeyboard...)}
		h.bot.Send(msg)
	}
}

func (h *Handler) lang(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	if err := u.SetLang(data.Value); err != nil {
		colorlog.Errorf("Set Language err [%s] : %v", CallbackQuery.From.String(), err)
	} else {
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, u.LoadLang("setLangSuccess"))
		msg.ReplyMarkup = langButton
		h.bot.Send(msg)
	}
}

func (h *Handler) openImageButton(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	spr, err := strconv.ParseBool(data.Value)
	if err != nil {
		return
	}
	msg := tgbotapi.NewEditMessageReplyMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, *imgButton(u, spr))
	h.bot.Send(msg)
}

func (h *Handler) defaultPanel(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	switch data.Value {
	case "panel":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, u.LoadLang("setDft"))
		msg.ReplyMarkup = setDefaultCfg(u)
		h.bot.Send(msg)
	case "mode":
		var msg tgbotapi.EditMessageTextConfig
		if u.UserInfo.UserDefaultMODE == "" {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprint(u.LoadLang("setDft"), " ", u.LoadLang("mode")))
			msg.ReplyMarkup = h.generateSetDftMODEButton(u)
		} else {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprint(u.LoadLang("setDft"), " ", u.LoadLang("mode")))
			msg.ReplyMarkup = h.generateSetDftMODEButton(u)
		}
		h.bot.Send(msg)
	case "uc":
		var msg tgbotapi.EditMessageTextConfig
		if u.UserInfo.UserDefaultUC == "" {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n```\n%s\n```", u.LoadLang("setDft"), u.LoadLang("unwanted"), api.DefauleUC()))
		} else {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n```\n%s\n```", u.LoadLang("setDft"), u.LoadLang("unwanted"), u.UserInfo.UserDefaultUC))
		}
		msg.ReplyMarkup = generateSetDftUCButton(u)
		msg.ParseMode = "Markdown"
		h.bot.Send(msg)
	case "number":
		var msg tgbotapi.EditMessageTextConfig
		if u.UserInfo.UserDefaultNumber == 0 {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprint(u.LoadLang("setDft"), " ", u.LoadLang("number")))
		} else {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprint(u.LoadLang("setDft"), " ", u.LoadLang("number")))
		}
		msg.ReplyMarkup = h.generateSetDftNumberButton(u)
		h.bot.Send(msg)
	case "scale":
		var msg tgbotapi.EditMessageTextConfig
		if u.UserInfo.UserDefaultScale == 0 {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n%d", u.LoadLang("setDft"), u.LoadLang("scale"), h.DefaultCfgScale))
		} else {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n%d", u.LoadLang("setDft"), u.LoadLang("scale"), u.UserInfo.UserDefaultScale))
		}
		msg.ReplyMarkup = generateSetDftScaleButton(u)
		h.bot.Send(msg)
	case "steps":
		var msg tgbotapi.EditMessageTextConfig
		if u.UserInfo.UserDefaultSteps == 0 {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n%d", u.LoadLang("setDft"), u.LoadLang("steps"), h.DefaultSteps))
		} else {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n%d", u.LoadLang("setDft"), u.LoadLang("steps"), u.UserInfo.UserDefaultSteps))
		}
		msg.ReplyMarkup = generateSetDftStepsButton(u)
		h.bot.Send(msg)
	}
}

func (h *Handler) setDft(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	before, after, found := strings.Cut(data.Value, ":")
	data.Key = before
	if found {
		data.Value = after
	} else {
		data.Value = before
	}
	switch data.Key {
	case "mode":
		u.ChangeDefaultMODE(data.Value)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprint(u.LoadLang("setDft"), " ", u.LoadLang("mode")))
		msg.ReplyMarkup = h.generateSetDftMODEButton(u)
		h.bot.Send(msg)
	case "uc":
		if data.Value == "reset" {
			u.ChangeDefaultUC("")
		} else {
			m, err := h.bot.NewMsgCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID)
			if err != nil {
				return
			}
			defer m.Close()
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, u.LoadLang("sendTag"))
			msg.ReplyMarkup = cancelButton(u)
			h.bot.Send(msg)
			c, err := h.bot.NewCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID, CallbackQuery.Message.MessageID)
			if err != nil {
				colorlog.Error(err)
				return
			}
			defer c.Close()
			t := time.NewTimer(time.Minute * 3)
			defer t.Stop()
			select {
			case <-t.C:
			case <-c.Chan():
			case msg, ok := <-m.MsgChan():
				if ok {
					u.ChangeDefaultUC(msg.Text)
				}
			}
		}
		var msg tgbotapi.EditMessageTextConfig
		if u.UserInfo.UserDefaultUC == "" {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n```\n%s\n```", u.LoadLang("setDft"), u.LoadLang("unwanted"), api.DefauleUC()))
		} else {
			msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n```\n%s\n```", u.LoadLang("setDft"), u.LoadLang("unwanted"), u.UserInfo.UserDefaultUC))
		}
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = generateSetDftUCButton(u)
		h.bot.Send(msg)
	case "number":
		u2, err := strconv.ParseUint(data.Value, 10, 64)
		if err != nil {
			colorlog.Error(err)
		}
		u.ChangeDefaultNumber(int(u2))
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprint(u.LoadLang("setDft"), " ", u.LoadLang("number")))
		msg.ReplyMarkup = h.generateSetDftNumberButton(u)
		h.bot.Send(msg)
	case "scale":
		if u.UserInfo.UserDefaultScale == 0 {
			u.UserInfo.UserDefaultScale = h.DefaultCfgScale
		}
		switch data.Value {
		case "+":
			u.ChangeDefaultScale(u.UserInfo.UserDefaultScale + 1)
		case "-":
			u.ChangeDefaultScale(u.UserInfo.UserDefaultScale - 1)
		}
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n%d", u.LoadLang("setDft"), u.LoadLang("scale"), u.UserInfo.UserDefaultScale))
		msg.ReplyMarkup = generateSetDftScaleButton(u)
		h.bot.Send(msg)
	case "steps":
		if u.UserInfo.UserDefaultSteps == 0 {
			u.UserInfo.UserDefaultSteps = h.DefaultSteps
		}
		switch data.Value {
		case "+":
			u.ChangeDefaultSteps(u.UserInfo.UserDefaultSteps + 1)
		case "-":
			u.ChangeDefaultSteps(u.UserInfo.UserDefaultSteps - 1)
		}
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s %s:\n%d", u.LoadLang("setDft"), u.LoadLang("steps"), u.UserInfo.UserDefaultSteps))
		msg.ReplyMarkup = generateSetDftStepsButton(u)
		h.bot.Send(msg)
	}
}

func (h *Handler) editImg(CallbackQuery *tgbotapi.CallbackQuery, u *UserInfo) {
	cfg := new(Config)
	fileID := utils.GetFileNamePrefix(CallbackQuery.Message.Document.FileName)
	var err error
	var prePhoto []byte
	prePhoto, err = h.cache.Get(fileID)
	if err != nil {
		prePhoto, err = h.bot.GetFileData(CallbackQuery.Message.Document.FileID)
		if err != nil {
			colorlog.Error(err)
			return
		}
		_, err := h.cache.Put(prePhoto)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
			return
		}
	}
	config, err := api.GetImgCfg(prePhoto)
	if err != nil {
		colorlog.Error(err)
		return
	}
	cfg.DrawConfig = *config
	cfg.PrePhotoID = fileID
	cfg.Steps = 28
	cfg.Strength = 0.6
	cfg.Mode = "DDIM"
	cfg.Seed = 0
	h.getConfig(u, cfg, CallbackQuery.Message.MessageID)
}

func getAllExtraModelName(tags, types string) (AllExtraModel []string) {
	re := regexp.MustCompile(fmt.Sprintf(`<%s:(.*?):((?:-?\d+)(?:\.\d+)?)>`, types))
	s := re.FindAllStringSubmatch(tags, -1)
	for _, v := range s {
		if len(v) == 3 {
			AllExtraModel = append(AllExtraModel, v[1])
		}
	}
	return
}

func getAllTag(tags string, re ...interface{ MatchString(b string) bool }) (allTag []string) {
	return purge(strings.Split(tags, ","), []func(string) string{strings.TrimSpace}, re)
}

func getAllTagName(tags string, re ...interface{ MatchString(b string) bool }) []string {
	return purge(strings.Split(tags, ","), []func(string) string{strings.TrimSpace, func(s string) string { name, _ := getTagName(s); return name }}, re)
}

func purge(all []string, preprocessing []func(string) string, re []interface{ MatchString(b string) bool }) (purged []string) {
	var Matchd bool
	for _, v := range all {
		for _, f := range preprocessing {
			v = f(v)
		}
		if v == "" {
			continue
		}
		for _, r := range re {
			if r.MatchString(v) {
				Matchd = true
				break
			}
		}
		if Matchd {
			Matchd = false
			continue
		}
		purged = append(purged, v)
	}
	return
}

func getTagName(tag string) (name string, strength float64) {
	if tag == "" {
		return "", 0
	}
	for (strings.HasPrefix(tag, "(") && strings.HasSuffix(tag, ")")) || (strings.HasPrefix(tag, "[") && strings.HasSuffix(tag, "]")) {
		tag = tag[1 : len(tag)-1]
	}
	i := strings.LastIndex(tag, ":")
	if i == -1 {
		return tag, 0
	}
	f, err := strconv.ParseFloat(tag[i+1:], 64)
	if err != nil {
		return tag, 0
	}
	return tag[:i], f
}

var reDefaultTag, _ = regexp.Compile(`^masterpiece$|^best quality$|^<lora:.*?:(-?\d+)(\.\d+)?>$|^<hypernet:.*?:(-?\d+)(\.\d+)?>$`)
var reDefaultUc, _ = regexp.Compile(`^lowres$|^text$`)

func (h *Handler) setCfg(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	before, after, found := strings.Cut(data.Value, ":")
	data.Key = before
	if found {
		data.Value = after
	} else {
		data.Value = before
	}
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(CallbackQuery.Message.Text), cfg)
	if err != nil {
		colorlog.Error(err)
		return
	}
	switch data.Key {
	case "editTag":
		i, err := strconv.ParseInt(data.Value, 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		allTag := getAllTag(cfg.Tag, reDefaultTag)
		h.bot.Send(tgbotapi.NewEditMessageReplyMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, *h.editTagButton(u, int(i), allTag, allTag)))
	case "changeTag":
		i := strings.LastIndex(data.Value, ":")
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		switch data.Value[:i] {
		case "Happend", "Eappend", "reset":
			m, err := h.bot.NewMsgCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID)
			if err != nil {
				return
			}
			defer m.Close()
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("sendTag"))))
			msg.ReplyMarkup = cancelButton(u)
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
			c, err := h.bot.NewCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID, CallbackQuery.Message.MessageID)
			if err != nil {
				colorlog.Error(err)
				return
			}
			defer c.Close()
			t := time.NewTimer(time.Minute * 3)
			defer t.Stop()
			select {
			case <-t.C:
			case <-c.Chan():
			case msg, ok := <-m.MsgChan():
				if ok {
					switch data.Value[:i] {
					case "Happend":
						cfg.Tag = fmt.Sprintf("%s, %s", msg.Text, cfg.Tag)
					case "Eappend":
						cfg.Tag = fmt.Sprintf("%s, %s", cfg.Tag, msg.Text)
					default:
						cfg.Tag = msg.Text
					}
				}
			}
		}
		arges := []ConfigFuncCorrentCfg{WithTag()}
		if data.Value[:i] == "translation" {
			arges = append(arges, WithTransTag())
		}
		h.CorrectCfg(cfg, u, arges...)
		allTag := getAllTag(cfg.Tag, reDefaultTag)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = h.editTagButton(u, int(page), allTag, allTag)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
		return
	case "sT":
		i := strings.LastIndex(data.Value, ":")
		tag := data.Value[:i]
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		allTag := getAllTag(cfg.Tag)
		var nowTag []string
		if k, ok := utils.InString(tag, allTag); ok {
			nowTag = append(nowTag, allTag[:k]...)
			nowTag = append(nowTag, allTag[k+1:]...)
			cfg.Tag = strings.Join(nowTag, ", ")
		} else {
			name, strength := getTagName(tag)
			if strength == 0 {
				strength = 1
			}
			allTag = append(allTag, fmt.Sprintf("(%s:%.1f)", name, strength))
			nowTag = allTag
			cfg.Tag = strings.Join(allTag, ", ")
		}
		h.CorrectCfg(cfg, u, WithTag())
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = h.editTagButton(u, int(page), nowTag, purge(allTag, []func(string) string{strings.TrimSpace}, []interface{ MatchString(b string) bool }{reDefaultTag}))
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "T":
		i := strings.LastIndex(data.Value, ":")
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		before := data.Value[:i]
		i = strings.LastIndex(before, ":")
		tag := before[:i]
		allTag := getAllTag(cfg.Tag)
		if k, ok := utils.InString(tag, allTag); ok {
			name, f := getTagName(allTag[k])
			if f <= 0 {
				f = 1
			}
			switch before[i+1:] {
			case "+":
				if f+0.1 >= 2 {
					f = 2
				} else {
					f += 0.1
				}
			case "-":
				if f-0.1 < 0 {
					f = 0.1
				} else {
					f -= 0.1
				}
			}
			allTag[k] = fmt.Sprintf("(%s:%.1f)", name, f)
		} else {
			name, f := getTagName(tag)
			if f == 0 {
				f = 1
			}
			allTag = append(allTag, fmt.Sprintf("(%s:%.1f)", name, f))
		}
		cfg.Tag = strings.Join(allTag, ", ")
		h.CorrectCfg(cfg, u, WithTag())
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = h.editTagButton(u, int(page), allTag, purge(allTag, []func(string) string{strings.TrimSpace}, []interface{ MatchString(b string) bool }{reDefaultTag}))
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "editUc":
		i, err := strconv.ParseInt(data.Value, 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		allUc := getAllTag(cfg.Uc, reDefaultUc)
		h.bot.Send(tgbotapi.NewEditMessageReplyMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, *h.editUcButton(u, int(i), allUc, allUc)))
	case "changeUc":
		i := strings.LastIndex(data.Value, ":")
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		switch data.Value[:i] {
		case "Happend", "Eappend", "reset":
			m, err := h.bot.NewMsgCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID)
			if err != nil {
				return
			}
			defer m.Close()
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("sendTag"))))
			msg.ReplyMarkup = cancelButton(u)
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
			c, err := h.bot.NewCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID, CallbackQuery.Message.MessageID)
			if err != nil {
				colorlog.Error(err)
				return
			}
			defer c.Close()
			t := time.NewTimer(time.Minute * 3)
			defer t.Stop()
			select {
			case <-t.C:
			case <-c.Chan():
			case msg, ok := <-m.MsgChan():
				if ok {
					switch data.Value[:i] {
					case "Happend":
						cfg.Uc = fmt.Sprintf("%s, %s", msg.Text, cfg.Uc)
					case "Eappend":
						cfg.Uc = fmt.Sprintf("%s, %s", cfg.Uc, msg.Text)
					default:
						cfg.Uc = msg.Text
					}
				}
			}
		}
		arges := []ConfigFuncCorrentCfg{WithUc()}
		if data.Value[:i] == "translation" {
			arges = append(arges, WithTransUc())
		}
		h.CorrectCfg(cfg, u, arges...)
		allUc := getAllTag(cfg.Uc, reDefaultUc)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = h.editUcButton(u, int(page), allUc, allUc)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
		return
	case "sU":
		i := strings.LastIndex(data.Value, ":")
		uc := data.Value[:i]
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		allUc := getAllTag(cfg.Uc)
		var nowUc []string
		if k, ok := utils.InString(uc, allUc); ok {
			nowUc = append(nowUc, allUc[:k]...)
			nowUc = append(nowUc, allUc[k+1:]...)
			cfg.Uc = strings.Join(nowUc, ", ")
		} else {
			name, strength := getTagName(uc)
			if strength == 0 {
				strength = 1
			}
			allUc = append(allUc, fmt.Sprintf("(%s:%.1f)", name, strength))
			nowUc = allUc
			cfg.Uc = strings.Join(allUc, ", ")
		}
		h.CorrectCfg(cfg, u, WithUc())
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = h.editUcButton(u, int(page), nowUc, purge(allUc, []func(string) string{strings.TrimSpace}, []interface{ MatchString(b string) bool }{reDefaultUc}))
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "uc":
		i := strings.LastIndex(data.Value, ":")
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		before := data.Value[:i]
		i = strings.LastIndex(before, ":")
		uc := before[:i]
		allUc := getAllTag(cfg.Uc)
		if k, ok := utils.InString(uc, allUc); ok {
			name, f := getTagName(allUc[k])
			if f <= 0 {
				f = 1
			}
			switch before[i+1:] {
			case "+":
				if f+0.1 >= 2 {
					f = 2
				} else {
					f += 0.1
				}
			case "-":
				if f-0.1 < 0 {
					f = 0.1
				} else {
					f -= 0.1
				}
			}
			allUc[k] = fmt.Sprintf("(%s:%.1f)", name, f)
		} else {
			name, f := getTagName(uc)
			if f == 0 {
				f = 1
			}
			allUc = append(allUc, fmt.Sprintf("(%s:%.1f)", name, f))
		}
		cfg.Uc = strings.Join(allUc, ", ")
		h.CorrectCfg(cfg, u, WithUc())
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = h.editUcButton(u, int(page), allUc, purge(allUc, []func(string) string{strings.TrimSpace}, []interface{ MatchString(b string) bool }{reDefaultUc}))
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "sL":
		group, after, ok := strings.Cut(data.Value, ":")
		if !ok {
			return
		}
		groupIndex, err := strconv.ParseInt(group, 10, 64)
		if err != nil {
			colorlog.Errorf("[%s] get group id err: %v", u.ChatMember.User.String(), err)
			return
		}
		model, after, ok := strings.Cut(after, ":")
		if !ok {
			return
		}
		modelIndex, err := strconv.ParseInt(model, 10, 64)
		if err != nil {
			colorlog.Errorf("[%s] get model id err: %v", u.ChatMember.User.String(), err)
			return
		}
		l := h.Index2ExtraModel(int(groupIndex), int(modelIndex))
		model = l.Name
		page, err := strconv.ParseInt(after, 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		reLora, err := regexp.Compile(fmt.Sprintf("[\\[\\(]*<%s:%s:((?:-?\\d+)(?:\\.\\d+)?)>[\\)\\]]*", l.Type, model))
		if err != nil {
			colorlog.Error(err)
			return
		}
		var showPreview bool
		if err != nil {
			cfg.Tag = reLora.ReplaceAllString(cfg.Tag, "")
		} else {
			if reLora.MatchString(cfg.Tag) {
				cfg.Tag = reLora.ReplaceAllString(cfg.Tag, "")
				if len(l.TriggerWords) != 0 {
					if k, ok := utils.InString(l.TriggerWords[0], getAllTagName(cfg.Tag)); ok {
						s := getAllTag(cfg.Tag)
						cfg.Tag = strings.Join(append(s[:k], s[k+1:]...), ", ")
					}
				}
			} else {
				showPreview = true
				if len(l.TriggerWords) != 0 {
					if _, ok := utils.InString(l.TriggerWords[0], getAllTagName(cfg.Tag, reDefaultTag)); !ok {
						cfg.Tag += fmt.Sprintf(", <lora:%s:0.6>, %s", model, l.TriggerWords[0])
					} else {
						cfg.Tag += fmt.Sprintf(", <lora:%s:0.6>", model)
					}
				} else {
					cfg.Tag += fmt.Sprintf(", <lora:%s:0.6>", model)
				}
			}
		}
		h.CorrectCfg(cfg, u, WithTag())
		var msg bytes.Buffer
		msg.WriteString(fmt.Sprintf("%s\n<pre># %s</pre>\n", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("extraModelInfo"))))
		for k, v := range l.TriggerWords {
			msg.WriteString(fmt.Sprintf("\n# <code>%s</code>", parse2HTML(v)))
			if k == len(l.TriggerWords)-1 {
				msg.WriteByte('\n')
			}
		}
		if showPreview && h.DB.DBType() == db.T_POSTGRESQL {
			photo, err := h.DB.FindImg(db.FindConfig{
				Deadline:   time.Now(),
				Order:      "id desc",
				Limit:      1,
				KeywordsRe: []string{fmt.Sprintf("<lora:%s:(-?\\d+)(\\.\\d+)?>", l.Name)},
			})
			if err == nil && len(photo) == 1 {
				msg.WriteString(fmt.Sprintf("\n# <a href=\"%s://%s/api/images/%s.png\">Preview</a>", parseflag.ApiScheme, parseflag.ApiHost, photo[0].FileID))
			}
		}
		tgMsg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, msg.String())
		tgMsg.ReplyMarkup = h.generateExtraModelButton(u, int(page), int(groupIndex), getAllExtraModelName(cfg.Tag, l.Type))
		tgMsg.ParseMode = "HTML"
		h.bot.Send(tgMsg)
	case "L":
		group, after, ok := strings.Cut(data.Value, ":")
		if !ok {
			return
		}
		groupIndex, err := strconv.ParseInt(group, 10, 64)
		if err != nil {
			colorlog.Errorf("[%s] get group id err: %v", u.ChatMember.User.String(), err)
			return
		}
		model, after, ok := strings.Cut(after, ":")
		if !ok {
			return
		}
		modelIndex, err := strconv.ParseInt(model, 10, 64)
		if err != nil {
			colorlog.Errorf("[%s] get model id err: %v", u.ChatMember.User.String(), err)
			return
		}
		l := h.Index2ExtraModel(int(groupIndex), int(modelIndex))
		model = l.Name
		t, p, found := strings.Cut(after, ":")
		if !found {
			return
		}
		page, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		reLora, err := regexp.Compile(fmt.Sprintf("[\\[\\(]*<%s:%s:((?:-?\\d+)(?:\\.\\d+)?)>[\\)\\]]*", l.Type, model))
		if err != nil {
			colorlog.Error(err)
			return
		}
		var showPreview bool
		if err != nil {
			cfg.Tag = reLora.ReplaceAllString(cfg.Tag, "")
		} else {
			if s := reLora.FindStringSubmatch(cfg.Tag); len(s) == 2 {
				f, err := strconv.ParseFloat(s[1], 64)
				if err != nil {
					colorlog.Error(err)
					return
				}
				switch t {
				case "+":
					if f+0.1 > 2 {
						return
					}
					cfg.Tag = reLora.ReplaceAllString(cfg.Tag, fmt.Sprintf("<%s:%s:%.1f>", l.Type, model, f+0.1))
				case "-":
					if f-0.1 <= 0 {
						return
					}
					cfg.Tag = reLora.ReplaceAllString(cfg.Tag, fmt.Sprintf("<%s:%s:%.1f>", l.Type, model, f-0.1))
				}
			} else {
				showPreview = true
				if len(l.TriggerWords) != 0 {
					if _, ok := utils.InString(l.TriggerWords[0], getAllTagName(cfg.Tag)); !ok {
						cfg.Tag += fmt.Sprintf(", <%s:%s:0.6>, %s", l.Type, model, l.TriggerWords[0])
					} else {
						cfg.Tag += fmt.Sprintf(", <%s:%s:0.6>", l.Type, model)
					}
				} else {
					cfg.Tag += fmt.Sprintf(", <%s:%s:0.6>", l.Type, model)
				}
			}
		}
		h.CorrectCfg(cfg, u, WithTag())
		var msg bytes.Buffer
		msg.WriteString(fmt.Sprintf("%s\n<pre># %s</pre>\n", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("extraModelInfo"))))
		for k, v := range l.TriggerWords {
			msg.WriteString(fmt.Sprintf("\n<code># %s</code>", parse2HTML(v)))
			if k == len(l.TriggerWords)-1 {
				msg.WriteByte('\n')
			}
		}
		if showPreview && h.DB.DBType() == db.T_POSTGRESQL {
			photo, err := h.DB.FindImg(db.FindConfig{
				Deadline:   time.Now(),
				Order:      "id desc",
				Limit:      1,
				KeywordsRe: []string{fmt.Sprintf("<%s:%s:(-?\\d+)(\\.\\d+)?>", l.Type, l.Name)},
			})
			if err == nil && len(photo) == 1 {
				msg.WriteString(fmt.Sprintf("\n# <a href=\"%s://%s/api/images/%s.png\">Preview</a>", parseflag.ApiScheme, parseflag.ApiHost, photo[0].FileID))
			}
		}
		tgMsg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, msg.String())
		tgMsg.ParseMode = "HTML"
		tgMsg.ReplyMarkup = h.generateExtraModelButton(u, int(page), int(groupIndex), getAllExtraModelName(cfg.Tag, l.Type))
		h.bot.Send(tgMsg)
	case "extraModel":
		i := strings.LastIndex(data.Value, ":")
		group := data.Value[:i]
		groupIndex, err := strconv.ParseInt(group, 10, 64)
		if err != nil {
			colorlog.Errorf("[%s] get group id err: %v", u.ChatMember.User.String(), err)
			return
		}
		page, err := strconv.ParseInt(data.Value[i+1:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("extraModelInfo"))))
		msg.ReplyMarkup = h.generateExtraModelButton(u, int(page), int(groupIndex), getAllExtraModelName(cfg.Tag, `\w+`))
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "extraModelGroup":
		page, err := strconv.ParseInt(data.Value, 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("extraModelInfo"))))
		msg.ReplyMarkup = h.generateAllExtraModelGroupButton(u, int(page))
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "resetSeed":
		cfg.Seed = uint32(rand.Intn(math.MaxUint32))
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "setImg":
		if cfg.PrePhotoID != "" {
			cfg.PrePhotoID = ""
			cfg.Strength = 0
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
			msg.ReplyMarkup = panelButton(u, false, cfg.ControlPhotoID != "")
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
			return
		}
		m, err := h.bot.NewMsgCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID)
		if err != nil {
			return
		}
		defer m.Close()
		c, err := h.bot.NewCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID, CallbackQuery.Message.MessageID)
		if err != nil {
			return
		}
		defer c.Close()
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s, %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("setImgInfo")), parse2HTML(u.LoadLang("sendImg"))))
		msg.ReplyMarkup = cancelButton(u)
		msg.ParseMode = "HTML"
		_, err = h.bot.Send(msg)
		if err != nil {
			return
		}
		t := time.NewTimer(time.Minute * 5)
		defer t.Stop()
		select {
		case <-t.C:
			break
		case <-c.Chan():
			break
		case msg, ok := <-m.MsgChan():
			m.Close()
			if !ok {
				return
			}
			var photo []byte
			if len(msg.Photo) > 0 {
				latestPhoto := msg.Photo[len(msg.Photo)-1]
				cfg.Width = latestPhoto.Width
				cfg.Height = latestPhoto.Height
				photo, err = h.bot.GetFileData(latestPhoto.FileID)
				if err != nil {
					colorlog.Errorf("Get photo err: %v", err)
					break
				}
				fi, err := h.cache.Put(photo)
				if err != nil {
					colorlog.Errorf("Put file err: %v", err)
					break
				}
				cfg.PrePhotoID = fi.FileID
			} else if msg.Document != nil {
				if _, ok := utils.InString(msg.Document.MimeType, avilableDocumentType); !ok {
					colorlog.Errorf("Get photo err: %s", "document type is not avilable")
					break
				}
				if photo, err = h.cache.Get(utils.GetFileNamePrefix(msg.Document.FileName)); err != nil {
					photo, err = h.bot.GetFileData(msg.Document.FileID)
					if err != nil {
						colorlog.Errorf("Parse Photo err: %v", err)
						break
					}
				}
				cfg.Width, cfg.Height, err = utils.GetPhotoSize(photo)
				if err != nil {
					colorlog.Errorf("Parse Photo err: %v", err)
					break
				}
				fi, err := h.cache.Put(photo)
				if err != nil {
					colorlog.Errorf("Put file err: %v", err)
					break
				}
				cfg.PrePhotoID = fi.FileID
			}
		}
		cfg.Strength = 0.7
		msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "setControl":
		if cfg.ControlPhotoID != "" {
			cfg.ControlPhotoID = ""
			h.CorrectCfg(cfg, u, WithCtrlPhoto())
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
			msg.ReplyMarkup = panelButton(u, false, cfg.ControlPhotoID != "")
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
			return
		}
		m, err := h.bot.NewMsgCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID)
		if err != nil {
			return
		}
		defer m.Close()
		c, err := h.bot.NewCbk(CallbackQuery.Message.Chat.ID, CallbackQuery.From.ID, CallbackQuery.Message.MessageID)
		if err != nil {
			return
		}
		defer c.Close()
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s, %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("setImgInfo")), parse2HTML(u.LoadLang("sendImg"))))
		msg.ReplyMarkup = cancelButton(u)
		msg.ParseMode = "HTML"
		_, err = h.bot.Send(msg)
		if err != nil {
			return
		}
		t := time.NewTimer(time.Minute * 5)
		defer t.Stop()
		select {
		case <-t.C:
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
			msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
			return
		case <-c.Chan():
			msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
			msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
			return
		case msg, ok := <-m.MsgChan():
			m.Close()
			if !ok {
				return
			}
			var photo []byte
			if len(msg.Photo) > 0 {
				latestPhoto := msg.Photo[len(msg.Photo)-1]
				photo, err = h.bot.GetFileData(latestPhoto.FileID)
				if err != nil {
					colorlog.Errorf("Get photo err: %v", err)
					break
				}
				fi, err := h.cache.Put(photo)
				if err != nil {
					colorlog.Errorf("Put file err: %v", err)
					break
				}
				cfg.ControlPhotoID = fi.FileID
			} else if msg.Document != nil {
				if _, ok := utils.InString(msg.Document.MimeType, avilableDocumentType); !ok {
					colorlog.Errorf("Get photo err: %s", "document type is not avilable")
					break
				}
				if photo, err = h.cache.Get(utils.GetFileNamePrefix(msg.Document.FileName)); err != nil {
					photo, err = h.bot.GetFileData(msg.Document.FileID)
					if err != nil {
						colorlog.Errorf("Parse Photo err: %v", err)
						break
					}
				}
				var fi cache.FileInfo
				fi, err = h.cache.Put(photo)
				if err != nil {
					colorlog.Errorf("Put file err: %v", err)
					break
				}
				cfg.ControlPhotoID = fi.FileID
			} else {
				msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
				msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
				msg.ParseMode = "HTML"
				h.bot.Send(msg)
				return
			}
			if cfg.PrePhotoID == "" {
				width, hight, err := utils.GetPhotoSize(photo)
				if err != nil {
					msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
					msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
					msg.ParseMode = "HTML"
					h.bot.Send(msg)
					return
				}
				cfg.Width = width
				cfg.Height = hight
			}
		}
		h.CorrectCfg(cfg, u, WithCtrlPhoto())
		msg = tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = editControlButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "editControl":
		h.bot.Send(tgbotapi.NewEditMessageReplyMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, *editControlButton(u)))
	case "controlPreprocess":
		h.bot.Send(tgbotapi.NewEditMessageReplyMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, *h.controlPreprocessButton(u, cfg.ControlPreprocess)))
	case "controlProcess":
		h.bot.Send(tgbotapi.NewEditMessageReplyMarkup(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, *h.controlProcessButton(cfg.ControlProcess)))
	case "preprocess":
		cfg.ControlPreprocess = data.Value
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = editControlButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "process":
		cfg.ControlProcess = data.Value
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = editControlButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "confirm":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "sizeType":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(fmt.Sprintf("%d * %d = %d (Min: H*W>=64*64, Max: H*W<= %d [Unsponsored maximum Size is %d])", cfg.Width, cfg.Height, cfg.Width*cfg.Height, parseflag.ImgMaxSize, GuestImgMaxSize))))
		msg.ParseMode = "HTML"
		switch data.Value {
		case "16:9":
			msg.ReplyMarkup = generate16_9Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "9:16":
			msg.ReplyMarkup = generate9_16Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "4:3":
			msg.ReplyMarkup = generate4_3Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "3:4":
			msg.ReplyMarkup = generate3_4Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "1:1":
			msg.ReplyMarkup = generate1_1Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "2:3":
			msg.ReplyMarkup = generate2_3Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "3:2":
			msg.ReplyMarkup = generate3_2Button(u, fmt.Sprint(cfg.Width, "*", cfg.Height))
		case "custom":
			msg.ReplyMarkup = custonSizeButton(u)
		}
		h.bot.Send(msg)
	case "custonSize":
		i, err := strconv.ParseInt(data.Value[2:], 10, 64)
		if err != nil {
			colorlog.Error(err)
			return
		}
		switch data.Value[:2] {
		case "w-":
			if cfg.Width-int(i) >= 64 {
				cfg.Width -= int(i)
			} else {
				cfg.Width = 64
			}
		case "w+":
			if (cfg.Width+int(i))*cfg.Height <= parseflag.ImgMaxSize {
				cfg.Width += int(i)
			}
		case "h-":
			if cfg.Height-int(i) >= 64 {
				cfg.Height -= int(i)
			} else {
				cfg.Height = 64
			}
		case "h+":
			if (cfg.Height+int(i))*cfg.Width <= parseflag.ImgMaxSize {
				cfg.Height += int(i)
			}
		}
		h.CorrectCfg(cfg, u)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(fmt.Sprintf("%d * %d = %d (Min: H*W>=64*64, Max: H*W<= %d [Unsponsored maximum Size is %d])", cfg.Width, cfg.Height, cfg.Width*cfg.Height, parseflag.ImgMaxSize, GuestImgMaxSize))))
		msg.ReplyMarkup = custonSizeButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "size":
		Width, Height, f := strings.Cut(data.Value, "*")
		if f {
			w, err := strconv.ParseInt(Width, 10, 64)
			if err == nil {
				cfg.Width = int(w)
			}
			h, err := strconv.ParseInt(Height, 10, 64)
			if err == nil {
				cfg.Height = int(h)
			}
		}
		h.CorrectCfg(cfg, u)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "num":
		i, err := strconv.ParseInt(data.Value, 10, 64)
		if err != nil {
			return
		}
		cfg.Num = int(i)
		h.CorrectCfg(cfg, u)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "mode":
		cfg.Mode = data.Value
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "strength":
		switch data.Value[:1] {
		case "+":
			f, err := strconv.ParseFloat(data.Value[1:], 32)
			if err != nil {
				colorlog.Error(err)
				return
			}
			if cfg.Strength+float64(f) <= 0.99 {
				cfg.Strength += float64(f)
			} else {
				cfg.Strength = 0.99
			}
			cfg.Strength = utils.TwoDot(cfg.Strength)
		case "-":
			f, err := strconv.ParseFloat(data.Value[1:], 32)
			if err != nil {
				colorlog.Error(err)
				return
			}
			if cfg.Strength-float64(f) > 0 {
				cfg.Strength -= float64(f)
			} else {
				cfg.Strength = 0
			}
			cfg.Strength = utils.TwoDot(cfg.Strength)
		}
		h.CorrectCfg(cfg, u, WithPhoto())
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("strengthInfo"))))
		msg.ReplyMarkup = strengthButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "scale":
		v, err := strconv.ParseInt(data.Value[1:], 10, 64)
		if err != nil {
			return
		}
		switch data.Value[:1] {
		case "+":
			if cfg.CfgScale+int(v) <= 30 {
				cfg.CfgScale += int(v)
			} else {
				cfg.CfgScale = 30
			}
		case "-":
			if cfg.CfgScale-int(v) >= 1 {
				cfg.CfgScale -= int(v)
			} else {
				cfg.CfgScale = 1
			}
		}
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("scaleInfo"))))
		msg.ReplyMarkup = scaleButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "model":
		cfg.Model = data.Value
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "steps":
		v, err := strconv.ParseInt(data.Value[1:], 10, 64)
		if err != nil {
			return
		}
		switch data.Value[:1] {
		case "-":
			if cfg.Steps-int(v) >= 15 {
				cfg.Steps -= int(v)
			} else {
				cfg.Steps = 15
			}
		case "+":
			if cfg.Steps+int(v) <= 50 {
				cfg.Steps += int(v)
			} else {
				cfg.Steps = 50
			}
		}
		h.CorrectCfg(cfg, u)
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("stepsInfo")), parse2HTML("(Min: 15, Max: 50 [Unsponsored maximum Steps is 28])")))
		msg.ReplyMarkup = stepsButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "cancel":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	}
}

func (h *Handler) panel(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	cfg := &Config{}
	err := yaml.Unmarshal([]byte(CallbackQuery.Message.Text), cfg)
	if err != nil {
		colorlog.Error(err)
		return
	}
	switch data.Value {
	case "confirm":
		task, err := u.AddTask(T_Draw)
		if err != nil {
			return
		}
		defer task.Down()
		h.CorrectCfg(cfg, u, WithTag())
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, string(cfg.Fomate2TgHTML()))
		msg.ReplyMarkup = reDrawButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
		var photo []byte
		if cfg.PrePhotoID != "" {
			photo, err = h.cache.Get(cfg.PrePhotoID)
			if err != nil {
				colorlog.Error(err)
				return
			}
		}
		var controlPhoto []byte
		if cfg.ControlPhotoID != "" {
			controlPhoto, err = h.cache.Get(cfg.ControlPhotoID)
			if err != nil {
				colorlog.Error(err)
				return
			}
		}
		h.drawAndSend(u, CallbackQuery.Message.MessageID, cfg, photo, controlPhoto)
		return
	case "size":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(fmt.Sprintf("%d * %d = %d (Min: H*W>=64*64, Max: H*W<= %d [Unsponsored maximum Size is %d])", cfg.Width, cfg.Height, cfg.Width*cfg.Height, parseflag.ImgMaxSize, GuestImgMaxSize))))
		msg.ReplyMarkup = sizeTypeButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "num":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("freeMaxNum"))))
		msg.ReplyMarkup = h.generateNUMButton(cfg.Num)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "mode":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("modeInfo"))))
		msg.ReplyMarkup = h.generateMODEButton(cfg.Mode)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "strength":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("strengthInfo"))))
		msg.ReplyMarkup = strengthButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "scale":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("scaleInfo"))))
		msg.ReplyMarkup = scaleButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "model":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("modelInfo"))))
		msg.ReplyMarkup = h.generateModelButton(cfg.Model)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	case "steps":
		msg := tgbotapi.NewEditMessageText(CallbackQuery.Message.Chat.ID, CallbackQuery.Message.MessageID, fmt.Sprintf("%s\n<pre># %s %s</pre>", cfg.Fomate2TgHTML(), parse2HTML(u.LoadLang("stepsInfo")), parse2HTML("(Min: 15, Max: 50 [Unsponsored maximum Steps is 28])")))
		msg.ReplyMarkup = stepsButton(u)
		msg.ParseMode = "HTML"
		h.bot.Send(msg)
	}
}

func (h *Handler) superResolution(CallbackQuery *tgbotapi.CallbackQuery, data *tgbotapi.ChanData, u *UserInfo) {
	if u.Permissions() == T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(h.bot))
		msg.ReplyMarkup = h.goJoinButton(u)
		msg.ReplyToMessageID = CallbackQuery.Message.MessageID
		h.bot.Send(msg)
		return
	}
	if CallbackQuery.Message.Document == nil || CallbackQuery.Message.Document.FileID == "" {
		return
	}
	task, err := u.AddTask(T_SuperResolution)
	if err != nil {
		return
	}
	defer task.Down()
	resize, err := strconv.ParseInt(data.Value, 10, 64)
	if err != nil {
		return
	}
	cfg := new(Config)
	fileID := utils.GetFileNamePrefix(CallbackQuery.Message.Document.FileName)
	var prePhoto []byte
	prePhoto, err = h.cache.Get(fileID)
	if err != nil {
		prePhoto, err = h.bot.GetFileData(CallbackQuery.Message.Document.FileID)
		if err != nil {
			colorlog.Error(err)
			return
		}
		_, err := h.cache.Put(prePhoto)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		}
	}
	cfg.DrawConfig = api.DrawConfig{}
	cfg.PrePhotoID = fileID
	cfg.Width, cfg.Height, err = utils.GetPhotoSize(prePhoto)
	if err != nil {
		colorlog.Error(err)
		return
	}
	h.superResolutionRun(u, CallbackQuery.Message.MessageID, cfg, prePhoto, int(resize))
}

func (h *Handler) fineTune(CallbackQuery *tgbotapi.CallbackQuery, Strings string, u *UserInfo) {
	if u.Permissions() == T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(h.bot))
		msg.ReplyMarkup = h.goJoinButton(u)
		msg.ReplyToMessageID = CallbackQuery.Message.MessageID
		h.bot.Send(msg)
		return
	}
	if CallbackQuery.Message.Document == nil || CallbackQuery.Message.Document.FileID == "" {
		return
	}
	task, err := u.AddTask(T_Draw)
	if err != nil {
		return
	}
	defer task.Down()
	cfg := new(Config)
	fileID := utils.GetFileNamePrefix(CallbackQuery.Message.Document.FileName)
	var prePhoto []byte
	prePhoto, err = h.cache.Get(fileID)
	if err != nil {
		prePhoto, err = h.bot.GetFileData(CallbackQuery.Message.Document.FileID)
		if err != nil {
			colorlog.Error(err)
			return
		}
		_, err := h.cache.Put(prePhoto)
		if err != nil {
			colorlog.Errorf("Put file err: %v", err)
		}
	}
	config, err := api.GetImgCfg(prePhoto)
	if err != nil {
		colorlog.Error(err)
		return
	}
	cfg.DrawConfig = *config
	cfg.Mode = "DDIM"
	if u.Permissions() == T_Subscribe {
		cfg.Steps = 50
	} else {
		cfg.Steps = 28
	}
	cfg.Num = 2
	cfg.PrePhotoID = fileID
	if Strings == "1" {
		cfg.Strength = 0.2
	} else if Strings == "2" {
		cfg.Strength = 0.4
	} else {
		cfg.Strength = 0.6
	}
	cfg.Seed = 0
	h.CorrectCfg(cfg, u, WithSeed(), WithTag(), WithUc(), WithPhoto())
	h.drawAndSend(u, CallbackQuery.Message.MessageID, cfg, prePhoto, nil)
}

func (h *Handler) reDraw(CallbackQuery *tgbotapi.CallbackQuery, u *UserInfo) {
	if u.Permissions() == T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(h.bot))
		msg.ReplyMarkup = h.goJoinButton(u)
		msg.ReplyToMessageID = CallbackQuery.Message.MessageID
		h.bot.Send(msg)
		return
	}
	task, err := u.AddTask(T_Draw)
	if err != nil {
		return
	}
	defer task.Down()
	cfg := &Config{}
	if yaml.Unmarshal([]byte(CallbackQuery.Message.Text), cfg) == nil {
		var prePhoto []byte
		if cfg.PrePhotoID != "" {
			prePhoto, err = h.cache.Get(cfg.PrePhotoID)
			if err != nil {
				colorlog.Error(err)
				return
			}
		}
		var controlPhoto []byte
		if cfg.ControlPhotoID != "" {
			controlPhoto, err = h.cache.Get(cfg.ControlPhotoID)
			if err != nil {
				colorlog.Error(err)
				return
			}
		}
		cfg.Seed = 0
		h.CorrectCfg(cfg, u, WithSeed())
		mc := tgbotapi.NewMessage(u.ChatMember.User.ID, string(cfg.Fomate2TgHTML()))
		mc.ReplyMarkup = reDrawButton(u)
		mc.ParseMode = "HTML"
		mc.ReplyToMessageID = CallbackQuery.Message.MessageID
		m, err := h.bot.Send(mc)
		if err != nil {
			colorlog.Error(err)
			return
		}
		h.drawAndSend(u, m.MessageID, cfg, prePhoto, controlPhoto)
	}
}
