package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/db"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"github.com/panjf2000/ants/v2"
	"github.com/zijiren233/go-colorlog"
)

type Resoult struct {
	Err     error
	Resoult [][]byte
}

func (cfg *config) Draw(ctx context.Context, free bool) <-chan *Resoult {
	cfg.ctx = ctx
	resoult := make(chan *Resoult, 1)
	cfg.resoultChan = resoult
	if free {
		cfg.a.freeWaitGloup <- cfg.draw
	} else {
		cfg.a.waitGloup <- cfg.draw
	}
	return resoult
}

func (cfg *config) Status() Status {
	if cfg.api.api == nil {
		cfg.api.status.Progress = 0.01
		return cfg.api.status
	}
	ctx, cancle := context.WithTimeout(cfg.ctx, time.Second*3)
	defer cancle()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.api.api.GenerateApi("progress?skip_current_image=true"), nil)
	if err != nil {
		return cfg.api.status
	}
	req.SetBasicAuth(cfg.api.api.Username, cfg.api.api.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cfg.api.status
	}
	defer resp.Body.Close()
	tmp := Status{}
	if json.NewDecoder(resp.Body).Decode(&tmp) != nil {
		return cfg.api.status
	}
	if tmp.Progress > cfg.api.status.Progress {
		if tmp.Progress < 0.05 {
			cfg.api.status.Progress = 0.05
		} else {
			cfg.api.status.Progress = utils.Round(tmp.Progress, 3)
		}
		cfg.api.status.EtaRelative = utils.Round(tmp.EtaRelative, 3)
	}
	return cfg.api.status
}

func (cfg *superResolutionCfg) SuperResolution(ctx context.Context, free bool) <-chan *Resoult {
	cfg.ctx = ctx
	resoult := make(chan *Resoult, 1)
	cfg.resoultChan = resoult
	if free {
		cfg.a.freeWaitGloup <- cfg.superResolution
	} else {
		cfg.a.waitGloup <- cfg.superResolution
	}
	return resoult
}

func (cfg *ctrlPhotoCfg) CtrlPhoto(ctx context.Context, free bool) <-chan *Resoult {
	cfg.ctx = ctx
	resoult := make(chan *Resoult, 1)
	cfg.resoultChan = resoult
	if free {
		cfg.a.freeWaitGloup <- cfg.ctrlPhoto
	} else {
		cfg.a.waitGloup <- cfg.ctrlPhoto
	}
	return resoult
}

func (cfg *interrogateCfg) Interrogate(ctx context.Context) <-chan *InterrogateResoult {
	cfg.ctx = ctx
	resoult := make(chan *InterrogateResoult, 1)
	cfg.resoultChan = resoult
	ants.Submit(cfg.interrogate)
	return resoult
}

func (cfg *config) draw() {
	var data [][]byte
	var err error
	defer func() {
		if err == nil && len(data) == 0 {
			err = errors.New("Resoult is nil")
		}
		cfg.resoultChan <- &Resoult{Err: err, Resoult: data}
	}()
	body, err := json.Marshal(cfg.cfg)
	if err != nil {
		return
	}
	api, done := cfg.a.next(cfg.cfg.Model)
	defer done()
	colorlog.Debug("get api: ", api.Url)
	cfg.api = struct {
		api    *apiUrl
		status Status
	}{
		api:    api,
		status: Status{Progress: 0.05},
	}
	err = api.ChangeOption(cfg.cfg)
	if err != nil {
		return
	}
	cfg.api.status.Progress = 0.1
	var req *http.Request
	if cfg.cfg.InitImages == nil {
		req, err = http.NewRequestWithContext(cfg.ctx, http.MethodPost, api.GenerateApi("txt2img"), bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(cfg.ctx, http.MethodPost, api.GenerateApi("img2img"), bytes.NewReader(body))
	}
	if err != nil {
		return
	}
	req.SetBasicAuth(api.Username, api.Password)
	resp, err := http.DefaultClient.Do(req)
	done()
	if err != nil {
		req, e := http.NewRequest(http.MethodPost, api.GenerateApi("interrupt"), nil)
		if e != nil {
			return
		}
		req.SetBasicAuth(api.Username, api.Password)
		http.DefaultClient.Do(req)
		return
	}
	defer resp.Body.Close()
	data, err = parseData(resp.Body)
}

func (cfg *superResolutionCfg) superResolution() {
	var err error
	var data [][]byte
	defer func() {
		if err == nil && len(data) == 0 {
			err = errors.New("Resoult is nil")
		}
		cfg.resoultChan <- &Resoult{Err: err, Resoult: data}
	}()
	body, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	api, done := cfg.a.next("")
	defer done()
	colorlog.Debug("get api: ", api.Url)
	err = api.ChangeOption(nil)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(cfg.ctx, http.MethodPost, api.GenerateApi("extra-batch-images"), bytes.NewReader(body))
	if err != nil {
		return
	}
	req.SetBasicAuth(api.Username, api.Password)
	resp, err := http.DefaultClient.Do(req)
	done()
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err = parseData(resp.Body)
}

func (cfg *ctrlPhotoCfg) ctrlPhoto() {
	var err error
	var data [][]byte
	defer func() {
		if err == nil && len(data) == 0 {
			err = errors.New("Resoult is nil")
		}
		cfg.resoultChan <- &Resoult{Err: err, Resoult: data}
	}()
	body, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	api, done := cfg.a.next("")
	defer done()
	colorlog.Debug("get api: ", api.Url)
	err = api.ChangeOption(nil)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(cfg.ctx, http.MethodPost, api.GenerateApi("/controlnet/detect"), bytes.NewReader(body))
	if err != nil {
		return
	}
	req.SetBasicAuth(api.Username, api.Password)
	resp, err := http.DefaultClient.Do(req)
	done()
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err = parseData(resp.Body)
}

func (cfg *interrogateCfg) interrogate() {
	var err error
	var data = &interrogateData{}
	defer func() {
		if err == nil && len(data.Caption) == 0 {
			err = errors.New("Resoult is nil")
		}
		cfg.resoultChan <- &InterrogateResoult{Err: err, Resoult: data.Caption}
	}()
	body, err := json.Marshal(cfg)
	if err != nil {
		return
	}
	api, done := cfg.a.next("")
	defer done()
	colorlog.Debug("get api: ", api.Url)
	err = api.ChangeOption(nil)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(cfg.ctx, http.MethodPost, api.GenerateApi("interrogate"), bytes.NewReader(body))
	if err != nil {
		return
	}
	req.SetBasicAuth(api.Username, api.Password)
	resp, err := http.DefaultClient.Do(req)
	done()
	if err != nil {
		req, e := http.NewRequest(http.MethodPost, api.GenerateApi("interrupt"), nil)
		if e != nil {
			return
		}
		req.SetBasicAuth(api.Username, api.Password)
		http.DefaultClient.Do(req)
		return
	}
	defer resp.Body.Close()
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(d, data)
}

type interrogateData struct {
	Caption string `json:"caption"`
}

type data struct {
	Images []string `json:"images"`
	Detail []struct {
		Loc  []interface{} `json:"loc"`
		Msg  string        `json:"msg"`
		Type string        `json:"type"`
	} `json:"detail"`
}

func parseData(rowData io.Reader) ([][]byte, error) {
	d, err := io.ReadAll(rowData)
	if err != nil {
		return nil, err
	}
	datas := [][]byte{}
	data := new(data)
	err = json.Unmarshal(d, data)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal err: %v, resp: %s", err, string(d))
	}
	if len(data.Detail) != 0 {
		msg := bytes.NewBuffer(nil)
		for _, v := range data.Detail {
			msg.WriteString(v.Msg)
			msg.WriteRune('\n')
		}
		return nil, errors.New(msg.String())
	}
	for _, v := range data.Images {
		data, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			continue
		} else {
			datas = append(datas, data)
		}
	}
	return datas, nil
}

func (api *API) back() {
	defer api.drawPool.Release()
	for {
		select {
		case run := <-api.waitGloup:
			api.drawPool.Submit(run)
		default:
			if api.drawPool.Free() > 0 {
				select {
				case run := <-api.freeWaitGloup:
					api.drawPool.Submit(run)
				default:
					time.Sleep(time.Millisecond * 10)
				}
			} else {
				time.Sleep(time.Millisecond * 10)
			}
		}
	}
}

func (api *API) DrawPoolCap() int {
	return api.drawPool.Cap()
}

func (api *API) DrawFree() int {
	return api.drawPool.Free()
}

func (api *API) DrawWait() int {
	return len(api.waitGloup) + len(api.freeWaitGloup)
}

type photoExif struct {
	Tag      string  `json:"-"`
	Width    int     `json:"-"`
	Height   int     `json:"-"`
	Steps    int     `json:"steps"`
	Sampler  string  `json:"sampler"`
	Seed     uint32  `json:"seed"`
	Strength float64 `json:"strength"`
	Noise    float64 `json:"noise"`
	Model    string  `json:"-"`
	Scale    float64 `json:"scale"`
	Uc       string  `json:"uc"`
}

var (
	reParseConfig, _  = regexp.Compile(`\s*([\w ]+):\s*("(?:\\"[^,]|\\"|\\|[^\"])+"|[^,]*)(?:,|$)`)
	reParseConfigS, _ = regexp.Compile(`^(?:\w+|\s*([\w ]+):\s*("(?:\\"[^,]|\\"|\\|[^\"])+"|[^,]*)(?:,|$)){3,}$`)
	negative, _       = regexp.Compile(`^Negative prompt: (.*)`)
)

func parseConfig(str []string) map[string]string {
	kvm := make(map[string]string)
	for _, v := range str {
		ret := reParseConfig.FindAllStringSubmatch(v, -1)
		for _, v := range ret {
			if len(v) == 3 {
				kvm[v[1]] = v[2]
			}
		}
	}
	return kvm
}

func init() {
	os.RemoveAll(path.Join(os.TempDir(), "exif-tmp"))
}

func GetImgCfg(photo []byte) (*db.Config, error) {
	if len(photo) == 0 {
		return nil, errors.New("photo len is nil")
	}
	dPath := path.Join(os.TempDir(), "exif-tmp")
	err := os.MkdirAll(dPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	fPath := path.Join(dPath, utils.Md5(photo))
	file, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(photo)
	file.Close()
	if err != nil {
		return nil, err
	}
	defer os.Remove(fPath)
	exiftool, err := utils.Exif()
	if err != nil {
		return nil, err
	}
	fm := exiftool.ExtractMetadata(fPath)
	if fm[0].Err != nil {
		return nil, fm[0].Err
	}
	exif := &photoExif{}
	if parameters, err := fm[0].GetString("Parameters"); err == nil {
		data := strings.Split(parameters, "\n")
		l := len(data)
		var lastLine = data[l-1]
		if !reParseConfigS.MatchString(lastLine) {
			lastLine = ""
		} else {
			data = data[:l-1]
		}
		var done_tag = false
		for _, v := range data {
			if s2 := negative.FindStringSubmatch(v); len(s2) == 2 {
				done_tag = true
				v = s2[1]
			}
			if done_tag {
				if exif.Uc != "" {
					v = "\n, " + v
				}
				exif.Uc = fmt.Sprint(exif.Uc, v)
			} else {
				if exif.Tag != "" {
					v = "\n, " + v
				}
				exif.Tag = fmt.Sprint(exif.Tag, v)
			}
		}
		dataMap := parseConfig([]string{lastLine})
		if f, err := strconv.ParseFloat(dataMap["Denoising strength"], 64); err == nil {
			exif.Strength = f
		}
		if f, err := strconv.ParseFloat(dataMap["Variation seed strength"], 64); err == nil {
			exif.Noise = f
		}
		if f, err := strconv.ParseFloat(dataMap["CFG scale"], 64); err == nil {
			exif.Scale = f
		}
		exif.Model = dataMap["Model"]
		exif.Sampler = dataMap["Sampler"]
		if i, err := strconv.ParseInt(dataMap["Steps"], 10, 64); err == nil {
			exif.Steps = int(i)
		}
		if u, err := strconv.ParseUint(dataMap["Seed"], 10, 32); err == nil {
			exif.Seed = uint32(u)
		}
	} else {
		var cfg string
		cfg, err = fm[0].GetString("Comment")
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(cfg), exif)
		if err != nil {
			return nil, err
		}
		exif.Tag, err = fm[0].GetString("Description")
		if err != nil {
			return nil, err
		}
	}
	exif.Width, exif.Height, _ = utils.GetPhotoSize(photo)
	cfg := &db.Config{Seed: exif.Seed, Steps: exif.Steps, Model: exif.Model, CfgScale: int(exif.Scale), Num: 0, Tag: exif.Tag, Mode: exif.Sampler, Width: exif.Width, Height: exif.Height, Uc: exif.Uc}
	return cfg, nil
}
