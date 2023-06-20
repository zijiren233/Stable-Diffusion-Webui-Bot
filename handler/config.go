package handler

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/user"

	tgbotapi "github.com/zijiren233/tg-bot-api/v6"
)

const (
	GuestImgMaxSize = 737280
)

type Config struct {
	api.DrawConfig `yaml:",inline"`
	PrePhotoID     string `json:"pre_photo_id,omitempty" yaml:"pre_photo_id,omitempty"`
	ControlPhotoID string `json:"control_photo_id,omitempty" yaml:"control_photo_id,omitempty"`
}

const markdownV2Reg = "[_*\\[\\]()~`>#+\\-=|{}\\\\\\.!]"
const restr2String = `[\*\.\?\+\$\^\[\]\(\)\{\}\|\\]`

var markdownV2Re = regexp.MustCompile(markdownV2Reg)
var restr2StringRe = regexp.MustCompile(restr2String)

func parseRestr2String(s string) string {
	return restr2StringRe.ReplaceAllString(s, `\$0`)
}

func parse2MarkdownV2(s string) string {
	return markdownV2Re.ReplaceAllString(s, `\$0`)
}

func parseString2YamlStyle(s string) string {
	for (strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) || (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) {
		s = s[1 : len(s)-1]
	}
	return strings.TrimLeft(api.ReplaceColon(s), `'"`)
}

func (c *Config) Fomate2TgMdV2() []byte {
	var buffer bytes.Buffer
	if c.Tag != "" {
		buffer.WriteString(fmt.Sprintf("*tag:* `%s`\n", parseString2YamlStyle(markdownV2Re.ReplaceAllString(c.Tag, `\$0`))))
	}
	if c.Mode != "" {
		buffer.WriteString(fmt.Sprintf("*mode:* `%s`\n", markdownV2Re.ReplaceAllString(c.Mode, `\$0`)))
	}
	if c.Num != 0 {
		buffer.WriteString(fmt.Sprintf("*num:* %d\n", c.Num))
	}
	if c.Steps != 0 {
		buffer.WriteString(fmt.Sprintf("*steps:* %d\n", c.Steps))
	}
	if c.Seed != 0 {
		buffer.WriteString(fmt.Sprintf("*seed:* `%d`\n", c.Seed))
	}
	if c.CfgScale != 0 {
		buffer.WriteString(fmt.Sprintf("*scale:* %d\n", c.CfgScale))
	}
	if c.Width != 0 {
		buffer.WriteString(fmt.Sprintf("*width:* %d\n", c.Width))
	}
	if c.Height != 0 {
		buffer.WriteString(fmt.Sprintf("*height:* %d\n", c.Height))
	}
	if c.Model != "" {
		buffer.WriteString(fmt.Sprintf("*model:* `%s`\n", markdownV2Re.ReplaceAllString(c.Model, `\$0`)))
	}
	if c.Uc != "" {
		buffer.WriteString(fmt.Sprintf("*uc:* `%s`\n", parseString2YamlStyle(markdownV2Re.ReplaceAllString(c.Uc, `\$0`))))
	}
	if c.PrePhotoID != "" {
		buffer.WriteString(fmt.Sprintf("*pre\\_photo\\_id:* `%s`\n", c.PrePhotoID))
	}
	if c.Strength != 0 {
		buffer.WriteString(fmt.Sprintf("*strength:* %s\n", markdownV2Re.ReplaceAllString(fmt.Sprint(c.Strength), `\$0`)))
	}
	if c.ControlPhotoID != "" {
		buffer.WriteString(fmt.Sprintf("*control\\_photo\\_id:* `%s`\n", c.ControlPhotoID))
	}
	if c.ControlPreprocess != "" {
		buffer.WriteString(fmt.Sprintf("*control\\_preprocess:* `%s`\n", markdownV2Re.ReplaceAllString(c.ControlPreprocess, `\$0`)))
	}
	if c.ControlProcess != "" {
		buffer.WriteString(fmt.Sprintf("*control\\_process:* `%s`\n", markdownV2Re.ReplaceAllString(c.ControlProcess, `\$0`)))
	}
	return buffer.Bytes()
}

var re = regexp.MustCompile(`[<>&]`)

func parse2HTML(str string) string {
	return re.ReplaceAllStringFunc(str, func(s string) string {
		switch s {
		case "<":
			return "&lt;"
		case ">":
			return "&gt;"
		case "&":
			return "&amp;"
		default:
			return s
		}
	})
}

func (c *Config) Fomate2TgHTML() []byte {
	var buffer bytes.Buffer
	if c.Tag != "" {
		buffer.WriteString(fmt.Sprintf("<b>tag:</b> <code>%s</code>\n", parseString2YamlStyle(parse2HTML(c.Tag))))
	}
	if c.Mode != "" {
		buffer.WriteString(fmt.Sprintf("<b>mode:</b> <code>%s</code>\n", parse2HTML(c.Mode)))
	}
	if c.Num != 0 {
		buffer.WriteString(fmt.Sprintf("<b>num:</b> %d\n", c.Num))
	}
	if c.Steps != 0 {
		buffer.WriteString(fmt.Sprintf("<b>steps:</b> %d\n", c.Steps))
	}
	if c.Seed != 0 {
		buffer.WriteString(fmt.Sprintf("<b>seed:</b> <code>%d</code>\n", c.Seed))
	}
	if c.CfgScale != 0 {
		buffer.WriteString(fmt.Sprintf("<b>scale:</b> %d\n", c.CfgScale))
	}
	if c.Width != 0 {
		buffer.WriteString(fmt.Sprintf("<b>width:</b> %d\n", c.Width))
	}
	if c.Height != 0 {
		buffer.WriteString(fmt.Sprintf("<b>height:</b> %d\n", c.Height))
	}
	if c.Model != "" {
		buffer.WriteString(fmt.Sprintf("<b>model:</b> <code>%s</code>\n", parse2HTML(c.Model)))
	}
	if c.Uc != "" {
		buffer.WriteString(fmt.Sprintf("<b>uc:</b> <code>%s</code>\n", parseString2YamlStyle(parse2HTML(c.Uc))))
	}
	if c.PrePhotoID != "" {
		buffer.WriteString(fmt.Sprintf("<b>pre_photo_id:</b> <code>%s</code>\n", c.PrePhotoID))
	}
	if c.Strength != 0 {
		buffer.WriteString(fmt.Sprintf("<b>strength:</b> %s\n", fmt.Sprintf("%.2f", c.Strength)))
	}
	if c.ControlPhotoID != "" {
		buffer.WriteString(fmt.Sprintf("<b>control_photo_id:</b> <code>%s</code>\n", c.ControlPhotoID))
	}
	if c.ControlPreprocess != "" {
		buffer.WriteString(fmt.Sprintf("<b>control_preprocess:</b> <code>%s</code>\n", parse2HTML(c.ControlPreprocess)))
	}
	if c.ControlProcess != "" {
		buffer.WriteString(fmt.Sprintf("<b>control_process:</b> <code>%s</code>\n", parse2HTML(c.ControlProcess)))
	}
	return buffer.Bytes()
}

func (cfg *Config) CorrectCfg(u *user.UserInfo, gTag, gUc, transTag, transUc, gSeed bool) {
	if u.Permissions() != user.T_Subscribe {
		if cfg.Steps > 28 {
			cfg.Steps = 28
		}
		if cfg.Num > 3 {
			cfg.Num = 3
		}
		if sum := cfg.Height * cfg.Width; sum > GuestImgMaxSize {
			a := math.Pow(float64(sum)/GuestImgMaxSize, 0.5)
			cfg.Width = int(float64(cfg.Width) / a)
			cfg.Height = int(float64(cfg.Height) / a)
		}
		cfg.Width -= cfg.Width % 8
		cfg.Height -= cfg.Height % 8
	}
	if len(cfg.ControlPhotoID) != 32 {
		cfg.ControlPhotoID = ""
	}
	if len(cfg.PrePhotoID) != 32 {
		cfg.PrePhotoID = ""
	}
	cfg.DrawConfig.CorrectCfg(gTag, gUc, len(cfg.PrePhotoID) == 32, len(cfg.ControlPhotoID) == 32, transTag, transUc, gSeed)
}

func getConfig(bot *tgbotapi.BotAPI, u *user.UserInfo, cfg *Config, replyMsgId int) (err error) {
	if cfg == nil {
		return errors.New("cfg is nil")
	}
	panel := panelButton(u, len(cfg.PrePhotoID) == 32, len(cfg.ControlPhotoID) == 32)
	cfg.CorrectCfg(u, true, true, false, false, true)
	// b, err := yaml.Marshal(cfg)
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("string(b): %v\n", string(b))
	mc := tgbotapi.NewMessage(u.ChatMember.User.ID, string(cfg.Fomate2TgHTML()))
	mc.ReplyMarkup = panel
	mc.ReplyToMessageID = replyMsgId
	mc.ParseMode = "HTML"
	mc.DisableWebPagePreview = false
	_, err = bot.Send(mc)
	return err
}
