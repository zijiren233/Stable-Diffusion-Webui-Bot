package router

import (
	"fmt"
	"time"

	_ "github.com/zijiren233/stable-diffusion-webui-bot/docs"
	tgbotapi "github.com/zijiren233/tg-bot-api/v6"
	"golang.org/x/time/rate"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/zijiren233/go-colorlog"
)

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

func collectRoute(eng *gin.Engine) *gin.Engine {
	eng.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return eng
}

var dataBucket = rate.NewLimiter(30, 1)
var eng *gin.Engine
var bot *tgbotapi.BotAPI

type Result struct {
	Id     string `json:"id"`
	Image  string `json:"image"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Cfg    any    `json:"cfg"`
}

func init() {
	gin.SetMode(gin.ReleaseMode)
	eng = gin.New()
}

func SetBot(b *tgbotapi.BotAPI) {
	bot = b
}

func Router() {
	apis(eng)
	collectRoute(eng)
}

func Eng() *gin.Engine {
	return eng
}
