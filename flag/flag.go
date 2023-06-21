package parseflag

import (
	"flag"
	"fmt"
)

var (
	TgToken              string
	Dev                  bool
	WebhookHost, APIHost string
	Port                 int
	OwnerID              int64
	DSN                  string
	ImgMaxSize           int
	MaxFree              int
	ImageSavePath        string
	ImageCacheNum        int
	// HOST                string
	I18nExtraPath string
	Listen        string
)

func init() {
	flag.BoolVar(&Dev, "dev", false, "development mode")
	flag.IntVar(&ImgMaxSize, "imgMax", 1638400, "image maximum resolution")
	flag.Int64Var(&OwnerID, "owner", 2143676086, "owner telegram id")
	flag.IntVar(&Port, "p", 8082, "port")
	flag.IntVar(&MaxFree, "mf", 0, "free user max free time")
	flag.IntVar(&ImageCacheNum, "ics", 100, "image cache to mem max num")
	// flag.StringVar(&HOST, "host", ``, "webhook or api host")
	flag.StringVar(&WebhookHost, "webhook", "", "enable telegram bot webhook: webhook.doamin.com")
	flag.StringVar(&Listen, "listen", "127.0.0.1", "listening address: 127.0.0.1 | 0.0.0.0")
	flag.StringVar(&APIHost, "api", fmt.Sprintf("http://%s:%d", Listen, Port), "set api url")
	flag.StringVar(&ImageSavePath, "isp", `./local-cache`, "Image Save Path")
	flag.StringVar(&DSN, "dsn", `./stable-diffusion-webui-bot.db`, "database, postgres|sqlite")
	flag.StringVar(&TgToken, "t", ``, "telegram bot token")
	flag.StringVar(&I18nExtraPath, "i18np", `./i18n-extra`, "i18n extra translated")
}
