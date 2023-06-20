package router

import (
	"net/http"
	"strconv"

	"github.com/zijiren233/stable-diffusion-webui-bot/user"
	tgbotapi "github.com/zijiren233/tg-bot-api/v6"

	"github.com/gin-gonic/gin"
)

func auth(bot *tgbotapi.BotAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userid, pwd, ok := ctx.Request.BasicAuth()
		if !ok {
			authErr(ctx)
			return
		}
		id, err := strconv.ParseInt(userid, 10, 64)
		if err != nil {
			authErr(ctx)
			return
		}
		ui, err := user.LoadUser(bot, id)
		if err != nil || ui.Passwd() != pwd {
			authErr(ctx)
			return
		}
		ctx.Set("user", ui)
	}
}

func authErr(ctx *gin.Context) {
	ctx.Writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	ctx.Writer.WriteHeader(http.StatusUnauthorized)
	ctx.Abort()
}
