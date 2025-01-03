package main

import (
	"errors"
	"os"

	"io/ioutil"
	"regexp"

	"strings"

	"fmt"

	"path/filepath"

	"goconfenv/goconfigobj"
)

type CemConf struct {
	confItems    []*CemConfItem
	confObj      *goconfigobj.ConfigObj
	cemSvcs      *Cemsvcs
	personalVals map[string]string
}

type CemConfItem struct {
	name    string
	from    string
	to      string
	cemSvcs *Cemsvcs
	cemConf *CemConf
}

type CemConfRenderResult struct {
	Succ     bool
	ConfItem *CemConfItem
	Vals     map[string]string
	Msg      string
}

func NewCemConfItem(name, from, to string, cemSvcs *Cemsvcs, cemConf *CemConf) *CemConfItem {
	return &CemConfItem{
		name:    name,
		from:    from,
		to:      to,
		cemSvcs: cemSvcs,
		cemConf: cemConf,
	}
}

func NewCemConf(cemSvcs *Cemsvcs) *CemConf {
	return &CemConf{
		confItems:    make([]*CemConfItem, 0, 2),
		cemSvcs:      cemSvcs,
		personalVals: make(map[string]string),
	}
}

func (cc *CemConf) ParseConf(cemConfPath string) error {
	fd, err := os.Open(cemConfPath)
	if err != nil {
		return err
	}
	defer fd.Close()

	confObj := goconfigobj.NewConfigObj(fd)
	confSection := confObj.Section("config_file")
	if confSection == nil {
		return errors.New("cem conf no config_file section")
	}

	sections := confSection.AllSections()

	if len(sections) == 0 {
		return nil
	}

	for name, s := range sections {
		from := s.Value("from")
		to := s.Value("to")
		if from != "" && to != "" {
			item := NewCemConfItem(name, from, to, cc.cemSvcs, cc)
			cc.confItems = append(cc.confItems, item)
		}
	}

	//parse personal vals
	cemConfAbsPath, err := filepath.Abs(cemConfPath)
	if err != nil {
		return nil
	}
	personalConfPath := filepath.Dir(cemConfAbsPath) + string(os.PathSeparator) + "personal.conf"
	personalFd, err := os.Open(personalConfPath)
	if err != nil {
		//no mater
		return nil
	}
	personalConfObj := goconfigobj.NewConfigObj(personalFd)
	personalVals := personalConfObj.AllDatas()
	cc.personalVals = personalVals

	return nil
}

// render the variable
func (cc *CemConf) Render() []*CemConfRenderResult {
	if len(cc.confItems) == 0 {
		return make([]*CemConfRenderResult, 0)
	}

	res := make([]*CemConfRenderResult, 0, len(cc.confItems))
	for _, confItem := range cc.confItems {
		ccrr := confItem.render()
		res = append(res, ccrr)
	}

	return res
}

func (cci *CemConfItem) render() *CemConfRenderResult {
	ccrr := new(CemConfRenderResult)
	ccrr.ConfItem = cci
	ccrr.Succ = false
	vars, err := cci.getVars()
	if err != nil {
		ccrr.Msg = err.Error()
		return ccrr
	}

	if len(vars) == 0 {
		ccrr.Succ = true
		ccrr.Vals = make(map[string]string)
		return ccrr
	}
	vals, err := cci.cemSvcs.callCemSvcsRender(vars)
	if err != nil {
		ccrr.Msg = err.Error()
		return ccrr
	}

	var notFoundVars []string
	for i := 0; i < len(vars); i++ {
		_, ok := vals[vars[i]]
		if !ok {
			notFoundVars = append(notFoundVars, vars[i])
		}
	}
	if len(notFoundVars) > 0 {
		ccrr.Msg = fmt.Sprintf("Not found following vars from remote: %s. pls check.", strings.Join(notFoundVars, ","))
		return ccrr
	}

	//use personal conf value
	if len(cci.cemConf.personalVals) != 0 {

		for valName, val := range cci.cemConf.personalVals {
			vals[valName] = val
		}
	}

	content, err := cci.getFromContent()
	if err != nil {
		ccrr.Msg = err.Error()
		return ccrr
	}

	if len(vals) > 0 {
		for varName, val := range vals {
			content = strings.Replace(content, "@@"+varName+"@@", val, -1)
		}
	}

	err = cci.putToContent(content)

	if err != nil {
		ccrr.Msg = err.Error()
	} else {
		ccrr.Succ = true
		ccrr.Vals = vals
	}

	return ccrr
}

func (cci *CemConfItem) putToContent(content string) error {
	fd, err := os.Create(cci.to)
	if err != nil {
		return err
	}

	defer fd.Close()
	_, err = fd.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}

// get vars from from file
func (cci *CemConfItem) getVars() ([]string, error) {
	reg := regexp.MustCompile("@@([^@]+)@@")
	content, err := cci.getFromContent()
	if err != nil {
		return nil, err
	}

	regRes := reg.FindAllStringSubmatch(content, -1)
	if len(regRes) == 0 {
		return make([]string, 0, 0), nil
	}

	res := make([]string, 0, len(regRes))

	for _, strList := range regRes {
		if len(strList) != 2 {
			continue
		}

		res = append(res, strList[1])
	}

	return res, nil
}

func (cci *CemConfItem) getFromContent() (string, error) {
	fd, err := os.Open(cci.from)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func (ccrr CemConfRenderResult) Print() {

	if ccrr.Succ {
		fmt.Printf("****** %s ******\n", ccrr.ConfItem.name)
		for varName, val := range ccrr.Vals {
			fmt.Printf("-- %s: %s\n", varName, val)
		}
		fmt.Printf("****** Succ ******\n")
	} else {

		fmt.Fprintf(os.Stderr, "****** %s ******\n", ccrr.ConfItem.name)
		fmt.Fprintf(os.Stderr, "%s\n", ccrr.Msg)
		fmt.Fprintf(os.Stderr, "****** Failed Var Not Exists OR error ******\n")
	}
}
