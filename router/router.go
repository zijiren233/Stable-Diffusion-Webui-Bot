package router

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/zijiren233/go-colorlog"
	_ "github.com/zijiren233/stable-diffusion-webui-bot/docs"
	"github.com/zijiren233/stable-diffusion-webui-bot/handler"
	"github.com/zijiren233/stable-diffusion-webui-bot/web"
	"golang.org/x/time/rate"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var logFormat = gin.LoggerWithFormatter(log)

func log(params gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if params.IsOutputColor() {
		statusColor = params.StatusCodeColor()
		methodColor = params.MethodColor()
		resetColor = params.ResetColor()
	}
	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	return colorlog.Format(params.TimeStamp, colorlog.L_Debug, fmt.Sprintf("|%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		statusColor, params.StatusCode, resetColor,
		params.Latency,
		params.ClientIP,
		methodColor, params.Method, resetColor,
		params.Path,
		params.ErrorMessage,
	))
}

func (r *Router) regDocs() *gin.Engine {
	rg := r.eng.Group("/docs")
	rg.Use(logFormat)
	rg.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r.eng
}

var dataBucket = rate.NewLimiter(30, 1)

type Result struct {
	Id     string `json:"id"`
	Image  string `json:"image"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Cfg    any    `json:"cfg"`
}

type Router struct {
	eng            *gin.Engine
	handler        *handler.Handler
	api            bool
	docs           bool
	webhookHandler func(w http.ResponseWriter, r *http.Request)
	webhookUriPath string
}

type ConfigFunc func(r *Router)

func WithAPI(handler *handler.Handler) ConfigFunc {
	return func(r *Router) { r.api = true; r.handler = handler }
}

func WithDocs() ConfigFunc {
	return func(r *Router) { r.docs = true }
}

func WithWebhook(webhookUriPath string, webhookHandler func(w http.ResponseWriter, r *http.Request)) ConfigFunc {
	return func(r *Router) { r.webhookHandler = webhookHandler; r.webhookUriPath = webhookUriPath }
}

func New(eng *gin.Engine, config ...ConfigFunc) (*Router, error) {
	r := &Router{eng: eng}
	for _, cf := range config {
		cf(r)
	}
	if r.api {
		r.apis()
	}
	if r.docs {
		r.regDocs()
	}
	if r.webhookHandler != nil {
		eng.POST(r.webhookUriPath, func(ctx *gin.Context) {
			r.webhookHandler(ctx.Writer, ctx.Request)
		})
	}
	fStatic, err := fs.Sub(web.Static, "static")
	if err != nil {
		return nil, err
	}
	assets := r.eng.Group("/assets")
	{
		assets.Use(logFormat)
		fAssets, err := fs.Sub(fStatic, "assets")
		if err != nil {
			return nil, err
		}
		assets.StaticFS("/", http.FS(fAssets))
	}
	r.eng.Use(logFormat)
	r.eng.NoRoute(func(ctx *gin.Context) {
		ctx.FileFromFS("/", http.FS(fStatic))
	})
	return r, nil
}

func (r *Router) Eng() *gin.Engine {
	return r.eng
}
