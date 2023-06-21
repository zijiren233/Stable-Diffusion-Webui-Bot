package start

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/handler"
	"github.com/zijiren233/stable-diffusion-webui-bot/router"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/zijiren233/tg-bot-api/v6"

	"github.com/panjf2000/ants/v2"
	"github.com/zijiren233/go-colorlog"
)

var (
	Bot      *tgbotapi.BotAPI
	mainPllo *ants.Pool
)

func init() {
	rand.Seed(time.Now().UnixMilli())

	if err := os.RemoveAll(path.Join(os.TempDir(), "tmp-stable-diffusion-webui-bot")); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(path.Join(os.TempDir(), "tmp-stable-diffusion-webui-bot"), os.ModePerm); err != nil {
		panic(err)
	}
	colorlog.SetLogLevle(colorlog.L_Debug)
}

func SetCommand() {
	var cmds tgbotapi.SetMyCommandsConfig

	Bot.Send(tgbotapi.NewDeleteMyCommands())

	bcs := tgbotapi.NewBotCommandScopeAllPrivateChats()

	cmds = tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{Command: "invite", Description: "Get invitation link"}, tgbotapi.BotCommand{Command: "web", Description: "Use Web Site version"}, tgbotapi.BotCommand{Command: "api", Description: "Get API documentation"}, tgbotapi.BotCommand{Command: "setdefault", Description: "Set default parameters"}, tgbotapi.BotCommand{Command: "subscribe", Description: "View subscription information"}, tgbotapi.BotCommand{Command: "share", Description: "Share images to image waterfall"}, tgbotapi.BotCommand{Command: "guesstag", Description: "Guess image tags"}, tgbotapi.BotCommand{Command: "superresolution", Description: "Image super-resolution"}, tgbotapi.BotCommand{Command: "help", Description: "Help"}, tgbotapi.BotCommand{Command: "history", Description: "View history of generated images"}, tgbotapi.BotCommand{Command: "language", Description: "Set language"}, tgbotapi.BotCommand{Command: "img2tag", Description: "Image to Tag conversion"})
	cmds.LanguageCode = ""
	cmds.Scope = &bcs
	Bot.Send(cmds)

	cmds = tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{Command: "invite", Description: "获取邀请链接"}, tgbotapi.BotCommand{Command: "web", Description: "使用 Web Site 版本"}, tgbotapi.BotCommand{Command: "api", Description: "获取 API 接口文档"}, tgbotapi.BotCommand{Command: "setdefault", Description: "设置默认参数"}, tgbotapi.BotCommand{Command: "subscribe", Description: "查看订阅信息"}, tgbotapi.BotCommand{Command: "share", Description: "公开图片到图片瀑布流"}, tgbotapi.BotCommand{Command: "guesstag", Description: "猜测图片Tag"}, tgbotapi.BotCommand{Command: "superresolution", Description: "图片超分辨率"}, tgbotapi.BotCommand{Command: "help", Description: "帮助"}, tgbotapi.BotCommand{Command: "history", Description: "查看历史生成图片"}, tgbotapi.BotCommand{Command: "language", Description: "设置语言"}, tgbotapi.BotCommand{Command: "img2tag", Description: "图片转Tag"})
	cmds.LanguageCode = "zh"
	cmds.Scope = &bcs
	Bot.Send(cmds)

	cmds = tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{Command: "invite", Description: "招待リンクを取得する"}, tgbotapi.BotCommand{Command: "web", Description: "Web Site バージョンを使用する"}, tgbotapi.BotCommand{Command: "api", Description: "API インターフェースドキュメントを取得する"}, tgbotapi.BotCommand{Command: "setdefault", Description: "デフォルトパラメータを設定する"}, tgbotapi.BotCommand{Command: "subscribe", Description: "サブスクリプション情報を表示する"}, tgbotapi.BotCommand{Command: "share", Description: "画像を画像ウォーターフォールに公開する"}, tgbotapi.BotCommand{Command: "guesstag", Description: "画像タグを推測する"}, tgbotapi.BotCommand{Command: "superresolution", Description: "画像の超解像度"}, tgbotapi.BotCommand{Command: "help", Description: "ヘルプ"}, tgbotapi.BotCommand{Command: "history", Description: "過去の生成画像を見る"}, tgbotapi.BotCommand{Command: "language", Description: "言語を設定する"}, tgbotapi.BotCommand{Command: "img2tag", Description: "画像をタグに変換する"})
	cmds.LanguageCode = "ja"
	cmds.Scope = &bcs
	Bot.Send(cmds)

	cmds = tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{Command: "invite", Description: "초대 링크 얻기"}, tgbotapi.BotCommand{Command: "web", Description: "웹 사이트 버전 사용"}, tgbotapi.BotCommand{Command: "api", Description: "API 인터페이스 문서 가져오기"}, tgbotapi.BotCommand{Command: "setdefault", Description: "기본 매개변수 설정"}, tgbotapi.BotCommand{Command: "subscribe", Description: "구독 정보 확인"}, tgbotapi.BotCommand{Command: "share", Description: "이미지를 이미지 물방울 흐름에 공개"}, tgbotapi.BotCommand{Command: "guesstag", Description: "이미지 태그 추측"}, tgbotapi.BotCommand{Command: "superresolution", Description: "이미지 초고해상도"}, tgbotapi.BotCommand{Command: "help", Description: "도움말"}, tgbotapi.BotCommand{Command: "history", Description: "과거 생성된 이미지 확인"}, tgbotapi.BotCommand{Command: "language", Description: "언어 설정"}, tgbotapi.BotCommand{Command: "img2tag", Description: "이미지를 태그로 변환"})
	cmds.LanguageCode = "ko"
	cmds.Scope = &bcs
	Bot.Send(cmds)

	cmds = tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{Command: "invite", Description: "Obter link de convite"}, tgbotapi.BotCommand{Command: "web", Description: "Usar a versão do Web Site"}, tgbotapi.BotCommand{Command: "api", Description: "Obter documentação da API"}, tgbotapi.BotCommand{Command: "setdefault", Description: "Definir parâmetros padrão"}, tgbotapi.BotCommand{Command: "subscribe", Description: "Ver informações de inscrição"}, tgbotapi.BotCommand{Command: "share", Description: "Compartilhar imagem publicamente no fluxo de imagens"}, tgbotapi.BotCommand{Command: "guesstag", Description: "Adivinhar Tag da imagem"}, tgbotapi.BotCommand{Command: "superresolution", Description: "Super resolução de imagem"}, tgbotapi.BotCommand{Command: "help", Description: "Ajuda"}, tgbotapi.BotCommand{Command: "history", Description: "Ver histórico de imagens geradas"}, tgbotapi.BotCommand{Command: "language", Description: "Definir idioma"}, tgbotapi.BotCommand{Command: "img2tag", Description: "Converter imagem em Tag"})
	cmds.LanguageCode = "pt"
	cmds.Scope = &bcs
	Bot.Send(cmds)

	cmds = tgbotapi.NewSetMyCommands(tgbotapi.BotCommand{Command: "invite", Description: "получить ссылку-приглашение"}, tgbotapi.BotCommand{Command: "web", Description: "использовать версию Web Site"}, tgbotapi.BotCommand{Command: "api", Description: "получить документацию API"}, tgbotapi.BotCommand{Command: "setdefault", Description: "установить параметры по умолчанию"}, tgbotapi.BotCommand{Command: "subscribe", Description: "просмотреть информацию о подписке"}, tgbotapi.BotCommand{Command: "share", Description: "опубликовать изображение в потоке изображений"}, tgbotapi.BotCommand{Command: "guesstag", Description: "угадать тег изображения"}, tgbotapi.BotCommand{Command: "superresolution", Description: "суперразрешение изображения"}, tgbotapi.BotCommand{Command: "help", Description: "помощь"}, tgbotapi.BotCommand{Command: "history", Description: "просмотр истории созданных изображений"}, tgbotapi.BotCommand{Command: "language", Description: "установить язык"}, tgbotapi.BotCommand{Command: "img2tag", Description: "преобразование изображения в тег"})
	cmds.LanguageCode = "ru"
	cmds.Scope = &bcs
	Bot.Send(cmds)

}

func Main() {
	flag.Parse()
	db.Init()
	err := gconfig.Load(gconfig.ConfigPath)
	if err != nil {
		panic(err)
	}
	go gconfig.Wathc()
	api.Load(gconfig.API())
	go func() {
		for range gconfig.NewWatchConfig() {
			api.Load(gconfig.API())
		}
	}()
	defer ants.Release()
	mainPllo, _ = ants.NewPool(ants.DefaultAntsPoolSize, ants.WithOptions(ants.Options{ExpiryDuration: time.Minute, PreAlloc: false, Logger: nil, DisablePurge: false, Nonblocking: false, PanicHandler: func(i interface{}) {
		colorlog.Fatalf(utils.PrintStackTrace(i))
	}}))
	defer mainPllo.Release()
	Bot, err = tgbotapi.NewBotAPI(parseflag.TgToken)
	if err != nil {
		panic(fmt.Sprintf("Init Telegram Bot With Token Error: %v", err))
	}
	router.SetBot(Bot)
	router.Router()
	SetCommand()
	Bot.Buffer = 1000
	Bot.Debug = false
	var updates tgbotapi.UpdatesChannel
	if parseflag.WebHook {
		if parseflag.HOST == "" {
			panic(errors.New("flag: host is nil"))
		}
		wh, _ := tgbotapi.NewWebhook(fmt.Sprintf("https://%s/api/v1/%s", parseflag.HOST, Bot.Token))
		wh.MaxConnections = 100
		wh.DropPendingUpdates = true
		wh.AllowedUpdates = []string{"message", "callback_query"}
		_, err = Bot.Request(wh)
		if err != nil {
			colorlog.Fatalf("Request wh err: %v", err)
			panic(err)
		}
		var fun func(w http.ResponseWriter, r *http.Request)
		updates, fun = Bot.NewWebhookHandler()
		router.Eng().POST("/api/v1/"+Bot.Token, func(ctx *gin.Context) {
			fun(ctx.Writer, ctx.Request)
		})
	} else {
		Bot.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true})
		updates = Bot.GetUpdatesChan(tgbotapi.NewUpdate(0))
		updates.Clear()
	}
	go router.Eng().Run(fmt.Sprintf("127.0.0.1:%d", parseflag.Port))
	limiter := utils.NewRateLimiter(3, 1)
	colorlog.Infof("Service started successfully!")
	for update := range updates {
		if update.Message != nil && update.Message.Chat.ID > 0 {
			mainPllo.Submit(func() {
				if !limiter.GetLimiter(update.Message.From.ID).Allow() {
					return
				}
				if update.Message.IsCommand() {
					colorlog.Infof("Get the message cmd [%s] : %s", update.Message.From.String(), update.Message.Command())
					handler.HandleCmd(Bot, *update.Message)
				} else if msgChan, ok := Bot.FindMsgCbk(update.Message.Chat.ID, update.Message.From.ID); ok {
					select {
					case msgChan.MsgChan() <- update.Message:
						colorlog.Infof("Get the message cbk [%s] : %s", update.Message.From.String(), update.Message.Text)
					default:
					}
				} else {
					colorlog.Infof("Get the message [%s] : %s%s", update.Message.From.String(), update.Message.Text, update.Message.Caption)
					handler.HandleMsg(Bot, update.Message)
				}
			})
		} else if update.CallbackQuery != nil && update.CallbackQuery.Message.Chat.ID > 0 {
			mainPllo.Submit(func() {
				if !limiter.GetLimiter(update.CallbackQuery.From.ID).Allow() {
					return
				}
				colorlog.Infof("Get the Callback [%s] : %s", update.CallbackQuery.From.String(), update.CallbackQuery.Data)
				handler.HandleCallback(Bot, update.CallbackQuery)
			})
		}
	}
}
