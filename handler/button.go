package handler

import (
	"fmt"

	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/i18n"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
	"github.com/zijiren233/stable-diffusion-webui-bot/user"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	tgbotapi "github.com/zijiren233/tg-bot-api/v6"
)

func goJoinButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return clictUrlButton(u, gconfig.GROUP())
}

func clictUrlButton(u *user.UserInfo, url string) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL(u.LoadLang("clickMe"), url),
	)}}
}

var poolButton = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("刷新", "drawpool"),
	),
)

func goGuideButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return clictUrlButton(u, gconfig.GUIDE())
}

func panelButton(u *user.UserInfo, photo, control bool) *tgbotapi.InlineKeyboardMarkup {
	var row = [][]tgbotapi.InlineKeyboardButton{}
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("editTag"), "setCfg:editTag:0"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("mode"), "panel:mode"),
	})
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("extraModel"), "setCfg:extraModelGroup:1"),
	})
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("number"), "panel:num"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("steps"), "panel:steps"),
	})
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("scale"), "panel:scale"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("size"), "panel:size"),
	})
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("model"), "panel:model"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("unwanted"), "setCfg:editUc:0"),
	})
	if photo {
		row = append(row, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("strength"), "panel:strength"),
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("clearImg"), "setCfg:setImg"),
		})
		row = append(row, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("resetSeed"), "setCfg:resetSeed"),
		})
	} else {
		row = append(row, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("setImg"), "setCfg:setImg"),
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("resetSeed"), "setCfg:resetSeed"),
		})
	}
	if control {
		row = append(row, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("editControl"), "setCfg:editControl"),
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("delControl"), "setCfg:setControl"),
		})
	} else {
		row = append(row, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("setControl"), "setCfg:setControl"),
		})
	}
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "panel:confirm"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "delete:cancel"),
	})
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func setDefaultCfg(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 3)
	row[0] = []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("mode"), "default:mode"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("unwanted"), "default:uc"),
	}
	row[1] = []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("number"), "default:number"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("scale"), "default:scale"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("steps"), "default:steps"),
	}
	row[2] = []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "delete:cancel"),
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateSetDftMODEButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	models := api.AllMode()
	lens := len(models) / MAXROW
	if len(models)%MAXROW != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens+1)
	rows := 0
	for k, v := range models {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if u.UserInfo.UserDefaultMODE == v {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v), fmt.Sprint("setDft:mode:", v)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v, fmt.Sprint("setDft:mode:", v)))
		}
	}
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "default:panel"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateSetDftUCButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 3)
	row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(u.LoadLang("setDft"), u.LoadLang("unwanted")), "setDft:uc"))
	row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("reset"), "setDft:uc:reset"))
	row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "default:panel"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateSetDftNumberButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	var row = [][]tgbotapi.InlineKeyboardButton{}
	i := 0
	for v := 1; v <= api.MaxNum; v++ {
		if v != 1 && (v-1)%MAXROW == 0 {
			i++
		}
		if len(row) < i+1 {
			row = append(row, []tgbotapi.InlineKeyboardButton{})
		}
		if u.UserInfo.UserDefaultNumber == v {
			row[i] = append(row[i], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", v), fmt.Sprint("setDft:number:", v)))
		} else {
			row[i] = append(row[i], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(v), fmt.Sprint("setDft:number:", v)))
		}
	}
	row = append(row, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "default:panel")})
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateSetDftScaleButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("-", "setDft:scale:-"),
				tgbotapi.NewInlineKeyboardButtonData("+", "setDft:scale:+"),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "default:panel"),
			},
		},
	}
}

func generateSetDftStepsButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("-", "setDft:steps:-"),
				tgbotapi.NewInlineKeyboardButtonData("+", "setDft:steps:+"),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "default:panel"),
			},
		},
	}
}

func editControlButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	var row = [][]tgbotapi.InlineKeyboardButton{}
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("controlPreprocess"), "setCfg:controlPreprocess"),
	})
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("controlProcess"), "setCfg:controlProcess"),
	})
	row = append(row, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("delControl"), "setCfg:setControl"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"),
	})
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func controlPreprocessButton(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	models := gconfig.PreProcess()
	lens := len(models) / MAXROW
	if len(models)%MAXROW != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens)
	rows := 0
	for k, v := range models {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if option == v.Name {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, u.LoadLang(v.Name)), fmt.Sprint("setCfg:preprocess:", v.Name)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang(v.Name), fmt.Sprint("setCfg:preprocess:", v.Name)))
		}
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func controlProcessButton(option string) *tgbotapi.InlineKeyboardMarkup {
	models := gconfig.Process()
	lens := len(models) / MAXROW
	if len(models)%MAXROW != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens)
	rows := 0
	for k, v := range models {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if option == v.Name {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v.Name), fmt.Sprint("setCfg:process:", v.Name)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v.Name, fmt.Sprint("setCfg:process:", v.Name)))
		}
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func cancelButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "delete:cancel"),
	)}}
}

func sprButton(u *user.UserInfo, option int) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 2)
	if option == 2 {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 2", "cmd-spr:2"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("2", "cmd-spr:2"))
	}
	if option == 3 {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 3", "cmd-spr:3"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("3", "cmd-spr:3"))
	}
	if option == 4 {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 4", "cmd-spr:4"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("4", "cmd-spr:4"))
	}
	row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "cancel:cancel"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateNUMButton(option int) *tgbotapi.InlineKeyboardMarkup {
	var row = [][]tgbotapi.InlineKeyboardButton{}
	i := 0
	for v := 1; v <= api.MaxNum; v++ {
		if v != 1 && (v-1)%MAXROW == 0 {
			i++
		}
		if len(row) < i+1 {
			row = append(row, []tgbotapi.InlineKeyboardButton{})
		}
		if option == v {
			row[i] = append(row[i], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", v), fmt.Sprint("setCfg:num:", v)))
		} else {
			row[i] = append(row[i], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(v), fmt.Sprint("setCfg:num:", v)))
		}
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateMODEButton(option string) *tgbotapi.InlineKeyboardMarkup {
	models := api.AllMode()
	lens := len(models) / MAXROW
	if len(models)%MAXROW != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens)
	rows := 0
	for k, v := range models {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if option == v {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v), fmt.Sprint("setCfg:mode:", v)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v, fmt.Sprint("setCfg:mode:", v)))
		}
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateModelButton(name string) *tgbotapi.InlineKeyboardMarkup {
	models := gconfig.MODELS()
	lens := len(models) / 3
	if len(models)%3 != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens)
	rows := 0
	for k, v := range models {
		if k != 0 && k%3 == 0 {
			rows += 1
		}
		if name == v.Name {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v.Name), fmt.Sprint("setCfg:model:", v.Name)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v.Name, fmt.Sprint("setCfg:model:", v.Name)))
		}
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

const MAXROW = 3
const MAXONEPAGEOBJ = MAXROW * 7

func generateAllExtraModelButton(u *user.UserInfo, page, groupIndex int, options []string) *tgbotapi.InlineKeyboardMarkup {
	loras := gconfig.GroupIndex2ExtraModels(groupIndex)
	if page == 0 {
		page = 1
	} else if page < 0 {
		page = -page
	}
	all := len(loras)
	var maxPage = all / MAXONEPAGEOBJ
	if all%MAXONEPAGEOBJ != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	index := (page - 1) * MAXONEPAGEOBJ
	if page*MAXONEPAGEOBJ > all {
		loras = loras[(page-1)*MAXONEPAGEOBJ:]
	} else if all != 0 {
		loras = loras[(page-1)*MAXONEPAGEOBJ : page*MAXONEPAGEOBJ]
	}
	all = len(loras)
	lens := all / MAXROW
	if all%MAXROW != 0 {
		lens += 1
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 || len(loras) == 0 {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+2)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+3)
	}
	rows := 0
	for k, v := range loras {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if _, ok := utils.InString(v.Name, options); ok {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, u.LoadExtraLang(v.Name)), fmt.Sprint("setCfg:sL:", groupIndex, ":", index, ":-", page)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadExtraLang(v.Name), fmt.Sprint("setCfg:sL:", groupIndex, ":", index, ":-", page)))
		}
		index += 1
	}
	if loras != nil {
		rows += 1
	}
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:extraModel:", groupIndex, ":-", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:extraModel:", groupIndex, ":-", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:extraModel:", groupIndex, ":-", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:extraModel:", groupIndex, ":-", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("switch"), fmt.Sprint("setCfg:extraModel:", groupIndex, ":1")))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "setCfg:extraModelGroup:0"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateAllExtraModelGroupButton(u *user.UserInfo, page int) *tgbotapi.InlineKeyboardMarkup {
	groups := gconfig.ExtraModelGroup()
	if page == 0 {
		page = 1
	} else if page < 0 {
		page = -page
	}
	all := len(groups)
	var maxPage = all / MAXONEPAGEOBJ
	if all%MAXONEPAGEOBJ != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	var index int = (page - 1) * MAXONEPAGEOBJ
	if page*MAXONEPAGEOBJ > all {
		groups = groups[(page-1)*MAXONEPAGEOBJ:]
	} else if all != 0 {
		groups = groups[(page-1)*MAXONEPAGEOBJ : page*MAXONEPAGEOBJ]
	}
	all = len(groups)
	lens := all / MAXROW
	if all%MAXROW != 0 {
		lens += 1
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 || len(groups) == 0 {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+1)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+2)
	}
	rows := 0
	for k, v := range groups {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadExtraLang(v), fmt.Sprint("setCfg:extraModel:", index, ":-", 1)))
		index += 1
	}
	if groups != nil {
		rows += 1
	}
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:extraModelGroup:-", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:extraModelGroup:-", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:extraModelGroup:-", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:extraModelGroup:-", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateExtraModelButton(u *user.UserInfo, page, groupIndex int, options []string) *tgbotapi.InlineKeyboardMarkup {
	models := gconfig.GroupIndex2ExtraModels(groupIndex)
	if len(models) == 0 || page <= 0 {
		return generateAllExtraModelButton(u, page, groupIndex, options)
	}
	all := len(models)
	var maxPage = all / MAXROW
	if all%MAXROW != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 {
		row = make([][]tgbotapi.InlineKeyboardButton, 5)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, 6)
	}
	var model []gconfig.ExtraModel
	index := (page - 1) * MAXROW
	if page == maxPage {
		model = models[(page-1)*MAXROW:]
	} else {
		model = models[(page-1)*MAXROW : page*MAXROW]
	}
	for _, v := range model {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("+", fmt.Sprint("setCfg:L:", groupIndex, ":", index, ":+", ":", page)))
		if _, ok := utils.InString(v.Name, options); ok {
			row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, u.LoadExtraLang(v.Name)), fmt.Sprint("setCfg:sL:", groupIndex, ":", index, ":", page)))
		} else {
			row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(u.LoadExtraLang(v.Name), fmt.Sprint("setCfg:sL:", groupIndex, ":", index, ":", page)))
		}
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("-", fmt.Sprint("setCfg:L:", groupIndex, ":", index, ":-", ":", page)))
		index += 1
	}
	rows := 3
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:extraModel:", groupIndex, ":", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:extraModel:", groupIndex, ":", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:extraModel:", groupIndex, ":", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:extraModel:", groupIndex, ":", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("switch"), fmt.Sprint("setCfg:extraModel:", groupIndex, ":", 0)))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), "setCfg:extraModelGroup:0"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateAllTagButton(u *user.UserInfo, page int, options, allTag []string) *tgbotapi.InlineKeyboardMarkup {
	if page == 0 {
		page = 1
	} else if page < 0 {
		page = -page
	}
	all := len(allTag)
	var maxPage = all / MAXONEPAGEOBJ
	if all%MAXONEPAGEOBJ != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	if page*MAXONEPAGEOBJ > all {
		allTag = allTag[(page-1)*MAXONEPAGEOBJ:]
	} else if all != 0 {
		allTag = allTag[(page-1)*MAXONEPAGEOBJ : page*MAXONEPAGEOBJ]
	}
	all = len(allTag)
	lens := all / MAXROW
	if all%MAXROW != 0 {
		lens += 1
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 || len(allTag) == 0 {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+4)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+5)
	}
	rows := 0
	for k, v := range allTag {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if _, ok := utils.InString(v, options); ok {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v), fmt.Sprint("setCfg:sT:", v, ":-", page)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`❌ `, v), fmt.Sprint("setCfg:sT:", v, ":-", page)))
		}
	}
	if allTag != nil {
		rows += 1
	}
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:editTag:-", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:editTag:-", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:editTag:-", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:editTag:-", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("switch"), "setCfg:editTag:1"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Happend"), fmt.Sprint("setCfg:changeTag:Happend:-", page)))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Eappend"), fmt.Sprint("setCfg:changeTag:Eappend:-", page)))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("reset"), fmt.Sprint("setCfg:changeTag:reset:-", page)))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("translation"), fmt.Sprint("setCfg:changeTag:translation:-", page)))
	row[rows+3] = append(row[rows+3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func editTagButton(u *user.UserInfo, page int, options, allTag []string) *tgbotapi.InlineKeyboardMarkup {
	if len(allTag) == 0 || page <= 0 {
		return generateAllTagButton(u, page, options, allTag)
	}
	all := len(allTag)
	var maxPage = all / MAXROW
	if all%MAXROW != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 {
		row = make([][]tgbotapi.InlineKeyboardButton, 7)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, 8)
	}
	var model []string
	if page == maxPage {
		model = allTag[(page-1)*MAXROW:]
	} else {
		model = allTag[(page-1)*MAXROW : page*MAXROW]
	}
	for _, v := range model {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("+", fmt.Sprint("setCfg:T:", v, ":+", ":", page)))
	}
	for _, v := range model {
		if _, ok := utils.InString(v, options); ok {
			row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v), fmt.Sprint("setCfg:sT:", v, ":", page)))
		} else {
			row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(v, fmt.Sprint("setCfg:sT:", v, ":", page)))
		}
	}
	for _, v := range model {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("-", fmt.Sprint("setCfg:T:", v, ":-", ":", page)))
	}
	rows := 3
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:editTag:", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:editTag:", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:editTag:", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:editTag:", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("switch"), "setCfg:editTag:0"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Happend"), "setCfg:changeTag:Happend:1"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Eappend"), "setCfg:changeTag:Eappend:1"))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("reset"), "setCfg:changeTag:reset:1"))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("translation"), "setCfg:changeTag:translation:1"))
	row[rows+3] = append(row[rows+3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generateAllUcButton(u *user.UserInfo, page int, options, allUc []string) *tgbotapi.InlineKeyboardMarkup {
	if page == 0 {
		page = 1
	} else if page < 0 {
		page = -page
	}
	all := len(allUc)
	var maxPage = all / MAXONEPAGEOBJ
	if all%MAXONEPAGEOBJ != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	if page*MAXONEPAGEOBJ > all {
		allUc = allUc[(page-1)*MAXONEPAGEOBJ:]
	} else if all != 0 {
		allUc = allUc[(page-1)*MAXONEPAGEOBJ : page*MAXONEPAGEOBJ]
	}
	all = len(allUc)
	lens := all / MAXROW
	if all%MAXROW != 0 {
		lens += 1
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 || len(allUc) == 0 {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+4)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, lens+5)
	}
	rows := 0
	for k, v := range allUc {
		if k != 0 && k%MAXROW == 0 {
			rows += 1
		}
		if _, ok := utils.InString(v, options); ok {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v), fmt.Sprint("setCfg:sU:", v, ":-", page)))
		} else {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v, fmt.Sprint("setCfg:sU:", v, ":-", page)))
		}
	}
	if allUc != nil {
		rows += 1
	}
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:editUc:-", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:editUc:-", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:editUc:-", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:editUc:-", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("switch"), "setCfg:editUc:1"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Happend"), fmt.Sprintf("setCfg:changeUc:Happend:-%d", page)))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Eappend"), fmt.Sprintf("setCfg:changeUc:Eappend:-%d", page)))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("reset"), fmt.Sprintf("setCfg:changeUc:reset:-%d", page)))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("translation"), fmt.Sprintf("setCfg:changeUc:translation:-%d", page)))
	row[rows+3] = append(row[rows+3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func editUcButton(u *user.UserInfo, page int, options, allUc []string) *tgbotapi.InlineKeyboardMarkup {
	if len(allUc) == 0 || page <= 0 {
		return generateAllUcButton(u, page, options, allUc)
	}
	var maxPage = len(allUc) / MAXROW
	if len(allUc)%MAXROW != 0 {
		maxPage += 1
	}
	if page > maxPage {
		page = maxPage
	}
	var row [][]tgbotapi.InlineKeyboardButton
	if maxPage == 1 {
		row = make([][]tgbotapi.InlineKeyboardButton, 7)
	} else {
		row = make([][]tgbotapi.InlineKeyboardButton, 8)
	}
	var model []string
	if page == maxPage {
		model = allUc[(page-1)*MAXROW:]
	} else {
		model = allUc[(page-1)*MAXROW : page*MAXROW]
	}
	for _, v := range model {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("+", fmt.Sprint("setCfg:uc:", v, ":+", ":", page)))
	}
	for _, v := range model {
		if _, ok := utils.InString(v, options); ok {
			row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(`✅ `, v), fmt.Sprint("setCfg:sU:", v, ":", page)))
		} else {
			row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData(v, fmt.Sprint("setCfg:sU:", v, ":", page)))
		}
	}
	for _, v := range model {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("-", fmt.Sprint("setCfg:uc:", v, ":-", ":", page)))
	}
	rows := 3
	if maxPage > 1 {
		var minT bool
		var minP = 1
		if page-4 > 0 {
			minT = true
			minP = page - 2
		}
		var maxT bool
		var maxP = maxPage
		if page+4 <= maxPage {
			maxT = true
			maxP = page + 2
		}
		if minT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(1, "..."), fmt.Sprint("setCfg:editUc:", 1)))
		}
		for i := minP; i <= maxP; i++ {
			if page == i {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("✅ ", i), fmt.Sprint("setCfg:editUc:", i)))
			} else {
				row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(i), fmt.Sprint("setCfg:editUc:", i)))
			}
		}
		if maxT {
			row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint("...", maxPage), fmt.Sprint("setCfg:editUc:", maxPage)))
		}
		rows += 1
	}
	row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("switch"), "setCfg:editUc:0"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Happend"), "setCfg:changeUc:Happend:1"))
	row[rows+1] = append(row[rows+1], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("Eappend"), "setCfg:changeUc:Eappend:1"))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("reset"), "setCfg:changeUc:reset:1"))
	row[rows+2] = append(row[rows+2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("translation"), "setCfg:changeUc:translation:1"))
	row[rows+3] = append(row[rows+3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func imgButton(u *user.UserInfo, superResolution bool) *tgbotapi.InlineKeyboardMarkup {
	if superResolution {
		return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("edit"), "editImg"),
		),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("FT:1", "fineTune:1"),
				tgbotapi.NewInlineKeyboardButtonData("FT:2", "fineTune:2"),
				tgbotapi.NewInlineKeyboardButtonData("FT:3", "fineTune:3"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("SPR:2", "spr:2"),
				tgbotapi.NewInlineKeyboardButtonData("SPR:3", "spr:3"),
				tgbotapi.NewInlineKeyboardButtonData("SPR:4", "spr:4"),
			)}}
	} else {
		return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("edit"), "editImg"),
		),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("FT:1", "fineTune:1"),
				tgbotapi.NewInlineKeyboardButtonData("FT:2", "fineTune:2"),
				tgbotapi.NewInlineKeyboardButtonData("FT:3", "fineTune:3"),
			)}}
	}
}

func sizeTypeButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("1:1", "setCfg:sizeType:1:1"),
		tgbotapi.NewInlineKeyboardButtonData("3:2", "setCfg:sizeType:3:2"),
		tgbotapi.NewInlineKeyboardButtonData("2:3", "setCfg:sizeType:2:3"),
	),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("16:9", "setCfg:sizeType:16:9"),
			tgbotapi.NewInlineKeyboardButtonData("9:16", "setCfg:sizeType:9:16"),
			tgbotapi.NewInlineKeyboardButtonData("4:3", "setCfg:sizeType:4:3"),
			tgbotapi.NewInlineKeyboardButtonData("3:4", "setCfg:sizeType:3:4"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("custom"), `setCfg:sizeType:custom`),
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), "setCfg:confirm:confirm"),
		)}}
}

func custonSizeButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("W+8", "setCfg:custonSize:w+8"),
		tgbotapi.NewInlineKeyboardButtonData("W+32", "setCfg:custonSize:w+32"),
		tgbotapi.NewInlineKeyboardButtonData("W+64", "setCfg:custonSize:w+64"),
		tgbotapi.NewInlineKeyboardButtonData("W+128", "setCfg:custonSize:w+128"),
	),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("W-8", "setCfg:custonSize:w-8"),
			tgbotapi.NewInlineKeyboardButtonData("W-32", "setCfg:custonSize:w-32"),
			tgbotapi.NewInlineKeyboardButtonData("W-64", "setCfg:custonSize:w-64"),
			tgbotapi.NewInlineKeyboardButtonData("W-128", "setCfg:custonSize:w-128"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("H+8", "setCfg:custonSize:h+8"),
			tgbotapi.NewInlineKeyboardButtonData("H+32", "setCfg:custonSize:h+32"),
			tgbotapi.NewInlineKeyboardButtonData("H+64", "setCfg:custonSize:h+64"),
			tgbotapi.NewInlineKeyboardButtonData("H+128", "setCfg:custonSize:h+128"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("H-8", "setCfg:custonSize:h-8"),
			tgbotapi.NewInlineKeyboardButtonData("H-32", "setCfg:custonSize:h-32"),
			tgbotapi.NewInlineKeyboardButtonData("H-64", "setCfg:custonSize:h-64"),
			tgbotapi.NewInlineKeyboardButtonData("H-128", "setCfg:custonSize:h-128"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`),
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), `setCfg:confirm:confirm`),
		)}}
}

func generate3_2Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 4)
	if option == "576*384" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 576*384", "setCfg:size:576*384"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("576*384", "setCfg:size:576*384"))
	}
	if option == "768*512" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 768*512", "setCfg:size:768*512"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("768*512", "setCfg:size:768*512"))
	}
	if option == "960*640" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 960*640", "setCfg:size:960*640"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("960*640", "setCfg:size:960*640"))
	}
	if option == "1152*768" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1152*768", "setCfg:size:1152*768"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1152*768", "setCfg:size:1152*768"))
	}
	if option == "1344*896" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1344*896", "setCfg:size:1344*896"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1344*896", "setCfg:size:1344*896"))
	}
	if option == "1440*960" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1440*960", "setCfg:size:1440*960"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1440*960", "setCfg:size:1440*960"))
	}
	if option == "1536*1024" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1536*1024", "setCfg:size:1536*1024"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1536*1024", "setCfg:size:1536*1024"))
	}
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generate2_3Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 4)
	if option == "384*576" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 384*576", "setCfg:size:384*576"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("384*576", "setCfg:size:384*576"))
	}
	if option == "512*768" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 512*768", "setCfg:size:512*768"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("512*768", "setCfg:size:512*768"))
	}
	if option == "640*960" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 640*960", "setCfg:size:640*960"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("640*960", "setCfg:size:640*960"))
	}
	if option == "768*1152" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 768*1152", "setCfg:size:768*1152"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("768*1152", "setCfg:size:768*1152"))
	}
	if option == "896*1344" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 896*1344", "setCfg:size:896*1344"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("896*1344", "setCfg:size:896*1344"))
	}
	if option == "960*1440" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 960*1440", "setCfg:size:960*1440"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("960*1440", "setCfg:size:960*1440"))
	}
	if option == "1024*1536" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1024*1536", "setCfg:size:1024*1536"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1024*1536", "setCfg:size:1024*1536"))
	}
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generate4_3Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 4)
	if option == "512*384" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 512*384", "setCfg:size:512*384"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("512*384", "setCfg:size:512*384"))
	}
	if option == "704*512" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 704*512", "setCfg:size:704*512"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("704*512", "setCfg:size:704*512"))
	}
	if option == "832*640" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 832*640", "setCfg:size:832*640"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("832*640", "setCfg:size:832*640"))
	}
	if option == "1024*768" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1024*768", "setCfg:size:1024*768"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1024*768", "setCfg:size:1024*768"))
	}
	if option == "1192*896" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1192*896", "setCfg:size:1192*896"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1192*896", "setCfg:size:1192*896"))
	}
	if option == "1280*960" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1280*960", "setCfg:size:1280*960"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1280*960", "setCfg:size:1280*960"))
	}
	if option == "1368*1024" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1368*1024", "setCfg:size:1368*1024"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1368*1024", "setCfg:size:1368*1024"))
	}
	if option == "1440*1088" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1440*1088", "setCfg:size:1440*1088"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1440*1088", "setCfg:size:1440*1088"))
	}
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generate3_4Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 4)
	if option == "384*512" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 384*512", "setCfg:size:384*512"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("384*512", "setCfg:size:384*512"))
	}
	if option == "512*704" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 512*704", "setCfg:size:512*704"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("512*704", "setCfg:size:512*704"))
	}
	if option == "640*832" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 640*832", "setCfg:size:640*832"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("640*832", "setCfg:size:640*832"))
	}
	if option == "768*1024" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 768*1024", "setCfg:size:768*1024"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("768*1024", "setCfg:size:768*1024"))
	}
	if option == "896*1192" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 896*1192", "setCfg:size:896*1192"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("896*1192", "setCfg:size:896*1192"))
	}
	if option == "960*1280" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 960*1280", "setCfg:size:960*1280"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("960*1280", "setCfg:size:960*1280"))
	}
	if option == "1024*1368" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1024*1368", "setCfg:size:1024*1368"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1024*1368", "setCfg:size:1024*1368"))
	}
	if option == "1088*1440" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1088*1440", "setCfg:size:1088*1440"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1088*1440", "setCfg:size:1088*1440"))
	}
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generate16_9Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 3)
	if option == "640*384" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 640*384", "setCfg:size:640*384"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("640*384", "setCfg:size:640*384"))
	}
	if option == "896*512" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 896*512", "setCfg:size:896*512"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("896*512", "setCfg:size:896*512"))
	}
	if option == "1152*640" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 1152*640", "setCfg:size:1152*640"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("1152*640", "setCfg:size:1152*640"))
	}
	if option == "1344*768" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1344*768", "setCfg:size:1344*768"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1344*768", "setCfg:size:1344*768"))
	}
	if option == "1592*896" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1592*896", "setCfg:size:1592*896"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1592*896", "setCfg:size:1592*896"))
	}
	row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: row,
	}
}

func generate9_16Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 3)
	if option == "384*640" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 384*640", "setCfg:size:384*640"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("384*640", "setCfg:size:384*640"))
	}
	if option == "512*896" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 512*896", "setCfg:size:512*896"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("512*896", "setCfg:size:512*896"))
	}
	if option == "640*1152" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 640*1152", "setCfg:size:640*1152"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("640*1152", "setCfg:size:640*1152"))
	}
	if option == "768*1344" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 768*1344", "setCfg:size:768*1344"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("768*1344", "setCfg:size:768*1344"))
	}
	if option == "896*1592" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 896*1592", "setCfg:size:896*1592"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("896*1592", "setCfg:size:896*1592"))
	}
	row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func generate1_1Button(u *user.UserInfo, option string) *tgbotapi.InlineKeyboardMarkup {
	var row = make([][]tgbotapi.InlineKeyboardButton, 4)
	if option == "384*384" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 384*384", "setCfg:size:384*384"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("384*384", "setCfg:size:384*384"))
	}
	if option == "512*512" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 512*512", "setCfg:size:512*512"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("512*512", "setCfg:size:512*512"))
	}
	if option == "640*640" {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("✅ 640*640", "setCfg:size:640*640"))
	} else {
		row[0] = append(row[0], tgbotapi.NewInlineKeyboardButtonData("640*640", "setCfg:size:640*640"))
	}
	if option == "768*768" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 768*768", "setCfg:size:768*768"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("768*768", "setCfg:size:768*768"))
	}
	if option == "896*896" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 896*896", "setCfg:size:896*896"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("896*896", "setCfg:size:896*896"))
	}
	if option == "1024*1024" {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("✅ 1024*1024", "setCfg:size:1024*1024"))
	} else {
		row[1] = append(row[1], tgbotapi.NewInlineKeyboardButtonData("1024*1024", "setCfg:size:1024*1024"))
	}
	if option == "1152*1152" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1152*1152", "setCfg:size:1152*1152"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1152*1152", "setCfg:size:1152*1152"))
	}
	if option == "1216*1216" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1216*1216", "setCfg:size:1216*1216"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1216*1216", "setCfg:size:1216*1216"))
	}
	if option == "1280*1280" {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("✅ 1280*1280", "setCfg:size:1280*1280"))
	} else {
		row[2] = append(row[2], tgbotapi.NewInlineKeyboardButtonData("1280*1280", "setCfg:size:1280*1280"))
	}
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("back"), `panel:size`))
	row[3] = append(row[3], tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("cancel"), "setCfg:confirm:confirm"))
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func scaleButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("+1", "setCfg:scale:+1"),
		tgbotapi.NewInlineKeyboardButtonData("+3", "setCfg:scale:+3"),
		tgbotapi.NewInlineKeyboardButtonData("+5", "setCfg:scale:+5"),
	),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("-1", "setCfg:scale:-1"),
			tgbotapi.NewInlineKeyboardButtonData("-3", "setCfg:scale:-3"),
			tgbotapi.NewInlineKeyboardButtonData("-5", "setCfg:scale:-5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), `setCfg:confirm:confirm`),
		)}}
}

func stepsButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("+1", "setCfg:steps:+1"),
		tgbotapi.NewInlineKeyboardButtonData("+3", "setCfg:steps:+3"),
		tgbotapi.NewInlineKeyboardButtonData("+5", "setCfg:steps:+5"),
	),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("-1", "setCfg:steps:-1"),
			tgbotapi.NewInlineKeyboardButtonData("-3", "setCfg:steps:-3"),
			tgbotapi.NewInlineKeyboardButtonData("-5", "setCfg:steps:-5"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), `setCfg:confirm:confirm`),
		)}}
}

var langButton = gLangButton()

func gLangButton() *tgbotapi.InlineKeyboardMarkup {
	langs := i18n.LangList()
	lens := len(langs) / 3
	if len(langs)%3 != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens)
	rows := 0
	for k, v := range langs {
		if k != 0 && k%3 == 0 {
			rows += 1
		}
		row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v.Name, fmt.Sprint("lang:", v.Code)))
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

var helpLangButton = gHelpLangButton()

func gHelpLangButton() *tgbotapi.InlineKeyboardMarkup {
	langs := i18n.LangList()
	lens := len(langs) / 3
	if len(langs)%3 != 0 {
		lens += 1
	}
	var row = make([][]tgbotapi.InlineKeyboardButton, lens)
	rows := 0
	for k, v := range langs {
		if k != 0 && k%3 == 0 {
			rows += 1
		}
		row[rows] = append(row[rows], tgbotapi.NewInlineKeyboardButtonData(v.Name, fmt.Sprint("helpLang:", v.Code)))
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: row}
}

func gShareButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("enable"), `share:1`),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("disable"), `share:0`),
	)}}
}

func strengthButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(`+0.01`, "setCfg:strength:+0.01"),
		tgbotapi.NewInlineKeyboardButtonData(`+0.05`, "setCfg:strength:+0.05"),
		tgbotapi.NewInlineKeyboardButtonData(`+0.1`, "setCfg:strength:+0.1"),
		tgbotapi.NewInlineKeyboardButtonData(`+0.3`, "setCfg:strength:+0.3"),
	),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(`-0.01`, "setCfg:strength:-0.01"),
			tgbotapi.NewInlineKeyboardButtonData(`-0.05`, "setCfg:strength:-0.05"),
			tgbotapi.NewInlineKeyboardButtonData(`-0.1`, "setCfg:strength:-0.1"),
			tgbotapi.NewInlineKeyboardButtonData(`-0.3`, "setCfg:strength:-0.3"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("confirm"), `setCfg:confirm:confirm`),
		)}}

}

func reDrawButton(u *user.UserInfo) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("reDraw"), "reDraw"),
		tgbotapi.NewInlineKeyboardButtonData(u.LoadLang("edit"), "editCfg"),
	)}}
}
