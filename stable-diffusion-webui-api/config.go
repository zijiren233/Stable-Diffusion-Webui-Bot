package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strings"

	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"
)

type drawType uint

const (
	T_tag2img drawType = iota
	T_img2img
	T_superResolution
)

type config struct {
	ctx         context.Context
	resoultChan chan<- *Resoult
	cfg         *drawConfig
	rawCfg      DrawConfig
	api         struct {
		api    *apiUrl
		status Status
	}
}

type Status struct {
	Progress    float64 `json:"progress"`
	EtaRelative float64 `json:"eta_relative"`
}

type superResolutionCfg struct {
	resoultChan               chan<- *Resoult `json:"-"`
	ctx                       context.Context
	UpscalingResize           int     `json:"upscaling_resize"`
	Upscaler1                 string  `json:"upscaler_1"`
	ExtrasUpscaler2Visibility float64 `json:"extras_upscaler_2_visibility"`
	ImageList                 []struct {
		Data string `json:"data"`
		Name string `json:"name"`
	} `json:"imageList"`
}

type ctrlPhotoCfg struct {
	resoultChan            chan<- *Resoult `json:"-"`
	ctx                    context.Context
	ControlnetModule       string   `json:"controlnet_module"`
	ControlnetInputImages  []string `json:"controlnet_input_images"`
	ControlnetProcessorRes int      `json:"controlnet_processor_res"`
	ControlnetThresholdA   int      `json:"controlnet_threshold_a"`
	ControlnetThresholdB   int      `json:"controlnet_threshold_b"`
}

type InterrogateResoult struct {
	Err     error
	Resoult string
}

type interrogateCfg struct {
	resoultChan chan<- *InterrogateResoult `json:"-"`
	ctx         context.Context
	Image       string `json:"image"`
	Model       string `json:"model"`
}

type DrawConfig struct {
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
	Strength          float64 `json:"strength,omitempty" yaml:"strength,omitempty"`
	ControlPreprocess string  `json:"control_preprocess,omitempty" yaml:"control_preprocess,omitempty"`
	ControlProcess    string  `json:"control_process,omitempty" yaml:"control_process,omitempty"`
}

type drawConfig struct {
	ResizeMode        int      `json:"resize_mode,omitempty"`
	EnableHr          bool     `json:"enable_hr,omitempty"`
	HrScale           float64  `json:"hr_scale,omitempty"`
	HrUpscaler        string   `json:"hr_upscaler,omitempty"`
	HrSecondPassSteps int      `json:"hr_second_pass_steps,omitempty"`
	InitImages        []string `json:"init_images,omitempty"`
	DenoisingStrength float64  `json:"denoising_strength"`
	Styles            []string `json:"styles,omitempty"`
	Prompt            string   `json:"prompt"`
	Count             int      `json:"n_iter,omitempty"`
	Num               int      `json:"batch_size,omitempty"`
	Seed              uint32   `json:"seed"`
	Steps             int      `json:"steps"`
	CfgScale          int      `json:"cfg_scale"`
	Width             int      `json:"width"`
	Height            int      `json:"height"`
	NegativePrompt    string   `json:"negative_prompt"`
	SamplerName       string   `json:"sampler_name"`
	SamplerIndex      string   `json:"sampler_index"`
	AlwaysonScripts   struct {
		Controlnet struct {
			Args []ControlnetUnits `json:"args,omitempty"`
		} `json:"controlnet,omitempty"`
	} `json:"alwayson_scripts,omitempty"`
}

type ControlnetUnits struct {
	InputImage    string `json:"input_image,omitempty"`
	Mask          string `json:"mask,omitempty"`
	Module        string `json:"module,omitempty"`
	Model         string `json:"model,omitempty"`
	Weight        int    `json:"weight,omitempty"`
	ResizeMode    string `json:"resize_mode,omitempty"`
	Lowvram       bool   `json:"lowvram,omitempty"`
	ProcessorRes  int    `json:"processor_res,omitempty"`
	ThresholdA    int    `json:"threshold_a,omitempty"`
	ThresholdB    int    `json:"threshold_b,omitempty"`
	Guidance      int    `json:"guidance,omitempty"`
	GuidanceStart int    `json:"guidance_start,omitempty"`
	GuidanceEnd   int    `json:"guidance_end,omitempty"`
	Guessmode     bool   `json:"guessmode,omitempty"`
}

var allMode = [...]string{"DPM++ 2M Karras", "DPM++ 2M SDE Karras", "DPM++ SDE Karras", "Euler a", "DPM2", "DPM adaptive", "DPM2 a Karras", "DPM2 Karras", "DPM++ 2M", "DPM++ 2S a", "DPM++ 2S a Karras", "DPM++ SDE", "LMS Karras", "Euler", "DDIM", "Heun", "UniPC"}
var parseRepeat, _ = regexp.Compile(`, *?,`)
var Ucmap = map[string]string{
	"low quality": "cropped, worst quality, low quality, normal quality, jpeg artifacts, signature, watermark, username, blurry",
	"bad anatomy": "bad anatomy, bad hands, error, missing fingers, extra digit, fewer digits",
}

var DefaultMode = allMode[0]

const (
	MaxHFSteps      = 20
	MaxNum          = 9
	DefaultCfgScale = 9
	DefaultSteps    = 20
	DefaultNum      = 1
)

var defaultUC = fmt.Sprint("lowres, text, ", Ucmap["bad anatomy"], ", ", Ucmap["low quality"])

func DefauleUC() string {
	return defaultUC
}

func AllMode() [len(allMode)]string {
	return allMode
}

func DefaultConfig() *DrawConfig {
	return &DrawConfig{
		Width:    512,
		Height:   768,
		Num:      1,
		Strength: 0.70,
		Mode:     DefaultMode,
		Steps:    DefaultSteps,
		CfgScale: DefaultCfgScale,
		Uc:       defaultUC,
		Model:    gconfig.MODELS()[0].Name,
	}
}

func (cfg *DrawConfig) generateUC(translated bool) {
	if translated {
		cfg.Uc = utils.Translate(ReplaceString(cfg.Uc))
	} else {
		cfg.Uc = ReplaceString(cfg.Uc)
	}
	low := strings.ToLower(cfg.Uc)
	if !strings.Contains(low, `lowres`) {
		cfg.Uc = `lowres, ` + cfg.Uc
	}
	if !strings.Contains(low, `text`) {
		cfg.Uc = `text, ` + cfg.Uc
	}
}

func GenerateUC(uc string) string {
	uc = ReplaceString(uc)
	low := strings.ToLower(uc)
	if !strings.Contains(low, `lowres`) {
		uc = `lowres, ` + uc
	}
	if !strings.Contains(low, `text`) {
		uc = `text, ` + uc
	}
	return uc
}

func (cfg *DrawConfig) generateTag(translated bool) {
	if translated {
		cfg.Tag = utils.Translate(ReplaceString(cfg.Tag))
	} else {
		cfg.Tag = ReplaceString(cfg.Tag)
	}
	low := strings.ToLower(cfg.Tag)
	if !strings.Contains(low, "best quality") {
		cfg.Tag = "best quality, " + cfg.Tag
	}
	if !strings.Contains(low, "masterpiece") {
		cfg.Tag = "masterpiece, " + cfg.Tag
	}
}

func New(cfg *DrawConfig, initPhoto, ControlPhoto []byte, CorrectCfg bool) (*config, error) {
	if cfg.Tag == "" {
		return nil, errors.New("tag can not be empty")
	}
	if CorrectCfg {
		cfg.CorrectCfg(true, true, len(initPhoto) != 0, len(ControlPhoto) != 0, false, false, true)
	}
	dCfg := new(drawConfig)
	dCfg.Prompt = cfg.Tag
	dCfg.Seed = cfg.Seed
	dCfg.SamplerName = cfg.Mode
	dCfg.SamplerIndex = cfg.Mode
	dCfg.Width = cfg.Width
	dCfg.Height = cfg.Height
	dCfg.CfgScale = cfg.CfgScale
	dCfg.Steps = cfg.Steps
	dCfg.NegativePrompt = cfg.Uc
	dCfg.Num = 1
	dCfg.Count = cfg.Num
	c := &config{cfg: dCfg, rawCfg: *cfg}
	if len(initPhoto) != 0 {
		dCfg.ResizeMode = 2
		c.cfg.InitImages = []string{base64.StdEncoding.EncodeToString(initPhoto)}
		c.cfg.DenoisingStrength = c.rawCfg.Strength
	} else {
		dCfg.Width /= 2
		dCfg.Height /= 2
		dCfg.EnableHr = true
		dCfg.DenoisingStrength = 0.55
		dCfg.HrScale = 2
		dCfg.HrUpscaler = "R-ESRGAN 4x+ Anime6B"
		if dCfg.Steps < MaxHFSteps {
			dCfg.HrSecondPassSteps = dCfg.Steps
		} else {
			dCfg.HrSecondPassSteps = MaxHFSteps
		}
	}
	if len(ControlPhoto) != 0 {
		var max int
		if dCfg.Width > dCfg.Height {
			max = dCfg.Width
		} else {
			max = dCfg.Height
		}
		ctrl := ControlnetUnits{
			Lowvram:      false,
			InputImage:   base64.StdEncoding.EncodeToString(ControlPhoto),
			Module:       cfg.ControlPreprocess,
			Model:        gconfig.Name2Process(cfg.ControlProcess).File,
			ProcessorRes: max,
		}
		dCfg.AlwaysonScripts.Controlnet.Args = append(dCfg.AlwaysonScripts.Controlnet.Args, ctrl)
	}
	return c, nil
}

func NewSuperResolution(photo [][]byte, resize int) (*superResolutionCfg, error) {
	if resize > 4 || resize < 1 {
		return nil, errors.New("resize muse be in 1 and 4")
	}
	if photo == nil {
		return nil, errors.New("photo is nil")
	}
	cfg := &superResolutionCfg{ExtrasUpscaler2Visibility: 1, UpscalingResize: resize, Upscaler1: "R-ESRGAN 4x+ Anime6B"}
	for k, v := range photo {
		fileType, err := utils.GetType(v)
		if err != nil {
			continue
		}
		cfg.ImageList = append(cfg.ImageList, struct {
			Data string "json:\"data\""
			Name string "json:\"name\""
		}{Data: fmt.Sprint("data:", fileType, ";base64,", base64.StdEncoding.EncodeToString(v)), Name: fmt.Sprint(k)})
	}
	return cfg, nil
}

func NewSuperResolutionWithBase64(photo []string, multiplier int) (*superResolutionCfg, error) {
	if multiplier > 4 || multiplier < 1 {
		return nil, errors.New("multiplier muse be in 1 and 4")
	}
	if len(photo) == 0 {
		return nil, errors.New("photo is nil")
	}
	cfg := &superResolutionCfg{ExtrasUpscaler2Visibility: 1, UpscalingResize: multiplier, Upscaler1: "R-ESRGAN 4x+ Anime6B"}
	for k, v := range photo {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, err
		}
		fileType, err := utils.GetType(b)
		if err != nil {
			continue
		}
		cfg.ImageList = append(cfg.ImageList, struct {
			Data string "json:\"data\""
			Name string "json:\"name\""
		}{Data: fmt.Sprint("data:", fileType, ";base64,", v), Name: fmt.Sprint(k)})
	}
	return cfg, nil
}

func NewCtrlPhoto(photo [][]byte, Processor string, ResSize int) (*ctrlPhotoCfg, error) {
	if len(photo) == 0 {
		return nil, errors.New("photo is nil")
	}
	if ResSize < 0 {
		return nil, errors.New("size is less than zero")
	}
	cfg := &ctrlPhotoCfg{ControlnetModule: ParsePreProcess(Processor)}
	if ResSize == 0 {
		width, hight, err := utils.GetPhotoSize(photo[0])
		if err != nil {
			return nil, err
		}
		if width >= hight {
			cfg.ControlnetProcessorRes = width
		} else {
			cfg.ControlnetProcessorRes = hight
		}
	} else {
		cfg.ControlnetProcessorRes = ResSize
	}
	for _, v := range photo {
		cfg.ControlnetInputImages = append(cfg.ControlnetInputImages, base64.StdEncoding.EncodeToString(v))
	}
	return cfg, nil
}

func NewCtrlPhotoWithBash64(photo []string, Processor string, ResSize int) (*ctrlPhotoCfg, error) {
	if len(photo) == 0 {
		return nil, errors.New("photo is nil")
	}
	if ResSize < 0 {
		return nil, errors.New("size is less than zero")
	}
	cfg := &ctrlPhotoCfg{ControlnetModule: ParsePreProcess(Processor), ControlnetInputImages: photo}
	if ResSize == 0 {
		b, err := base64.StdEncoding.DecodeString(photo[0])
		if err != nil {
			return nil, err
		}
		width, hight, err := utils.GetPhotoSize(b)
		if err != nil {
			return nil, err
		}
		if width >= hight {
			cfg.ControlnetProcessorRes = width
		} else {
			cfg.ControlnetProcessorRes = hight
		}
	} else {
		cfg.ControlnetProcessorRes = ResSize
	}
	return cfg, nil
}

func NewInterrogate(photo []byte) (*interrogateCfg, error) {
	if photo == nil {
		return nil, errors.New("photo is nil")
	}
	fileType, err := utils.GetType(photo)
	if err != nil {
		return nil, err
	}
	cfg := &interrogateCfg{Image: fmt.Sprint("data:", fileType, ";base64,", base64.StdEncoding.EncodeToString(photo)), Model: "deepdanbooru"}
	return cfg, nil
}

func (cfg *config) GetCfg() DrawConfig {
	return cfg.rawCfg
}

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

func ParseCfgScalse(scale int) int {
	if scale <= 0 || scale > 30 {
		return DefaultCfgScale
	}
	return scale
}

func ParseSteps(steps int) int {
	if steps < 15 || steps > 50 {
		return DefaultSteps
	}
	return steps
}

func ParseNum(num int) int {
	if num <= 0 || num > MaxNum {
		return DefaultNum
	}
	return num
}

func ParsePreProcess(PreProcess string) string {
	cpp := gconfig.PreProcess()
	if PreProcess == "" {
		return cpp[0].Name
	}
	if _, ok := utils.In(cpp, func(c gconfig.ControlPreProcess) bool {
		return c.Name == PreProcess
	}); !ok {
		return cpp[0].Name
	}
	return PreProcess
}

func ParseProcess(Process string) string {
	cp := gconfig.Process()
	if Process == "" {
		return cp[0].Name
	}
	if _, ok := utils.In(cp, func(c gconfig.ControlProcess) bool {
		return c.Name == Process
	}); !ok {
		return cp[0].Name
	}
	return Process
}

func (cfg *DrawConfig) CorrectCfg(gTag, gUc, photo, ctrlPhoto, transTag, transUc, gSeed bool) {
	if !ctrlPhoto {
		cfg.ControlPreprocess = ""
		cfg.ControlProcess = ""
	} else {
		cfg.ControlPreprocess = ParsePreProcess(cfg.ControlPreprocess)
		cfg.ControlProcess = ParseProcess(cfg.ControlProcess)
	}
	if !photo {
		cfg.Strength = 0
	} else if cfg.Strength < 0 || cfg.Strength >= 1 {
		cfg.Strength = 0.70
	}
	cfg.CfgScale = ParseCfgScalse(cfg.CfgScale)
	if gSeed && cfg.Seed == 0 {
		cfg.Seed = uint32(rand.Intn(math.MaxUint32))
	}
	cfg.Model = gconfig.Name2Model(cfg.Model).Name
	cfg.Steps = ParseSteps(cfg.Steps)
	if gUc {
		cfg.generateUC(transUc)
	}
	if gTag {
		cfg.generateTag(transTag)
	}
	cfg.Num = ParseNum(cfg.Num)
	if cfg.Height < 64 || cfg.Width < 64 {
		cfg.Width = 512
		cfg.Height = 768
	} else if sum := cfg.Height * cfg.Width; sum > parseflag.ImgMaxSize {
		a := math.Pow(float64(sum)/float64(parseflag.ImgMaxSize), 0.5)
		cfg.Width = int(float64(cfg.Width) / a)
		cfg.Height = int(float64(cfg.Height) / a)
	}
	cfg.Width -= cfg.Width % 8
	cfg.Height -= cfg.Height % 8
	if _, ok := utils.InString(cfg.Mode, allMode[:]); !ok {
		cfg.Mode = allMode[0]
	}
}
