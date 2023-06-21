package router

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/handler"
	"github.com/zijiren233/stable-diffusion-webui-bot/i18n"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/user"

	"github.com/gin-gonic/gin"
	"github.com/zijiren233/go-colorlog"
	"golang.org/x/time/rate"
	"gopkg.in/yaml.v3"
)

type Resp struct {
	Time int64  `json:"time,omitempty"`
	Err  string `json:"err,omitempty"`
	Data any    `json:"data,omitempty"`
}

// I18N
// @Summary      Get i18n language
// @Description  Get i18n language by lang code, eg: /i18n/zh_cn
// @Tags         I18N
// @Produce      text/plain
// @Param        code   path      string  true  "Language code"
// @Success      200  {object}  I18NS
// @Router       /i18n/{code} [get]
func i18nYaml(ctx *gin.Context) {
	code := strings.Trim(ctx.Param("code"), "/")
	groups := gconfig.ExtraModelGroup()
	var alli18ns I18NS
	allLang := i18n.LoadAllExtraLang(code)
	for _, group := range groups {
		lora := gconfig.ExtraModelWithGroup(group)
		if name, ok := allLang[group]; ok {
			alli18ns = append(alli18ns, &I18N{Key: group, Value: name})
		} else {
			alli18ns = append(alli18ns, &I18N{Key: group, Value: group})
		}
		alli18ns = append(alli18ns, &I18N{Key: group, Value: group})
		i18ns := I18NS{}
		for _, l := range lora {
			if name, ok := allLang[l.Name]; ok {
				i18ns = append(i18ns, &I18N{Key: l.Name, Groups: strings.Join(l.Group, " | "), Value: name})
			} else {
				i18ns = append(i18ns, &I18N{Key: l.Name, Groups: strings.Join(l.Group, " | "), Value: l.Name})
			}
		}
		alli18ns = append(alli18ns, i18ns...)
	}
	alli18ns = uniqueI18NS(alli18ns)
	var msg bytes.Buffer
	for _, v := range alli18ns {
		if v.Groups != "" {
			msg.WriteString(fmt.Sprintf("%s: %s # %s\n", v.Key, v.Value, v.Groups))
		} else {
			msg.WriteString(fmt.Sprintf("%s: %s\n", v.Key, v.Value))
		}
	}
	_, err := ctx.Writer.Write(msg.Bytes())
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

// I18N-Json
// @Summary      Get i18n language
// @Description  Get i18n language by lang code, eg: /i18n-json/zh_cn
// @Tags         I18N
// @Produce      json
// @Param        code   path      string  true  "Language code"
// @Success      200  body  I18NS
// @Router       /i18n-json/{code} [get]
func i18nJson(ctx *gin.Context) {
	code := strings.Trim(ctx.Param("code"), "/")
	groups := gconfig.ExtraModelGroup()
	var alli18ns I18NS
	allLang := i18n.LoadAllExtraLang(code)
	for _, group := range groups {
		lora := gconfig.ExtraModelWithGroup(group)
		if name, ok := allLang[group]; ok {
			alli18ns = append(alli18ns, &I18N{Key: group, Value: name})
		} else {
			alli18ns = append(alli18ns, &I18N{Key: group, Value: group})
		}
		alli18ns = append(alli18ns, &I18N{Key: group, Value: group})
		i18ns := I18NS{}
		for _, l := range lora {
			if name, ok := allLang[l.Name]; ok {
				i18ns = append(i18ns, &I18N{Key: l.Name, Groups: strings.Join(l.Group, " | "), Value: name})
			} else {
				i18ns = append(i18ns, &I18N{Key: l.Name, Groups: strings.Join(l.Group, " | "), Value: l.Name})
			}
		}
		alli18ns = append(alli18ns, i18ns...)
	}
	ctx.JSON(http.StatusOK, uniqueI18NS(alli18ns))
}

// Get All Models
// @Summary      Get All Models
// @Description  Get All Models
// @Tags         Models
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /models [get]
func models(ctx *gin.Context) {
	data := []string{}
	m := gconfig.MODELS()
	for _, m2 := range m {
		data = append(data, m2.Name)
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
}

// Get Extra Model Group
// @Summary      Get Extra Model Group
// @Description  Get Extra Model Group
// @Tags         Models
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /extra-model-groups [get]
func extraModelGroups(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: gconfig.ExtraModelGroup(),
	})
}

// Get Extra Model With Group
// @Summary      Get Extra Model With Group
// @Description  Get Extra Model With Group
// @Tags         Models
// @Param        group   path      string  true  "Group Name"
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /extra-model-groups/{group} [get]
func extraModelWithGroups(ctx *gin.Context) {
	value := strings.Trim(ctx.Param("group"), "/")
	fmt.Printf("value: %v\n", value)
	data := []ExtraModel{}
	em := gconfig.ExtraModelWithGroup(value)
	for _, em2 := range em {
		data = append(data, ExtraModel{
			Name:         em2.Name,
			Type:         em2.Type,
			Preview:      em2.Preview,
			TriggerWords: em2.TriggerWords,
			Group:        em2.Group,
		})
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
}

type ExtraModel struct {
	Name         string
	Preview      string
	Type         string
	TriggerWords []string
	Group        []string
}

// Get All Extra Model
// @Summary      Get All Extra Model
// @Description  Get All Extra Model
// @Tags         Models
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /extra-models [get]
func extraModel(ctx *gin.Context) {
	data := []ExtraModel{}
	em := gconfig.ALLExtraModel()
	for _, em2 := range em {
		data = append(data, ExtraModel{
			Name:         em2.Name,
			Type:         em2.Type,
			Preview:      em2.Preview,
			TriggerWords: em2.TriggerWords,
			Group:        em2.Group,
		})
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
}

// Get All Control PreProcess
// @Summary      Get All Control PreProcess
// @Description  Get All Control PreProcess
// @Tags         Models
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /control-preprocess [get]
func controlPreprocess(ctx *gin.Context) {
	data := []string{}
	cpp := gconfig.PreProcess()
	for _, cpp2 := range cpp {
		data = append(data, cpp2.Name)
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
}

// Get All Control Process
// @Summary      Get All Control Process
// @Description  Get All Control Process
// @Tags         Models
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /control-process [get]
func controlProcess(ctx *gin.Context) {
	data := []string{}
	cp := gconfig.Process()
	for _, cp2 := range cp {
		data = append(data, cp2.Name)
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
}

// Get All Models
// @Summary      Get All Models
// @Description  Get All Models
// @Tags         Models
// @Produce      json
// @Success      200  {object}  Resp
// @Router       /modes [get]
func allModels(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: api.AllMode(),
	})
}

// Test Draw Config
// @Summary      Test Draw Config
// @Description  Test Draw Config
// @Tags         Image Handler
// @Accept       json
// @Produce      json
// @Param        config  body  Any2Img  true  "Draw Config"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /test-draw-config [post]
func testDrawConfig(ctx *gin.Context) {
	cfg := new(Any2Img)
	err := json.NewDecoder(ctx.Request.Body).Decode(cfg)
	if err != nil && err != io.EOF {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	photo, err := base64.StdEncoding.DecodeString(cfg.PrePhoto)
	if err != nil && err != io.EOF {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	ctrlphoto, err := base64.StdEncoding.DecodeString(cfg.ControlPhoto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	cfg.CorrectCfg(true, true, len(photo) != 0, len(ctrlphoto) != 0, false, false, true)
	var data = struct {
		Cfg           api.DrawConfig `json:"config"`
		Pre_photo     bool           `json:"pre_photo"`
		Control_photo bool           `json:"control_photo"`
	}{
		Cfg:           cfg.DrawConfig,
		Pre_photo:     len(photo) != 0,
		Control_photo: len(ctrlphoto) != 0,
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
}

// Draw
// @Summary      Draw Img
// @Description  Any to Img
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Accept       json
// @Produce      json
// @Param        config  body  Any2Img  true  "Draw Config"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /draw [post]
func (r *Router) drawPost(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := UserInfo.AddTask(user.T_Draw)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	ct, cancel := context.WithCancel(context.Background())
	task.Set("cancel", cancel)
	cfg := new(Any2Img)
	err = json.NewDecoder(ctx.Request.Body).Decode(cfg)
	if err != nil && err != io.EOF {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	photo, err := base64.StdEncoding.DecodeString(cfg.PrePhoto)
	if err != nil && err != io.EOF {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	ctrlphoto, err := base64.StdEncoding.DecodeString(cfg.ControlPhoto)
	if err != nil {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	c, err := api.New(&cfg.DrawConfig, photo, ctrlphoto, true)
	if err != nil {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	var data = struct {
		Cfg           api.DrawConfig `json:"config"`
		Pre_photo     bool           `json:"pre_photo"`
		Control_photo bool           `json:"control_photo"`
	}{
		Cfg:           c.GetCfg(),
		Pre_photo:     len(photo) != 0,
		Control_photo: len(ctrlphoto) != 0,
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: data,
	})
	task.Set("result", c.Draw(ct, UserInfo.Permissions() != user.T_Subscribe))
	task.Set("status", c.Status)
}

// Draw
// @Summary      Draw Img
// @Description  Any to Img
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Produce      json
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /draw [get]
func (r *Router) drawGet(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := user.GetTask(UserInfo.UserInfo.UserID, user.T_Draw)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	t := time.NewTimer(time.Second * 3)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
		ctx.JSON(http.StatusOK, Resp{
			Time: time.Now().Unix(),
			Err:  "re long polling",
			Data: task.Value("status").(func() api.Status)().Progress,
		})
	case res := <-task.Value("result").(<-chan *api.Resoult):
		task.Down()
		if res.Err != nil {
			ctx.JSON(http.StatusInternalServerError, Resp{
				Time: time.Now().Unix(),
				Err:  "Internal Server Error",
			})
		} else {
			if UserInfo.Permissions() == user.T_Guest {
				UserInfo.UseFree(1)
			}
			msg := []string{}
			for _, v := range res.Resoult {
				msg = append(msg, base64.StdEncoding.EncodeToString(v))
			}
			ctx.JSON(http.StatusOK, Resp{
				Time: time.Now().Unix(),
				Data: msg,
			})
		}
	}
}

// Interrupt
// @Summary      Interrupt Draw Task
// @Description  Interrupt Draw Task
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Produce      json
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /interrupt [get]
func (r *Router) interrupt(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := user.GetTask(UserInfo.UserInfo.UserID, user.T_Draw)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	task.Value("cancel").(context.CancelFunc)()
	task.Down()
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: "OK",
	})
}

type CtrlPhotoCfg struct {
	ControlPhoto []string `json:"control_photo"`
	PreProcessor string   `json:"control_preprocess"`
	ResSize      int      `json:"res_size"`
}

// Detect
// @Summary      Detect to Ctrl Photo
// @Description  Detect to Ctrl Photo
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Accept       json
// @Produce      json
// @Param        config  body  CtrlPhotoCfg  true  "Ctrl Photo Config"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /detect-ctrl-photo [post]
func (r *Router) detectCtrlPhotoPost(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := UserInfo.AddTask(user.T_CtrlPhoto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	cfg := new(CtrlPhotoCfg)
	err = json.NewDecoder(ctx.Request.Body).Decode(cfg)
	if err != nil && err != io.EOF {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	if len(cfg.ControlPhoto) > 3 {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  "photo max num is 3",
		})
		ctx.Abort()
		return
	}
	c, err := api.NewCtrlPhotoWithBash64(cfg.ControlPhoto, cfg.PreProcessor, cfg.ResSize)
	if err != nil {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: "pls get /api/detect-ctrl-photo",
	})
	task.Set("result", c.CtrlPhoto(context.Background(), UserInfo.Permissions() != user.T_Subscribe))
}

// Detect
// @Summary      Detect to Ctrl Photo
// @Description  Detect to Ctrl Photo
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Produce      json
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /detect-ctrl-photo [get]
func (r *Router) detectCtrlPhotoGet(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := user.GetTask(UserInfo.UserInfo.UserID, user.T_CtrlPhoto)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	t := time.NewTimer(time.Second * 3)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
		ctx.JSON(http.StatusOK, Resp{
			Time: time.Now().Unix(),
			Err:  "re long polling",
		})
	case res := <-task.Value("result").(<-chan *api.Resoult):
		task.Down()
		if res.Err != nil {
			ctx.JSON(http.StatusInternalServerError, Resp{
				Time: time.Now().Unix(),
				Err:  "Internal Server Error",
			})
		} else {
			if UserInfo.Permissions() == user.T_Guest {
				UserInfo.UseFree(1)
			}
			msg := []string{}
			for _, v := range res.Resoult {
				msg = append(msg, base64.StdEncoding.EncodeToString(v))
			}
			ctx.JSON(http.StatusOK, Resp{
				Time: time.Now().Unix(),
				Data: msg,
			})
		}
	}
}

type SuperResolutionCfg struct {
	Photo      []string
	Multiplier int
}

// SuperResolution
// @Summary      SuperResolution
// @Description  SuperResolution
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Accept       json
// @Produce      json
// @Param        config  body  SuperResolutionCfg  true  "SuperResolution Config"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /super-resolution [post]
func (r *Router) superResolutionPost(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := UserInfo.AddTask(user.T_SuperResolution)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	ct, cancel := context.WithCancel(context.Background())
	task.Set("cancel", cancel)
	cfg := new(SuperResolutionCfg)
	err = json.NewDecoder(ctx.Request.Body).Decode(cfg)
	if err != nil && err != io.EOF {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	if len(cfg.Photo) > 3 {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  "photo num mast less than 3",
		})
		ctx.Abort()
		return
	}
	for i, v := range cfg.Photo {
		if len(v) > 1024*1024*20 {
			task.Down()
			ctx.JSON(http.StatusInternalServerError, Resp{
				Time: time.Now().Unix(),
				Err:  fmt.Sprintf("The %d picture is too large", i),
			})
			ctx.Abort()
			return
		}
	}
	c, err := api.NewSuperResolutionWithBase64(cfg.Photo, cfg.Multiplier)
	if err != nil {
		task.Down()
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	ctx.JSON(http.StatusOK, Resp{
		Time: time.Now().Unix(),
		Data: "pls get /api/superResolution",
	})
	task.Set("result", c.SuperResolution(ct, UserInfo.Permissions() != user.T_Subscribe))
}

// SuperResolution
// @Summary      SuperResolution
// @Description  SuperResolution
// @Tags         Image Handler
// @securityDefinitions.basic BasicAuth
// @Produce      json
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /super-resolution [get]
func (r *Router) superResolutionGet(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	if UserInfo.Permissions() == user.T_Prohibit {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  UserInfo.ProhibitString(r.handler.Bot()),
		})
		ctx.Abort()
		return
	}
	task, err := user.GetTask(UserInfo.UserInfo.UserID, user.T_SuperResolution)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Resp{
			Time: time.Now().Unix(),
			Err:  err.Error(),
		})
		ctx.Abort()
		return
	}
	t := time.NewTimer(time.Second * 3)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
		ctx.JSON(http.StatusOK, Resp{
			Time: time.Now().Unix(),
			Err:  "re long polling",
		})
	case res := <-task.Value("result").(<-chan *api.Resoult):
		task.Down()
		if res.Err != nil {
			ctx.JSON(http.StatusInternalServerError, Resp{
				Time: time.Now().Unix(),
				Err:  "Internal Server Error",
			})
		} else {
			if UserInfo.Permissions() == user.T_Guest {
				UserInfo.UseFree(1)
			}
			msg := []string{}
			for _, v := range res.Resoult {
				msg = append(msg, base64.StdEncoding.EncodeToString(v))
			}
			ctx.JSON(http.StatusOK, Resp{
				Time: time.Now().Unix(),
				Data: msg,
			})
		}
	}
}

type ImagesData struct {
	MaxCount int64    `json:"maxcount,omitempty"`
	Results  []Result `json:"resoult"`
}

// searchImages
// @Summary      search Images
// @Description  search Images
// @Tags         Images
// @Param keywords query string true "keywords"
// @Param maxcount query boolean  false "maxcount"
// @Param order_type query string false "order type: latest | random"
// @Param page query string false "page"
// @Param time query string false "time"
// @Param cfg_type query string false "cfg type: json | yaml"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /search-images [get]
func searchImages(ctx *gin.Context) {
	dataBucket.Wait(ctx)
	keyWord := ctx.Query("keywords")
	var photo []db.PhotoInfo
	var maxCount int64 = 0
	switch ctx.Query("order_type") {
	case "latest":
		page, err := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
		if err != nil {
			ctx.Abort()
			return
		}
		if page == 0 {
			page = 1
		}
		i, err := strconv.ParseInt(ctx.DefaultQuery("time", fmt.Sprint(time.Now().Unix())), 10, 64)
		if err != nil {
			ctx.Abort()
			return
		}
		parseT := time.UnixMilli(i)
		now := time.Now()
		var Time time.Time
		if parseT.After(now.Add(time.Second * 3)) {
			Time = time.Now()
		} else {
			Time = parseT
		}
		kw := strings.Split(strings.ReplaceAll(keyWord, "，", ","), ",")
		photo, err = db.FindImg(db.FindConfig{
			Deadline: Time,
			Order:    "id desc",
			Limit:    20,
			Offset:   20 * (page - 1),
			Keywords: kw,
		})
		if err != nil {
			ctx.Abort()
			return
		}
		if ctx.Query("maxcount") == "true" {
			maxCount = db.GetMaxCount(db.MaxCountCfg{Deadline: Time, User_id: nil, Keywords: kw})
		}
	default:
		var err error
		photo, err = db.FindImg(db.FindConfig{
			Deadline: time.Now(),
			Order:    "RANDOM()",
			Limit:    20,
			Offset:   0,
			Keywords: strings.Split(strings.ReplaceAll(keyWord, "，", ","), ","),
		})
		if err != nil {
			ctx.Abort()
			return
		}
	}
	resp := []Result{}
	for _, v := range photo {
		switch ctx.Query("cfg_type") {
		case "json":
			resp = append(resp, Result{Id: v.FileID, Image: fmt.Sprintf("%s://%s/api/images/%s.png", parseflag.ApiScheme, parseflag.ApiHost, v.FileID), Cfg: handler.Config{DrawConfig: v.Config, PrePhotoID: v.PrePhotoID, ControlPhotoID: v.ControlPhotoID}, Width: 230, Height: int(float64(v.Config.Height) / float64(v.Config.Width) * 230)})
		default:
			cfg, err := yaml.Marshal(handler.Config{DrawConfig: v.Config, PrePhotoID: v.PrePhotoID, ControlPhotoID: v.ControlPhotoID})
			if err != nil {
				continue
			}
			resp = append(resp, Result{Id: v.FileID, Image: fmt.Sprintf("%s://%s/api/images/%s.png", parseflag.ApiScheme, parseflag.ApiHost, v.FileID), Cfg: string(cfg), Width: 230, Height: int(float64(v.Config.Height) / float64(v.Config.Width) * 230)})
		}
	}
	ctx.JSON(http.StatusOK, Resp{Time: time.Now().Unix(), Data: ImagesData{
		MaxCount: maxCount,
		Results:  resp,
	}})
}

// searchUserImages
// @Summary      search User Images
// @Description  search User Images
// @Tags         Images
// @Param keywords query string true "keywords"
// @Param maxcount query boolean  false "maxcount"
// @Param order_type query string false "order type: latest | random"
// @Param page query string false "page"
// @Param time query string false "time"
// @Param cfg_type query string false "cfg type: json | yaml"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /search-user-images [get]
func searchUserImages(ctx *gin.Context) {
	u, exists := ctx.Get("user")
	if !exists {
		authErr(ctx)
		return
	}
	UserInfo := u.(*user.UserInfo)
	dataBucket.Wait(ctx)
	keyWord := ctx.Query("keywords")
	var photo []db.PhotoInfo
	var maxCount int64 = 0
	switch ctx.Query("order_type") {
	case "latest":
		page, err := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
		if err != nil {
			ctx.Abort()
			return
		}
		if page == 0 {
			page = 1
		}
		i, err := strconv.ParseInt(ctx.DefaultQuery("time", fmt.Sprint(time.Now().Unix())), 10, 64)
		if err != nil {
			ctx.Abort()
			return
		}
		parseT := time.UnixMilli(i)
		now := time.Now()
		var Time time.Time
		if parseT.After(now.Add(time.Second * 3)) {
			Time = time.Now()
		} else {
			Time = parseT
		}
		kw := strings.Split(strings.ReplaceAll(keyWord, "，", ","), ",")
		photo, err = db.FindImg(db.FindConfig{
			Deadline: Time,
			Order:    "id desc",
			User_id:  UserInfo.UserInfo.UserID,
			Limit:    20,
			Offset:   20 * (page - 1),
			Keywords: kw,
		})
		if err != nil {
			ctx.Abort()
			return
		}
		if ctx.Query("maxcount") == "true" {
			maxCount = db.GetMaxCount(db.MaxCountCfg{Deadline: Time, User_id: UserInfo.UserInfo.UserID, Keywords: kw})
		}
	default:
		var err error
		photo, err = db.FindImg(db.FindConfig{
			Deadline: time.Now(),
			Order:    "RANDOM()",
			User_id:  UserInfo.UserInfo.UserID,
			Limit:    20,
			Offset:   0,
			Keywords: strings.Split(strings.ReplaceAll(keyWord, "，", ","), ","),
		})
		if err != nil {
			ctx.Abort()
			return
		}
	}
	resp := []Result{}
	for _, v := range photo {
		switch ctx.Query("cfg_type") {
		case "json":
			resp = append(resp, Result{Id: v.FileID, Image: fmt.Sprintf("%s://%s/api/images/%s.png", parseflag.ApiScheme, parseflag.ApiHost, v.FileID), Cfg: handler.Config{DrawConfig: v.Config, PrePhotoID: v.PrePhotoID, ControlPhotoID: v.ControlPhotoID}, Width: 230, Height: int(float64(v.Config.Height) / float64(v.Config.Width) * 230)})
		default:
			cfg, err := yaml.Marshal(handler.Config{DrawConfig: v.Config, PrePhotoID: v.PrePhotoID, ControlPhotoID: v.ControlPhotoID})
			if err != nil {
				continue
			}
			resp = append(resp, Result{Id: v.FileID, Image: fmt.Sprintf("%s://%s/api/images/%s.png", parseflag.ApiScheme, parseflag.ApiHost, v.FileID), Cfg: string(cfg), Width: 230, Height: int(float64(v.Config.Height) / float64(v.Config.Width) * 230)})
		}
	}
	ctx.JSON(http.StatusOK, Resp{Time: time.Now().Unix(), Data: ImagesData{
		MaxCount: maxCount,
		Results:  resp,
	}})
}

var imgL = rate.NewLimiter(30, 1)

// Images
// @Summary      Images
// @Description  Images
// @Tags         Images
// @Param filename path string true "filename"
// @Success      200  {object}  Resp
// @Failure      500  {object}  Resp
// @Router       /images/{filename} [get]
func (r *Router) Images(ctx *gin.Context) {
	imgL.Wait(ctx)
	filename := strings.Trim(ctx.Param("filename"), "/")
	id := strings.TrimSuffix(filename, `.png`)
	if strings.HasSuffix(filename, `.png`) && len(id) == 32 {
		data, err := r.handler.Cache().Get(id)
		if err != nil {
			colorlog.Errorf("website err: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get file: object not found", "data": nil})
			return
		}
		ctx.Writer.Write(data)
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get file: object not found", "data": nil})
	}
}
