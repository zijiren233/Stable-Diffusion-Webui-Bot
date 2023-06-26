package parseflag

import (
	"flag"
	"fmt"
)

var (
	TgToken                         string
	Dev                             bool
	WebhookHost, ApiHost, ApiScheme string
	Port                            int
	OwnerID                         int64
	MaxNum							int
	DSN                             string
	ImgMaxSize                      int
	MaxFree                         int
	ImageSavePath                   string
	ImageCacheNum                   int
	EnableInvite                    bool
	I18nExtraPath                   string
	Listen                          string
)

func init() {
	flag.BoolVar(&Dev, "dev", false, "development mode")
	flag.IntVar(&ImgMaxSize, "img-max", 1638400, "image maximum resolution")
	flag.Int64Var(&OwnerID, "owner", 2143676086, "owner telegram id")
	flag.IntVar(&Port, "port", 8082, "port")
	flag.IntVar(&MaxFree, "max-free", 0, "free user max free time")
	flag.IntVar(&MaxNum, "max-num", 6 , "max number of images")
	flag.IntVar(&ImageCacheNum, "img-cache-num", 100, "image cache to mem max num")
	flag.BoolVar(&EnableInvite, "invite", false, "Enable Invite")
	flag.StringVar(&WebhookHost, "webhook-host", "", "enable telegram bot webhook: webhook.doamin.com")
	flag.StringVar(&Listen, "listen", "127.0.0.1", "listening address: 127.0.0.1 | 0.0.0.0")
	flag.StringVar(&ApiScheme, "api-scheme", "http", "api scheme: http | https")
	flag.StringVar(&ApiHost, "api-host", fmt.Sprintf("%s:%d", Listen, Port), "set api url")
	flag.StringVar(&ImageSavePath, "img-save-path", `./local-cache`, "Image Save Path")
	flag.StringVar(&DSN, "dsn", `./stable-diffusion-webui-bot.db`, "database, postgres|sqlite")
	flag.StringVar(&TgToken, "tg-token", ``, "telegram bot token")
	flag.StringVar(&I18nExtraPath, "i18n-extra-path", `./i18n-extra`, "i18n extra translated")
}
