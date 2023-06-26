package handler

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"

	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

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

func (h *Handler) ParseCfgScalse(scale int) int {
	if scale <= 0 || scale > 30 {
		return h.DefaultCfgScale
	}
	return scale
}

func (h *Handler) ParseSteps(steps int) int {
	if steps < 15 || steps > 50 {
		return h.DefaultSteps
	}
	return steps
}

func (h *Handler) ParseNum(num int) int {
	if num <= 0 || num > h.MaxNum {
		return h.DefaultNum
	}
	return num
}

func (h *Handler) ParsePreProcess(PreProcess string) string {
	if PreProcess == "" {
		return h.ControlPreProcess[0].Name
	}
	if _, ok := utils.In(h.ControlPreProcess, func(c gconfig.ControlPreProcess) bool {
		return c.Name == PreProcess
	}); !ok {
		return h.ControlPreProcess[0].Name
	}
	return ""
}

func (h *Handler) ParseProcess(Process string) string {
	if Process == "" {
		return h.ControlProcess[0].Name
	}
	if _, ok := utils.In(h.ControlProcess, func(c gconfig.ControlProcess) bool {
		return c.Name == Process
	}); !ok {
		return h.ControlProcess[0].Name
	}
	return ""
}

func (h *Handler) MODELFILETONAME(file string) (string, error) {
	for _, m := range h.Models {
		if m.File == file {
			return m.Name, nil
		}
	}
	return "", errors.New("find models error")
}

func (h *Handler) Name2Model(name string) (gconfig.Model, error) {
	for _, v := range h.Models {
		if v.Name == name {
			return v, nil
		}
	}
	return gconfig.Model{}, errors.New("find models error")
}

type CorrectConfig struct {
	Tag       bool
	Uc        bool
	Photo     bool
	CtrlPhoto bool
	TransTag  bool
	TransUc   bool
	Seed      bool
	Mode      bool
	Model     bool
}

type ConfigFuncCorrentCfg func(*CorrectConfig)

func WithMode() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Mode = true
	}
}

func WithModel() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Model = true
	}
}

func WithTag() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Tag = true
	}
}

func WithUc() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Uc = true
	}
}

func WithPhoto() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Photo = true
	}
}

func WithCtrlPhoto() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.CtrlPhoto = true
	}
}

func WithTransTag() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.TransTag = true
	}
}

func WithTransUc() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.TransUc = true
	}
}

func WithSeed() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Seed = true
	}
}

var parseRepeat, _ = regexp.Compile(`, *?,`)
var replaceColonRe = regexp.MustCompile(`: *`)

func ReplaceColon(s string) string {
	return replaceColonRe.ReplaceAllString(s, ":")
}

func ReplaceString(src string) string {
	if src == "" {
		return ""
	}
	src = strings.ReplaceAll(src, "，", ",")
	src = strings.ReplaceAll(src, "\n", ",")
	src = strings.ReplaceAll(src, "（", "(")
	src = strings.ReplaceAll(src, "）", ")")
	for parseRepeat.MatchString(src) {
		src = parseRepeat.ReplaceAllString(src, ",")
	}
	src = strings.TrimLeft(strings.TrimRight(src, ", "), ", ")
	return replaceColonRe.ReplaceAllString(src, ":")
}

func generateUC(Uc string) string {
	Uc = ReplaceString(Uc)
	low := strings.ToLower(Uc)
	if !strings.Contains(low, `lowres`) {
		Uc = `lowres, ` + Uc
	}
	if !strings.Contains(low, `text`) {
		Uc = `text, ` + Uc
	}
	return Uc
}

func generateTag(Tag string) string {
	Tag = ReplaceString(Tag)
	low := strings.ToLower(Tag)
	if !strings.Contains(low, "best quality") {
		Tag = "best quality, " + Tag
	}
	if !strings.Contains(low, "masterpiece") {
		Tag = "masterpiece, " + Tag
	}
	return Tag
}

func (h *Handler) CorrectCfg(cfg *Config, u *UserInfo, c ...ConfigFuncCorrentCfg) {
	config := &CorrectConfig{}
	for _, f := range c {
		f(config)
	}
	if config.CtrlPhoto {
		if cfg.ControlPhotoID == "" || cfg.ControlPreprocess == "" || cfg.ControlProcess == "" {
			cfg.ControlPreprocess = ""
			cfg.ControlProcess = ""
		} else {
			cfg.ControlPreprocess = h.ParsePreProcess(cfg.ControlPreprocess)
			cfg.ControlProcess = h.ParseProcess(cfg.ControlProcess)
		}
	}
	if config.Photo {
		if cfg.PrePhotoID == "" {
			cfg.Strength = 0
		} else if cfg.Strength < 0 || cfg.Strength >= 1 {
			cfg.Strength = 0.70
		}
	}
	cfg.CfgScale = h.ParseCfgScalse(cfg.CfgScale)
	if config.Seed && cfg.Seed == 0 {
		cfg.Seed = uint32(rand.Intn(math.MaxUint32))
	}
	if config.Model {
		m, err := h.Name2Model(cfg.Model)
		if err != nil {
			cfg.Model = h.Models[0].Name
		} else {
			cfg.Model = m.Name
		}
	}
	if config.Tag {
		cfg.Tag = generateTag(cfg.Tag)
	}
	if config.TransTag {
		cfg.Tag = utils.Translate(cfg.Tag)
	}
	if config.Uc {
		cfg.Uc = generateUC(cfg.Uc)
	}
	if config.TransUc {
		cfg.Uc = utils.Translate(cfg.Uc)
	}
	if u == nil {
		cfg.Steps = h.ParseSteps(cfg.Steps)
		cfg.Num = h.ParseNum(cfg.Num)
		if cfg.Height < 64 || cfg.Width < 64 {
			cfg.Width = 512
			cfg.Height = 768
		} else if sum := cfg.Height * cfg.Width; sum > h.ImgMaxSize {
			a := math.Pow(float64(sum)/float64(h.ImgMaxSize), 0.5)
			cfg.Width = int(float64(cfg.Width) / a)
			cfg.Height = int(float64(cfg.Height) / a)
		}
	} else if u.Permissions() != T_Subscribe {
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
	}
	cfg.Width -= cfg.Width % 8
	cfg.Height -= cfg.Height % 8
	if config.Mode {
		if _, ok := utils.InString(cfg.Mode, h.mode); !ok {
			cfg.Mode = h.mode[0]
		}
	}
}

func (h *Handler) getConfig(u *UserInfo, cfg *Config, replyMsgId int) (err error) {
	if cfg == nil {
		return errors.New("cfg is nil")
	}
	panel := panelButton(u, cfg.PrePhotoID != "", cfg.ControlPhotoID != "")
	h.CorrectCfg(cfg, u)
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
	_, err = h.bot.Send(mc)
	return err
}

func (h *Handler) DefaultConfig() *api.DrawConfig {
	return &api.DrawConfig{
		Width:    512,
		Height:   768,
		Num:      1,
		Strength: 0.70,
		Mode:     h.DefaultMode,
		Steps:    h.DefaultSteps,
		CfgScale: h.DefaultCfgScale,
		Uc:       h.DefaultUC,
		Model:    h.Models[0].Name,
	}
}

func (h *Handler) Name2Process(name string) (gconfig.ControlProcess, error) {
	for _, v := range h.ControlProcess {
		if v.Name == name {
			return v, nil
		}
	}
	return gconfig.ControlProcess{}, errors.New("cannot find process")
}
