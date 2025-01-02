package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type Cemsvcs struct {
	svcsHost string
	env      string
}

type cemSvcsResp struct {
	Code    int
	Msg     string
	Data    interface{}
	Success bool
}

type cemSvcsRenderParam struct {
	Project string   `json:"project"`
	Env     string   `json:"env"`
	Vars    []string `json:"vars"`
}

func NewCemsvcs(svcsHost, env string) *Cemsvcs {
	return &Cemsvcs{
		svcsHost: svcsHost,
		env:      env,
	}
}

// render the vars
func (cs *Cemsvcs) Render(vars []string) (map[string]string, error) {
	vals, err := cs.callCemSvcsRender(vars)

	return vals, err
}

func (cs *Cemsvcs) callCemSvcsRender(vars []string) (map[string]string, error) {
	postParam := cemSvcsRenderParam{
		Project: *projectName,
		Env:     cs.env,
		Vars:    vars,
	}

	jsonBytes, err := json.Marshal(postParam)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://"+cs.svcsHost, "application/json", bytes.NewReader(jsonBytes))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := cemSvcsResp{}
	err = json.Unmarshal(respData, &r)

	if err != nil {
		return nil, err
	}

	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}

	respDataMap, ok := r.Data.(map[string]interface{})
	if !ok {
		return nil, errors.New(r.Msg)
	}

	vals := make(map[string]string)

	for k, v := range respDataMap {
		vStr, ok := v.(string)
		if !ok {
			return nil, errors.New("cem svcs return val not string")
		}
		vals[k] = vStr
	}

	return vals, nil
}
