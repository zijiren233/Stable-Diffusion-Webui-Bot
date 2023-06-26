package start

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zijiren233/stable-diffusion-webui-bot/cache"
	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/handler"
	"github.com/zijiren233/stable-diffusion-webui-bot/router"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"

	"github.com/panjf2000/ants/v2"
	"github.com/zijiren233/go-colorlog"
)

func init() {
	rand.Seed(time.Now().UnixMilli())
	gin.SetMode(gin.ReleaseMode)

	colorlog.SetLogLevle(colorlog.L_Debug)
}

func Main() {
	flag.Parse()
	d := db.New(parseflag.DSN)
	err := gconfig.Load(gconfig.ConfigPath)
	if err != nil {
		panic(err)
	}
	// go gconfig.Wathc()
	// api.Load(gconfig.API())
	// go func() {
	// 	w := gconfig.NewWatchConfig(context.Background())
	// 	go func() {
	// 		time.Sleep(5 * time.Second)
	// 		w.Close()
	// 	}()
	// 	c := w.Ch()
	// 	for range c {
	// 		api.Load(gconfig.API())
	// 	}
	// }()
	defer ants.Release()
	c, err := cache.NewCache(cache.WithSavePath(parseflag.ImageSavePath), cache.WithCacheNum(parseflag.ImageCacheNum))
	if err != nil {
		panic(fmt.Errorf("new Cache Error: %v", err))
	}
	hConfigs := []handler.ConfigFunc{handler.WithCache(c), handler.WithImgMaxSize(parseflag.ImgMaxSize), handler.WithMaxNum(parseflag.MaxNum)}
	if parseflag.WebhookHost != "" {
		hConfigs = append(hConfigs, handler.WithWebhook(parseflag.WebhookHost))
	}
	a, err := api.New(gconfig.API(), gconfig.MODELS())
	if err != nil {
		panic(err)
	}
	h, err := handler.New(parseflag.TgToken, a, d, hConfigs...)
	if err != nil {
		panic(err)
	}
	rConfigs := []router.ConfigFunc{}
	rConfigs = append(rConfigs, router.WithDocs(), router.WithAPI(h))
	if parseflag.WebhookHost != "" {
		rConfigs = append(rConfigs, router.WithWebhook(h.WebhookUriPath(), h.WebhookHandler()))
	}
	r, err := router.New(gin.New(), rConfigs...)
	if err != nil {
		panic(err)
	}
	go r.Eng().Run(fmt.Sprintf("%s:%d", parseflag.Listen, parseflag.Port))
	go h.Run(context.Background())
	colorlog.Infof("Service started successfully!\n%s://%s:%d", parseflag.ApiScheme, parseflag.Listen, parseflag.Port)
	select {}
}
