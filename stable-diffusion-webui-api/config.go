package api

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/zijiren233/go-colorlog"
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
	cfg         *Config
	api         struct {
		api    *apiUrl
		status Status
	}
	a *API
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
	a *API
}

type ctrlPhotoCfg struct {
	resoultChan            chan<- *Resoult `json:"-"`
	ctx                    context.Context
	ControlnetModule       string   `json:"controlnet_module"`
	ControlnetInputImages  []string `json:"controlnet_input_images"`
	ControlnetProcessorRes int      `json:"controlnet_processor_res"`
	ControlnetThresholdA   int      `json:"controlnet_threshold_a"`
	ControlnetThresholdB   int      `json:"controlnet_threshold_b"`
	a                      *API
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
	a           *API
}

type Config struct {
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
	Model             string   `json:"-"`
	Vae               string   `json:"-"`
	ClipSkip          int      `json:"-"`
	CfgScale          int      `json:"cfg_scale"`
	Width             int      `json:"width"`
	Height            int      `json:"height"`
	NegativePrompt    string   `json:"negative_prompt"`
	SamplerName       string   `json:"sampler_name"`
	SamplerIndex      string   `json:"sampler_index"`
	AlwaysonScripts   struct {
		Controlnet *struct {
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

type API struct {
	drawPool      *ants.Pool
	loadBalance   *api
	waitGloup     chan func()
	freeWaitGloup chan func()
	getApiL       *sync.Mutex
	models        []string
}

func New(apis []gconfig.Api, models []string) (*API, error) {
	a := &API{models: models, getApiL: &sync.Mutex{}, waitGloup: make(chan func(), math.MaxInt16), freeWaitGloup: make(chan func(), math.MaxInt16)}
	drawPool, err := ants.NewPool(1, ants.WithOptions(ants.Options{PreAlloc: false, Logger: nil, DisablePurge: false, Nonblocking: false, PanicHandler: func(i interface{}) {
		colorlog.Fatal(utils.PrintStackTrace(i))
	}}))
	if err != nil {
		return nil, err
	}
	a.loadBalance = &api{apiList: &[]*apiUrl{},
		lock: &sync.RWMutex{}}
	a.drawPool = drawPool
	a.Load(apis)
	return a, nil
}

func (api *API) New(cfg *Config, initPhoto, ControlPhoto []byte) (*config, error) {
	return &config{cfg: cfg, a: api}, nil
}

func (api *API) NewSuperResolution(photo [][]byte, resize int) (*superResolutionCfg, error) {
	if resize > 4 || resize < 1 {
		return nil, errors.New("resize muse be in 1 and 4")
	}
	if photo == nil {
		return nil, errors.New("photo is nil")
	}
	cfg := &superResolutionCfg{ExtrasUpscaler2Visibility: 1, UpscalingResize: resize, Upscaler1: "R-ESRGAN 4x+ Anime6B", a: api}
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

func (api *API) NewSuperResolutionWithBase64(photo []string, multiplier int) (*superResolutionCfg, error) {
	if multiplier > 4 || multiplier < 1 {
		return nil, errors.New("multiplier muse be in 1 and 4")
	}
	if len(photo) == 0 {
		return nil, errors.New("photo is nil")
	}
	cfg := &superResolutionCfg{ExtrasUpscaler2Visibility: 1, UpscalingResize: multiplier, Upscaler1: "R-ESRGAN 4x+ Anime6B", a: api}
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

func (api *API) NewCtrlPhoto(photo [][]byte, Processor string, ResSize int) (*ctrlPhotoCfg, error) {
	if len(photo) == 0 {
		return nil, errors.New("photo is nil")
	}
	if ResSize < 0 {
		return nil, errors.New("size is less than zero")
	}
	cfg := &ctrlPhotoCfg{ControlnetModule: Processor, a: api}
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

func (api *API) NewCtrlPhotoWithBash64(photo []string, Processor string, ResSize int) (*ctrlPhotoCfg, error) {
	if len(photo) == 0 {
		return nil, errors.New("photo is nil")
	}
	if ResSize < 0 {
		return nil, errors.New("size is less than zero")
	}
	cfg := &ctrlPhotoCfg{ControlnetModule: Processor, ControlnetInputImages: photo, a: api}
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

func (api *API) NewInterrogate(photo []byte) (*interrogateCfg, error) {
	if photo == nil {
		return nil, errors.New("photo is nil")
	}
	fileType, err := utils.GetType(photo)
	if err != nil {
		return nil, err
	}
	cfg := &interrogateCfg{a: api, Image: fmt.Sprint("data:", fileType, ";base64,", base64.StdEncoding.EncodeToString(photo)), Model: "deepdanbooru"}
	return cfg, nil
}
