package i18n

import (
	"embed"
	"fmt"

	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"gopkg.in/yaml.v3"
)

//go:embed extra
var extra embed.FS

var lang = map[string]text{}

var langList = []text{}

type text struct {
	language map[string]string
	Code     string
	Name     string
}

func init() {
	extras, err := extra.ReadDir("extra")
	if err != nil {
		panic(err)
	}
	for _, meta := range extras {
		if !meta.IsDir() {
			data, err := extra.ReadFile(fmt.Sprintf("extra/%s", meta.Name()))
			if err != nil {
				panic(err)
			}
			i18nMap := make(map[string]string)
			err = yaml.Unmarshal(data, i18nMap)
			if err != nil {
				panic(err)
			}
			code := utils.GetFileNamePrefix(meta.Name())
			lang[fmt.Sprintf("%s:Extra", code)] = text{language: i18nMap, Code: fmt.Sprintf("%s:Extra", code)}
		}
	}
}

func register(l text) {
	lang[l.Code] = l
	langList = append(langList, l)
}

func LoadLang(langType, tag string) string {
	l := lang[langType].language[tag]
	if l != "" {
		return l
	} else {
		return lang["en_us"].language[tag]
	}
}

func LoadExtraLang(langType, tag string) string {
	if langType == "en_us" {
		return tag
	}
	l := lang[fmt.Sprintf("%s:Extra", langType)].language[tag]
	if l != "" {
		return l
	} else {
		return tag
	}
}

func LoadAllExtraLang(langType string) map[string]string {
	if l, ok := lang[fmt.Sprintf("%s:Extra", langType)]; ok {
		return l.language
	} else {
		return nil
	}
}

func LangList() []text {
	return langList
}
