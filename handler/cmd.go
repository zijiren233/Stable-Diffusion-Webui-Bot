package handler

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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
)

func HandleCmd(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	u, err := user.LoadAndInitUser(bot, Message.From.ID)
	if err != nil {
		colorlog.Errorf("Load And Init User Err: %v", err)
		return
	}
	switch Message.Command() {
	case "drawpool":
		drawPool(bot, Message)
	case "gettoken":
		getToken(bot, Message)
	case "token":
		token(bot, Message, u)
	case "usefree":
		usefree(bot, Message, u)
	case "unsubscribe":
		unsubscribe(bot, Message)
	case "subscribe":
		subscribe(bot, Message, u)
	case "info":
		info(bot, Message)
	case "setdefault":
		newSetDft(bot, Message, u)
	case "start":
		start(bot, Message, u)
	case "help":
		msg := tgbotapi.NewMessage(Message.Chat.ID, u.LoadLang("help"))
		msg.ReplyMarkup = goGuideButton(u)
		bot.Send(msg)
	case "history":
		history(bot, Message, u)
	case "api":
		apis(bot, Message, u)
	case "web":
		web(bot, Message, u)
	case "language":
		setLanguage(bot, Message)
	case "img2tag":
		img2tag(bot, Message, u)
	case "superresolution":
		superresolution(bot, Message, u)
	case "guesstag":
		guessTag(bot, Message, u)
	case "dev":
		dev(bot, Message)
	case "share":
		share(bot, Message, u)
	case "invite":
		invite(bot, Message, u)
	}
}

func invite(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	rawUrl, err := url.Parse(fmt.Sprintf("https://t.me/%s", bot.Self.UserName))
	if err != nil {
		colorlog.Error(err)
		return
	}
	query := rawUrl.Query()
	query.Set("start", fmt.Sprintf("invite-%d", u.UserInfo.UserID))
	rawUrl.RawQuery = query.Encode()
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\n`%s`", u.LoadLang("invite"), rawUrl.String()))
	msg.ReplyToMessageID = Message.MessageID
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func start(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	args := strings.Split(Message.CommandArguments(), "-")
	switch args[0] {
	case "invite":
		handelInvite(bot, Message, u, args)
	default:
		msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\n\nYou Can Use Website\nUser ID: `%d`\nPassword: `%s`", u.LoadLang("help"), u.UserInfo.UserID, u.Passwd()))
		msg.ReplyToMessageID = Message.MessageID
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: append(helpLangButton.InlineKeyboard, clictUrlButton(u, fmt.Sprintf("https://%s/login", parseflag.WEBSITE)).InlineKeyboard...)}
		bot.Send(msg)
	}
}

func usefree(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	if Message.From.ID != parseflag.MyID {
		return
	}
	args := strings.Split(Message.CommandArguments(), " ")
	if len(args) != 2 {
		return
	}
	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || userID == 0 {
		return
	}
	useCount, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil || useCount == 0 {
		return
	}
	msg := tgbotapi.NewMessage(Message.Chat.ID, "")
	msg.ReplyToMessageID = Message.MessageID
	useUser, err := user.LoadUser(bot, userID)
	if err != nil {
		msg.Text = err.Error()
		bot.Send(msg)
		return
	}
	useUser.UseFree(int(useCount))
	msg.Text = fmt.Sprintf("user: %s\nfree count: %d", useUser.ChatMember.User.String(), useUser.Subscribe.FreeAmount)
	bot.Send(msg)
}

var (
	invitedList = []int64{}
)

func handelInvite(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo, args []string) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, "")
	msg.ReplyToMessageID = Message.MessageID
	if len(args) != 2 {
		msg.Text = "Error"
		bot.Send(msg)
		return
	}
	id, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		msg.Text = "Error"
		bot.Send(msg)
		return
	}
	if id == u.UserInfo.UserID || u.Subscribe.CreatedAt.Before(time.Now().Add(time.Minute*(-15))) || func() bool {
		_, in := utils.In(invitedList, func(v int64) bool {
			return v == u.Subscribe.UserID
		})
		return in
	}() {
		msg.Text = fmt.Sprintf("%s\n\nYou Can Use Website\nUser ID: `%d`\nPassword: `%s`", u.LoadLang("help"), u.UserInfo.UserID, u.Passwd())
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: append(helpLangButton.InlineKeyboard, clictUrlButton(u, fmt.Sprintf("https://%s/login", parseflag.WEBSITE)).InlineKeyboard...)}
		bot.Send(msg)
		return
	}
	iUser, err := user.LoadUser(bot, id)
	if err != nil {
		msg.Text = "Error"
		bot.Send(msg)
	}
	iUser.UseFree(-5)
	u.UseFree(-10)
	invitedList = append(invitedList, u.Subscribe.UserID)
	bot.Send(tgbotapi.NewMessage(iUser.UserInfo.UserID, strings.ReplaceAll(strings.ReplaceAll(iUser.LoadLang("inviteSuccess"), "{{user}}", u.ChatMember.User.String()), "{{freeAmount}}", fmt.Sprint(iUser.Subscribe.FreeAmount))))
	msg.Text = strings.ReplaceAll(strings.ReplaceAll(iUser.LoadLang("wasInvite"), "{{user}}", iUser.ChatMember.User.String()), "{{freeAmount}}", fmt.Sprint(u.Subscribe.FreeAmount))
	bot.Send(msg)
}

func newSetDft(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, u.LoadLang("setDft"))
	msg.ReplyToMessageID = Message.MessageID
	msg.ReplyMarkup = setDefaultCfg(u)
	bot.Send(msg)
}

func info(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	if Message.From.ID != parseflag.MyID {
		return
	}
	uID, err := strconv.ParseInt(Message.CommandArguments(), 10, 64)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprint(err)))
		return
	}
	u, err := user.LoadUser(bot, uID)
	if err != nil {
		colorlog.Errorf("Load User Err: %v", err)
		return
	}
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("Type: %v\nPwd: `%s`\nLeft: %v\nBanned: %v\nDeadline: `%s`", u.Permissions(), u.UserInfo.Passwd, u.ChatMember.HasLeft(), u.ChatMember.WasKicked(), u.Subscribe.Deadline.Format("2006-01-02 15:04:05")))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func subscribe(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	if u.Subscribe.Deadline.Before(time.Now()) {
		msg := tgbotapi.NewMessage(Message.Chat.ID, u.LoadLang("noSubscribe"))
		msg.ReplyToMessageID = Message.MessageID
		msg.ReplyMarkup = goJoinButton(u)
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("Deadline: `%s`", u.Subscribe.Deadline.Format("2006-01-02 15:04:05")))
		msg.ReplyToMessageID = Message.MessageID
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	}
}

func unsubscribe(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	if Message.From.ID != parseflag.MyID {
		return
	}
	uID, err := strconv.ParseInt(Message.CommandArguments(), 10, 64)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprint(err)))
		return
	}
	ui, err := user.LoadUser(bot, uID)
	if err != nil {
		colorlog.Errorf("Load User Err: %v", err)
		return
	}
	ui.Subscribe.Deadline = time.Now()
	ret := db.DB().Model(&db.Subscribe{}).Omit("updated_at").Where("user_id = ?", uID).Update("deadline", time.Now())
	if ret.RowsAffected == 0 {
		colorlog.Error(Message.Chat.ID, "user not found")
		bot.Send(tgbotapi.NewMessage(Message.Chat.ID, "user not found"))
	}
}

func share(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	if u.Permissions() != user.T_Subscribe {
		msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\n%s", u.LoadLang("shareInfo"), u.LoadLang("mustShare")))
		msg.ReplyToMessageID = Message.MessageID
		msg.ReplyMarkup = goJoinButton(u)
		bot.Send(msg)
		return
	}
	var option string
	if u.UserInfo.SharePhoto {
		option = u.LoadLang("enable")
	} else {
		option = u.LoadLang("disable")
	}
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s: %s", u.LoadLang("shareInfo"), option))
	msg.ReplyToMessageID = Message.MessageID
	msg.ReplyMarkup = gShareButton(u)
	bot.Send(msg)
}

func history(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\nUser ID: `%d`\nPassword: `%s`", u.LoadLang("history"), u.UserInfo.UserID, u.Passwd()))
	msg.ReplyToMessageID = Message.MessageID
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = clictUrlButton(u, fmt.Sprintf("https://%s/waterfall", parseflag.WEBSITE))
	bot.Send(msg)
}

func apis(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("API use Basic Auth\nUser ID: `%d`\nPassword: `%s`", u.UserInfo.UserID, u.Passwd()))
	msg.ReplyToMessageID = Message.MessageID
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = clictUrlButton(u, fmt.Sprintf("https://%s/docs/index.html", parseflag.HOST))
	bot.Send(msg)
}

func web(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("You Can Use Website\nUser ID: `%d`\nPassword: `%s`", u.UserInfo.UserID, u.Passwd()))
	msg.ReplyToMessageID = Message.MessageID
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = clictUrlButton(u, fmt.Sprintf("https://%s/login", parseflag.WEBSITE))
	bot.Send(msg)
}

func superresolution(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	if u.Permissions() == user.T_Prohibit {
		msg := tgbotapi.NewMessage(u.ChatMember.User.ID, u.ProhibitString(bot))
		msg.ReplyMarkup = goJoinButton(u)
		msg.ReplyToMessageID = Message.MessageID
		bot.Send(msg)
		return
	}
	task, err := u.AddTask(user.T_SuperResolution)
	if err != nil {
		return
	}
	defer task.Down()
	m2, err := bot.NewMsgCbk(Message.Chat.ID, Message.From.ID)
	if err != nil {
		colorlog.Errorf("NewMsgCbk err [%s] : %v", Message.From.String(), err)
		return
	}
	defer m2.Close()
	var resize int64 = 4
	mc := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\n↓ %s ↓", u.LoadLang("sendImg"), u.LoadLang("magnification")))
	mc.ReplyToMessageID = Message.MessageID
	mc.ReplyMarkup = sprButton(u, int(resize))
	m, err := bot.Send(mc)
	if err != nil {
		return
	}
	c, err := bot.NewCbk(Message.Chat.ID, Message.From.ID, m.MessageID)
	if err != nil {
		colorlog.Errorf("NewCbk err [%s] : %v", Message.From.String(), err)
		return
	}
	defer c.Close()
	var msg *tgbotapi.Message
	var ok bool
	t := time.NewTimer(time.Minute * 3)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m.MessageID))
			return
		case data, ok := <-c.Chan():
			if !ok || data.Value == "cancel" {
				bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m.MessageID))
				return
			}
			resize, err = strconv.ParseInt(data.Value, 10, 64)
			if err != nil {
				bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m.MessageID))
				return
			}
			bot.Send(tgbotapi.NewEditMessageReplyMarkup(Message.Chat.ID, m.MessageID, *sprButton(u, int(resize))))
		case msg, ok = <-m2.MsgChan():
			bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m.MessageID))
			if !ok {
				return
			}
			c.Close()
			m2.Close()
			goto RUN
		}
	}
RUN:
	cfg := new(Config)
	cfg.DrawConfig = api.DrawConfig{}
	var prePhoto []byte
	if len(msg.Photo) > 0 {
		latestPhoto := msg.Photo[len(msg.Photo)-1]
		cfg.Width = int(latestPhoto.Width)
		cfg.Height = int(latestPhoto.Height)
		if cfg.Width*cfg.Height > 4194304 {
			colorlog.Errorf("Size err [%s] : %d", u.ChatMember.User.String(), cfg.Width*cfg.Height)
			m := tgbotapi.NewMessage(u.ChatMember.User.ID, u.LoadLang("bigImg"))
			m.ReplyToMessageID = msg.MessageID
			bot.Send(m)
			return
		}
		prePhoto, err = bot.GetFileData(latestPhoto.FileID)
		if err != nil {
			colorlog.Errorf("Get File Data err [%s] : %v", u.ChatMember.User.String(), err)
			return
		}
		fi, err := cache.Put(prePhoto)
		if err != nil {
			colorlog.Errorf("Put file err [%s] : %v", u.ChatMember.User.String(), err)
		}
		cfg.PrePhotoID = fi.Md5
	} else if msg.Document != nil {
		if _, ok := utils.InString(msg.Document.MimeType, avilableDocumentType); !ok {
			colorlog.Errorf("Document Type err [%s] : %s", u.ChatMember.User.String(), msg.Document.MimeType)
			return
		}
		if prePhoto, err = cache.GetFile(utils.GetFileNamePrefix(msg.Document.FileName)); err == nil {
			cfg.PrePhotoID = utils.GetFileNamePrefix(msg.Document.FileName)
		} else {
			prePhoto, err = bot.GetFileData(msg.Document.FileID)
			if err != nil {
				colorlog.Error(err)
				return
			}
			fi, err := cache.Put(prePhoto)
			if err != nil {
				colorlog.Errorf("Put file err: %v", err)
			}
			cfg.PrePhotoID = fi.Md5
		}
		cfg.Width, cfg.Height, err = utils.GetPhotoSize(prePhoto)
		if err != nil {
			colorlog.Errorf("Compress Image err [%s] : %v", u.ChatMember.User.String(), err)
			return
		}
		if cfg.Width*cfg.Height > 4194304 {
			colorlog.Errorf("Size err [%s] : %d", u.ChatMember.User.String(), cfg.Width*cfg.Height)
			m := tgbotapi.NewMessage(u.ChatMember.User.ID, u.LoadLang("bigImg"))
			m.ReplyToMessageID = msg.MessageID
			bot.Send(m)
			return
		}
	} else {
		return
	}
	superResolutionRun(bot, u, msg.MessageID, cfg, prePhoto, int(resize))
}

func dev(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	if Message.From.ID != parseflag.MyID {
		return
	}
	parseflag.Dev = !parseflag.Dev
	bot.Send(tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprint(parseflag.Dev)))
}

func token(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	token := Message.CommandArguments()
	if len(token) != 64 {
		return
	}
	tk := db.Token{}
	db.DB().Where("token = ?", token).Find(&tk)
	if tk.ValidDate == 0 {
		tokenErrMsg := tgbotapi.NewMessage(Message.Chat.ID, u.LoadLang("tokenErr"))
		bot.Send(tokenErrMsg)
		return
	}
	add, _ := time.ParseDuration(fmt.Sprintf("%dh", tk.ValidDate*24))
	now := time.Now()
	if u.Subscribe.Deadline.Before(now) {
		u.Subscribe.Deadline = now.Add(add)
	} else {
		u.Subscribe.Deadline = u.Subscribe.Deadline.Add(add)
	}
	db.DB().Model(&db.Subscribe{}).Where("user_id = ?", Message.From.ID).Update("deadline", u.Subscribe.Deadline)
	db.DB().Where("token = ?", tk.Token).Delete(&tk)
	mc := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("Success! Deadline: `%s`", u.Subscribe.Deadline.Format("2006-01-02 15:04:05")))
	mc.ReplyToMessageID = Message.MessageID
	mc.ParseMode = "Markdown"
	bot.Send(mc)
}

func getToken(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	if Message.From.ID != parseflag.MyID {
		return
	}
	u, err := strconv.ParseUint(Message.CommandArguments(), 10, 64)
	if err != nil || u == 0 {
		return
	}
	token := utils.RandomString(64)
	d := db.DB().Create(&db.Token{Token: token, ValidDate: u})
	var msg tgbotapi.MessageConfig
	if d.Error != nil {
		msg = tgbotapi.NewMessage(Message.From.ID, err.Error())
	} else {
		msg = tgbotapi.NewMessage(Message.From.ID, fmt.Sprintf("`/token %s`", token))
		msg.ParseMode = "Markdown"
	}
	bot.Send(msg)
}

func drawPool(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	if Message.From.ID != parseflag.MyID {
		return
	}
	msg := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("```\npool: %d\nfree: %d\nwait: %d\ntime: %v\n```", api.DrawPoolCap(), api.DrawFree(), api.DrawWait(), time.Now().Format("01-02 15:04:05")))
	msg.ReplyMarkup = poolButton
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func img2tag(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	m, err := bot.NewMsgCbk(Message.Chat.ID, Message.From.ID)
	if err != nil {
		return
	}
	defer m.Close()
	mc := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\n%s\n%s", u.LoadLang("sendImg"), u.LoadLang("wait"), u.LoadLang("dontDelMsg")))
	mc.ReplyMarkup = cancelButton(u)
	mc.ReplyToMessageID = Message.MessageID
	m2, err := bot.Send(mc)
	if err != nil {
		return
	}
	c, err := bot.NewCbk(Message.Chat.ID, Message.From.ID, m2.MessageID)
	if err != nil {
		return
	}
	defer c.Close()
	t := time.NewTimer(time.Minute * 5)
	defer t.Stop()
	select {
	case <-t.C:
		bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m2.MessageID))
		return
	case <-c.Chan():
		bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m2.MessageID))
		return
	case msg, ok := <-m.MsgChan():
		m.Close()
		if !ok {
			return
		}
		if msg.Document == nil || msg.Document.MimeType != "image/png" {
			colorlog.Errorf("Parse Photo err: Document is nil or MimeType err")
			bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, u.LoadLang("parsePhotoErr")))
			return
		}
		var photo []byte
		if photo, err = cache.GetFile(utils.GetFileNamePrefix(msg.Document.FileName)); err != nil {
			photo, err = bot.GetFileData(msg.Document.FileID)
			if err != nil {
				colorlog.Errorf("Parse Photo err: %v", err)
				bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, u.LoadLang("parsePhotoErr")))
				return
			}
		}
		dc, err := api.GetImgCfg(photo)
		if err != nil {
			colorlog.Errorf("Parse Photo err: %v", err)
			bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, u.LoadLang("parsePhotoErr")))
			return
		}
		ms := tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, string((&Config{DrawConfig: *dc}).Fomate2TgHTML()))
		ms.ReplyMarkup = reDrawButton(u)
		ms.ParseMode = "HTML"
		bot.Send(ms)
	}
}

func guessTag(bot *tgbotapi.BotAPI, Message tgbotapi.Message, u *user.UserInfo) {
	task, err := u.AddTask(user.T_GuessTag)
	if err != nil {
		return
	}
	defer task.Down()
	m, err := bot.NewMsgCbk(Message.Chat.ID, Message.From.ID)
	if err != nil {
		return
	}
	defer m.Close()
	mc := tgbotapi.NewMessage(Message.Chat.ID, fmt.Sprintf("%s\n%s\n%s", u.LoadLang("sendImg"), u.LoadLang("wait"), u.LoadLang("dontDelMsg")))
	mc.ReplyMarkup = cancelButton(u)
	mc.ReplyToMessageID = Message.MessageID
	m2, err := bot.Send(mc)
	if err != nil {
		return
	}
	c, err := bot.NewCbk(Message.Chat.ID, Message.From.ID, m2.MessageID)
	if err != nil {
		return
	}
	defer c.Close()
	t := time.NewTimer(time.Minute * 5)
	defer t.Stop()
	select {
	case <-t.C:
		bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m2.MessageID))
		return
	case <-c.Chan():
		bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m2.MessageID))
		return
	case msg, ok := <-m.MsgChan():
		m.Close()
		if !ok {
			return
		}
		cfg := u.DefaultConfig()
		var photo []byte
		if len(msg.Photo) > 0 {
			latestPhoto := msg.Photo[len(msg.Photo)-1]
			cfg.Width = latestPhoto.Width
			cfg.Height = latestPhoto.Height
			photo, err = bot.GetFileData(latestPhoto.FileID)
			if err != nil {
				colorlog.Errorf("Guess Tag err: %v", err)
				bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "Something Error"))
				return
			}
		} else if msg.Document != nil {
			if _, ok := utils.InString(msg.Document.MimeType, avilableDocumentType); !ok {
				colorlog.Errorf("Guess Tag err: %v", "document type is not avilable")
				bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "document type is not avilable"))
				return
			}
			photo, err = bot.GetFileData(msg.Document.FileID)
			if err != nil {
				colorlog.Errorf("Guess Tag err: %v", err)
				bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "Something Error"))
				return
			}
			cfg.Width, cfg.Height, err = utils.GetPhotoSize(photo)
			if err != nil {
				colorlog.Errorf("Guess Tag err: %v", err)
				bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "Something Error"))
				return
			}
		} else {
			bot.Send(tgbotapi.NewDeleteMessage(Message.Chat.ID, m2.MessageID))
			return
		}
		ic, err := api.NewInterrogate(photo)
		if err != nil {
			colorlog.Errorf("Guess Tag err: %v", err)
			bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "Something Error"))
			return
		}
		c2 := ic.Interrogate(context.Background())
		ret, ok := <-c2
		if !ok {
			bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "Something Error"))
			return
		}
		if ret.Err != nil {
			colorlog.Errorf("Guess Tag err: %v", ret.Err)
			bot.Send(tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, "Something Error"))
			return
		}
		cfg.Tag = ret.Resoult
		cfg.CorrectCfg(true, false, false, false, false, false, false)
		ms := tgbotapi.NewEditMessageText(Message.Chat.ID, m2.MessageID, string((&Config{DrawConfig: *cfg}).Fomate2TgHTML()))
		ms.ParseMode = "HTML"
		ms.ReplyMarkup = reDrawButton(u)
		bot.Send(ms)
	}
}

func setLanguage(bot *tgbotapi.BotAPI, Message tgbotapi.Message) {
	mc := tgbotapi.NewMessage(Message.Chat.ID, "set language:")
	mc.ReplyToMessageID = Message.MessageID
	mc.ReplyMarkup = langButton
	bot.Send(mc)
}
