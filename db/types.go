package db

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/utils"
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
	FileID  string `gorm:"unique;not null"`
	UserID  int64  `gorm:"not null"`
	UnShare bool   `gorm:"default:false"`
	Config  Config `gorm:"embedded"`
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

type Config struct {
	Tag               string  `json:"tag,omitempty" yaml:"tag,omitempty"`
	Mode              string  `json:"mode,omitempty" yaml:"mode,omitempty"`
	Num               int     `json:"num,omitempty" yaml:"num,omitempty"`
	Steps             int     `json:"steps,omitempty" yaml:"steps,omitempty"`
	Seed              uint32  `json:"seed,omitempty" yaml:"seed,omitempty"`
	CfgScale          int     `json:"scale,omitempty" yaml:"scale,omitempty"`
	Width             int     `json:"width" yaml:"width"`
	Height            int     `json:"height" yaml:"height"`
	Model             string  `json:"model,omitempty" yaml:"model,omitempty"`
	Uc                string  `json:"uc,omitempty" yaml:"uc,omitempty"`
	PrePhotoID        string  `json:"pre_photo_id,omitempty" yaml:"pre_photo_id,omitempty"`
	Strength          float64 `json:"strength,omitempty" yaml:"strength,omitempty"`
	ControlPhotoID    string  `json:"control_photo_id,omitempty" yaml:"control_photo_id,omitempty"`
	ControlPreprocess string  `json:"control_preprocess,omitempty" yaml:"control_preprocess,omitempty"`
	ControlProcess    string  `json:"control_process,omitempty" yaml:"control_process,omitempty"`
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

var replaceColonRe = regexp.MustCompile(`: *`)

func ReplaceColon(s string) string {
	return replaceColonRe.ReplaceAllString(s, ":")
}

func parseString2YamlStyle(s string) string {
	for (strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) || (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) {
		s = s[1 : len(s)-1]
	}
	return strings.TrimLeft(ReplaceColon(s), `'"`)
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

func (c *Config) Fomate2TgHTML() []byte {
	var buffer bytes.Buffer
	if c.Tag != "" {
		buffer.WriteString(fmt.Sprintf("<b>tag:</b> <code>%s</code>\n", parseString2YamlStyle(utils.Parse2HTML(c.Tag))))
	}
	if c.Mode != "" {
		buffer.WriteString(fmt.Sprintf("<b>mode:</b> <code>%s</code>\n", utils.Parse2HTML(c.Mode)))
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
		buffer.WriteString(fmt.Sprintf("<b>model:</b> <code>%s</code>\n", utils.Parse2HTML(c.Model)))
	}
	if c.Uc != "" {
		buffer.WriteString(fmt.Sprintf("<b>uc:</b> <code>%s</code>\n", parseString2YamlStyle(utils.Parse2HTML(c.Uc))))
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
		buffer.WriteString(fmt.Sprintf("<b>control_preprocess:</b> <code>%s</code>\n", utils.Parse2HTML(c.ControlPreprocess)))
	}
	if c.ControlProcess != "" {
		buffer.WriteString(fmt.Sprintf("<b>control_process:</b> <code>%s</code>\n", utils.Parse2HTML(c.ControlProcess)))
	}
	return buffer.Bytes()
}
