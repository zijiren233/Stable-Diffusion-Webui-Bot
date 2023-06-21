package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/gconfig"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"github.com/zijiren233/go-colorlog"
)

var loadBalance = &api{
	apiList: &[]*apiUrl{},
	lock:    &sync.RWMutex{},
}

func GetWoker() (m *sync.Map, c chan *apiUrl, a []*apiUrl) {
	loadBalance.lock.RLock()
	defer loadBalance.lock.RUnlock()
	if loadBalance.working != nil {
		m = *loadBalance.working
	}
	if loadBalance.apiPool != nil {
		c = *loadBalance.apiPool
	}
	if loadBalance.apiList != nil {
		a = *loadBalance.apiList
	}
	return
}

type api struct {
	apiList *[]*apiUrl
	apiPool *chan *apiUrl
	working **sync.Map // api -> bool
	lock    *sync.RWMutex
}

var backendOnce = &sync.Once{}

func Load(apis []gconfig.Api) {
	loadAPI(apis)
	backendOnce.Do(func() {
		go back()
		go failover()
	})
}

type apiUrl struct {
	gconfig.Api
	Models       []Model
	CurrentModel string
	LoadedModels *sync.Map
}

func (api *apiUrl) ChangeOption(model string) error {
	option := map[string]interface{}{"add_version_to_infotext": false, "lora_add_hashes_to_infotext": false, "add_model_hash_to_info": false, "add_model_name_to_info": true, "deepbooru_use_spaces": true, "interrogate_clip_dict_limit": 0, "interrogate_return_ranks": true, "deepbooru_sort_alpha": false, "interrogate_deepbooru_score_threshold": 0.5, "interrogate_clip_min_length": 15, "interrogate_clip_max_length": 50, "live_previews_enable": false, "sd_vae_as_default": false, "sd_checkpoint_cache": 0, "sd_vae_checkpoint_cache": 0, "grid_save": false, "eta_noise_seed_delta": 31337, "eta_ancestral": 1, "samples_save": false, "enable_emphasis": true}
	if model != "" {
		m := gconfig.Name2Model(model)
		option["sd_model_checkpoint"] = m.File
		option["sd_vae"] = m.Vae + m.VaeExt
		option["CLIP_stop_at_last_layers"] = m.ClipSkip
	}
	b, err := json.Marshal(option)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, api.GenerateApi("options"), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.SetBasicAuth(api.Username, api.Password)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if string(body) == "null" {
		if model != "" {
			api.CurrentModel = model
			api.LoadedModels.Store(model, nil)
		}
		return nil
	}
	return fmt.Errorf("change option err: %s", string(body))
}

var getApiL = sync.Mutex{}

func next(tryModel string) (*apiUrl, func()) {
	getApiL.Lock()
	defer getApiL.Unlock()
	var (
		working           *sync.Map
		apiChan           chan *apiUrl
		currentApiChanPtr chan *apiUrl
		api               *apiUrl
		firstApi          *apiUrl
		ok                bool
		count             = 0
	)
	for {
		working, apiChan, _ = GetWoker()
		if len(apiChan) == 0 {
			continue
		}
		if currentApiChanPtr == nil {
			currentApiChanPtr = apiChan
		} else if currentApiChanPtr != apiChan {
			currentApiChanPtr = apiChan
			firstApi = nil
			count = 0
		}
		api, ok = <-apiChan
		if !ok {
			continue
		}
		if v, ok := working.Load(api.Url); ok && v.(bool) {
			if tryModel == "" || api.CurrentModel == tryModel || func() bool {
				if count == 0 {
					return false
				}
				_, ok := api.LoadedModels.Load(tryModel)
				return ok
			}() {
				break
			}
			if firstApi == nil {
				firstApi = api
				count = 0
			} else if api == firstApi {
				count++
				if count == 2 {
					break
				}
			}
		}
		apiChan <- api
	}
	down := new(sync.Once)
	return api, func() {
		down.Do(func() {
			apiChan <- api
		})
	}
}

func (api *apiUrl) GenerateApi(types string) string {
	if strings.HasPrefix(types, "/") {
		return fmt.Sprint(api.Url, types)
	}
	return fmt.Sprint(api.Url, "/sdapi/v1/", types)
}

func loadAPI(apis []gconfig.Api) {
	loadBalance.lock.Lock()
	defer loadBalance.lock.Unlock()
	working := &sync.Map{}
	apiPool := make(chan *apiUrl, len(apis))
	api := []*apiUrl{}
	for _, v := range apis {
		working.Store(v, false)
		apiU := apiUrl{Api: v, LoadedModels: &sync.Map{}}
		apiPool <- &apiU
		api = append(api, &apiU)
	}
	if SliceEqualBCE(api, *loadBalance.apiList) {
		return
	}
	loadBalance.apiList = &api
	loadBalance.apiPool = &apiPool
	loadBalance.working = &working
}

func SliceEqualBCE(a, b []*apiUrl) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v.Url != b[i].Url {
			return false
		}
	}

	return true
}

var errReturn = errors.New("detail return error")
var errModels = errors.New("not allow models")

type Model struct {
	ModelName string `json:"model_name"`
}

func failover() {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	wait := &sync.WaitGroup{}
	var workers uint32
	for range timer.C {
		working, _, apiList := GetWoker()
		allowModels := gconfig.MODELS()
		for _, v := range apiList {
			wait.Add(1)
			go func(api *apiUrl) {
				defer wait.Done()
				err := utils.Retry(2, false, 0, func() (bool, error) {
					ctx, cf := context.WithTimeout(context.Background(), time.Second*3)
					defer cf()
					req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.GenerateApi("sd-models"), nil)
					if err != nil {
						return false, err
					}
					req.SetBasicAuth(api.Username, api.Password)
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return true, err
					}
					defer resp.Body.Close()
					b, err := io.ReadAll(resp.Body)
					if err != nil {
						return false, err
					}
					if json.Unmarshal(b, &api.Models) != nil {
						return false, errReturn
					}
					if !ModelAllowed(api.Models, allowModels) {
						return false, errModels
					}
					return false, nil
				})
				if err != nil {
					colorlog.Errorf("api [%s] have some error: %v", api.Url, err)
					working.Store(api.Url, false)
					api.LoadedModels.Range(func(key, value any) bool {
						api.LoadedModels.Delete(key)
						return true
					})
				} else {
					if v, ok := working.Load(api); ok && !v.(bool) {
						api.CurrentModel = allowModels[0].Name
					}
					api.LoadedModels.Store(allowModels[0].Name, nil)
					working.Store(api.Url, true)
					atomic.AddUint32(&workers, 1)
				}
			}(v)
		}
		wait.Wait()
		if workers == 0 {
			workers = 1
		}
		drawPool.Tune(int(workers))
		workers = 0
		timer.Reset(time.Second * 10)
	}
}

func ModelAllowed(Model []Model, AllAllowModels []gconfig.Model) bool {
	if len(AllAllowModels) > len(Model) {
		return false
	}
	for _, v := range AllAllowModels {
		for k, model := range Model {
			if v.File == model.ModelName {
				break
			}
			if k == len(Model)-1 {
				return false
			}
		}
	}
	return true
}
