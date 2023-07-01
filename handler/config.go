package handler

import (
	"errors"
	"math"
	"math/rand"
	"regexp"
	"strings"

	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"
)

const (
	GuestImgMaxSize = 737280
)

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
	if k := utils.In(h.ControlPreProcess, func(c gconfig.ControlPreProcess) bool {
		return c.Name == PreProcess
	}); k == -1 {
		return h.ControlPreProcess[0].Name
	}
	return ""
}

func (h *Handler) ParseProcess(Process string) string {
	if Process == "" {
		return h.ControlProcess[0].Name
	}
	if k := utils.In(h.ControlProcess, func(c gconfig.ControlProcess) bool {
		return c.Name == Process
	}); k == -1 {
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
	Strength  bool
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

func WithStrength() ConfigFuncCorrentCfg {
	return func(c *CorrectConfig) {
		c.Strength = true
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

func (h *Handler) CorrectCfg(cfg *db.Config, u *UserInfo, c ...ConfigFuncCorrentCfg) {
	config := &CorrectConfig{}
	for _, f := range c {
		f(config)
	}
	if cfg.ControlPhotoID == "" || cfg.ControlPreprocess == "" || cfg.ControlProcess == "" {
		cfg.ControlPreprocess = ""
		cfg.ControlProcess = ""
	}
	if config.CtrlPhoto {
		cfg.ControlPreprocess = h.ParsePreProcess(cfg.ControlPreprocess)
		cfg.ControlProcess = h.ParseProcess(cfg.ControlProcess)
	}
	if cfg.PrePhotoID == "" {
		cfg.Strength = 0
	}
	if config.Strength && (cfg.Strength <= 0 || cfg.Strength >= 1) {
		cfg.Strength = 0.70
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
		if k := utils.In(h.mode, func(s string) bool {
			return cfg.Mode == s
		}); k == -1 {
			cfg.Mode = h.mode[0]
		}
	}
}

func (h *Handler) DefaultConfig() *db.Config {
	cfg := &db.Config{Width: 512,
		Height:   768,
		Num:      h.DefaultNum,
		Mode:     h.DefaultMode,
		Steps:    h.DefaultSteps,
		CfgScale: h.DefaultCfgScale,
		Uc:       h.DefaultUC,
		Model:    h.Models[0].Name}
	h.CorrectCfg(cfg, nil, WithTag(), WithUc())
	return cfg
}

func (h *Handler) Name2Process(name string) (gconfig.ControlProcess, error) {
	for _, v := range h.ControlProcess {
		if v.Name == name {
			return v, nil
		}
	}
	return gconfig.ControlProcess{}, errors.New("cannot find process")
}
