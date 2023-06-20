package gconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/zijiren233/go-colorlog"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"gopkg.in/yaml.v3"
)

var lock sync.RWMutex

const ConfigPath = "./config.yaml"

var config Config

var extraModelAllGroup []string

// group -> []Lora
var extraModel = make(map[string][]ExtraModel)

type Config struct {
	Api               []Api
	Model             []Model
	Embedding         []Embedding
	ExtraModel        []ExtraModel
	GroupID           int64
	Group             string
	Guide             string
	ControlPreProcess []ControlPreProcess
	ControlProcess    []ControlProcess
}

type Api struct {
	Url, Username, Password string
	Cached                  int
}

type Model struct {
	Name     string
	File     string
	FileExt  string
	Vae      string
	VaeExt   string
	ClipSkip int
}

type Embedding struct {
	Name    string
	NameExt string
	Type    uint
	Note    string
}

type ExtraModel struct {
	Name         string
	NameExt      string
	Preview      string
	Type         string
	TriggerWords []string
	Group        []string
}

type ControlProcess struct {
	Name string
	File string
	Note string
}

type ControlPreProcess struct {
	Name string
}

func init() {
	go watch(ConfigPath)
}

func Load(configPath string) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	tmp := new(Config)
	if err := yaml.Unmarshal(b, tmp); err != nil {
		return err
	}
	if err := checkDuplicates(tmp.Api); err != nil {
		return err
	}
	if err := checkDuplicatesExtraModel(tmp.ExtraModel); err != nil {
		return err
	}
	for i := 0; i < len(tmp.Api); i++ {
		tmp.Api[i].Url = strings.TrimRight(tmp.Api[i].Url, "/")
	}
	config = *tmp
	extraModelAllGroup = []string{}
	extraModel = make(map[string][]ExtraModel)
	sort.Sort(SortExtraModel(config.ExtraModel))
	for _, v := range config.ExtraModel {
		if v.Type == "" {
			v.Type = "lora"
		}
		for _, g := range v.Group {
			extraModel[g] = append(extraModel[g], v)
			if _, ok := utils.InString(g, extraModelAllGroup); !ok {
				extraModelAllGroup = append(extraModelAllGroup, g)
			}
		}
	}
	sort.Strings(extraModelAllGroup)
	sort.Sort(SortEmbedding(config.Embedding))
	if len(config.Api) == 0 {
		colorlog.Warning("API list is empty")
	}
	if len(config.Model) == 0 {
		colorlog.Warning("Model list is empty")
	}
	return nil
}

type SortExtraModel []ExtraModel

func (x SortExtraModel) Len() int           { return len(x) }
func (x SortExtraModel) Less(i, j int) bool { return x[i].Name < x[j].Name }
func (x SortExtraModel) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

type SortEmbedding []Embedding

func (x SortEmbedding) Len() int           { return len(x) }
func (x SortEmbedding) Less(i, j int) bool { return x[i].Name < x[j].Name }
func (x SortEmbedding) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func GROUPID() int64 {
	lock.RLock()
	defer lock.RUnlock()
	return config.GroupID
}

func GUIDE() string {
	lock.RLock()
	defer lock.RUnlock()
	return config.Guide
}

func API() []Api {
	lock.RLock()
	defer lock.RUnlock()
	return config.Api
}

func ExtraModelWithGroup(group string) []ExtraModel {
	lock.RLock()
	defer lock.RUnlock()
	return extraModel[group]
}

func ExtraModelGroup() []string {
	lock.RLock()
	defer lock.RUnlock()
	return extraModelAllGroup
}

func Process() []ControlProcess {
	lock.RLock()
	defer lock.RUnlock()
	return config.ControlProcess
}

func PreProcess() []ControlPreProcess {
	lock.RLock()
	defer lock.RUnlock()
	return config.ControlPreProcess
}

func Name2Process(name string) ControlProcess {
	p := Process()
	for _, v := range p {
		if v.Name == name {
			return v
		}
	}
	return p[0]
}

func MODELS() []Model {
	lock.RLock()
	defer lock.RUnlock()
	return config.Model
}

func MODELFILETONAME(file string) string {
	if file == "" {
		return ""
	}
	lock.RLock()
	defer lock.RUnlock()
	for _, m := range config.Model {
		if m.File == file {
			return m.Name
		}
	}
	return config.Model[0].Name
}

func GroupIndex2ExtraModels(groupIndex int) []ExtraModel {
	s := ExtraModelGroup()
	if groupIndex < 0 {
		return ExtraModelWithGroup(s[0])
	} else if groupIndex+1 > len(s) {
		return ExtraModelWithGroup(s[len(s)-1])
	} else {
		return ExtraModelWithGroup(s[groupIndex])
	}
}

func GroupIndex2GroupName(groupIndex int) string {
	s := ExtraModelGroup()
	if groupIndex < 0 {
		return s[0]
	} else if groupIndex+1 > len(s) {
		return s[len(s)-1]
	} else {
		return s[groupIndex]
	}
}

func Index2ExtraModel(GroupIndex, LoraIndex int) ExtraModel {
	s := GroupIndex2ExtraModels(GroupIndex)
	if LoraIndex < 0 {
		return s[0]
	} else if LoraIndex+1 > len(s) {
		return s[len(s)-1]
	} else {
		return s[LoraIndex]
	}
}

func EMBEDDING() []Embedding {
	lock.RLock()
	defer lock.RUnlock()
	return config.Embedding
}

func Name2Model(name string) Model {
	models := MODELS()
	for _, v := range models {
		if v.Name == name {
			return v
		}
	}
	return models[0]
}

func GetExtraModelWithGroup(group, name string) (ExtraModel, error) {
	loras := ExtraModelWithGroup(group)
	for _, v := range loras {
		if v.Name == name {
			return v, nil
		}
	}
	return ExtraModel{}, errors.New("not find lora")
}

func GetExtraModel(name string) (ExtraModel, error) {
	extraModels := ALLExtraModel()
	for _, v := range extraModels {
		if v.Name == name {
			return v, nil
		}
	}
	return ExtraModel{}, errors.New("not find lora")
}

func ALLMODELS() []Model {
	lock.RLock()
	defer lock.RUnlock()
	return config.Model
}

func ALLExtraModel() []ExtraModel {
	lock.RLock()
	defer lock.RUnlock()
	return config.ExtraModel
}

func ALLEmbedding() []Embedding {
	lock.RLock()
	defer lock.RUnlock()
	return config.Embedding
}

func GROUP() string {
	lock.RLock()
	defer lock.RUnlock()
	return config.Group
}

func checkDuplicates(list []Api) error {
	tmpMap := make(map[string]int)
	for k, v := range list {
		if _, ok := tmpMap[v.Url]; ok {
			return fmt.Errorf("%v is Duplicates", v)
		} else {
			tmpMap[v.Url] = k
		}
	}
	return nil
}

func checkDuplicatesExtraModel(list []ExtraModel) error {
	tmpMap := make(map[string]int)
	for k, v := range list {
		if _, ok := tmpMap[v.Name]; ok {
			return fmt.Errorf("%v is Duplicates", v)
		} else {
			tmpMap[v.Name] = k
		}
	}
	return nil
}
