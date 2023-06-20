package i18n

import (
	"fmt"
	"os"
	"path"

	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"gopkg.in/yaml.v3"
)

var lang = map[string]text{}

var langList = []text{}

var defaultCode = ""

type text struct {
	language map[string]string
	Code     string
	Name     string
	Default  bool
}

func init() {
	if err := os.MkdirAll(parseflag.I18nExtraPath, os.ModePerm); err != nil {
		panic(err)
	}
	extras, err := os.ReadDir(parseflag.I18nExtraPath)
	if err != nil {
		panic(err)
	}
	for _, meta := range extras {
		if !meta.IsDir() {
			file, err := os.OpenFile(path.Join(parseflag.I18nExtraPath, meta.Name()), os.O_CREATE|os.O_RDONLY, os.ModePerm)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			i18nMap := make(map[string]string)
			if err = yaml.NewDecoder(file).Decode(i18nMap); err != nil {
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
	if l.Default {
		defaultCode = l.Code
	}
}

func LoadLang(langType, tag string) string {
	l := lang[langType].language[tag]
	if l != "" {
		return l
	} else {
		return lang[defaultCode].language[tag]
	}
}

func LoadExtraLang(langType, tag string) string {
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
