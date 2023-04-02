package hf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/dlclark/regexp2"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

const LinkModeWs int = 0
const LinkModePoll int = 1

type Client interface {
	SetWSReadLimit(limit int64)
	SetTimeout(timeout time.Duration)
	SetUrl(url string)
	SendWsMsg(fnIndex int, data interface{}) (interface{}, error)
	SendPollMsg(fnIndex int, data interface{}) (interface{}, error)
	GetData(id int64) (interface{}, error)
	SetData(id int64, elem interface{}) bool
	GetComponent(id int64) (*Component, error)
	GetAppStructure() map[string]structure
	GetInfo() error
	InvokeFunc(id int64, linkMode int) error
}

type message struct {
	Action      string      `json:"action,omitempty"`
	FnIndex     int         `json:"fn_index"`
	SessionHash string      `json:"session_hash,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	Hash        string      `json:"hash,omitempty"`
}

type response struct {
	Msg    string `json:"msg,omitempty"`
	Output struct {
		Data interface{} `json:"data"`
	} `json:"output,omitempty"`
	Success bool   `json:"success,omitempty"`
	Hash    string `json:"hash,omitempty"`
	Data    struct {
		Data interface{} `json:"data"`
	} `json:"data,omitempty"`
	Status string `json:"status,omitempty"`
}

type structure struct {
	elem     *Component
	children map[string]structure
}

type client struct {
	wsLimit       int64
	url           string
	info          *SpaceConf
	ComponentList map[int64]*Component
	appStructure  map[string]structure
	timeout       time.Duration
	readLimit     int
}

func GetClient() (Client, error) {
	var c Client = &client{timeout: 3 * time.Minute}
	return c, nil
}

// GetClientWithUrl 获取 Client 时一并初始化 url
func GetClientWithUrl(url string, getInfo bool) (Client, error) {
	c, err := GetClient()
	if err != nil {
		return nil, err
	}
	c.SetUrl(url)
	if getInfo {
		err = c.GetInfo()
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *client) SetWSReadLimit(limit int64) {
	c.wsLimit = limit
}

func (c *client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func GetWsUrl(url string) string {
	arr := strings.Split(url, "/")
	length := len(arr)
	return strings.ToLower("wss://" + arr[length-2] + "-" + arr[length-1] + ".hf.space/queue/join")
}

func GetPollUrl(url string) string {
	arr := strings.Split(url, "/")
	length := len(arr)
	return strings.ToLower("https://" + arr[length-2] + "-" + arr[length-1] + ".hf.space/api/queue")
}

func GetSpaceUrl(url string) string {
	arr := strings.Split(url, "/")
	length := len(arr)
	return strings.ToLower("https://" + arr[length-2] + "-" + arr[length-1] + ".hf.space/")
}

func (c *client) SetUrl(url string) {
	c.url = url
}

func GenerateSessionHash() (string, error) {
	h := md5.New()
	_, err := io.WriteString(h, time.Now().String())
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil))[:11], nil
}

func (c *client) GetData(id int64) (interface{}, error) {
	component, err := c.GetComponent(id)
	if err != nil {
		return nil, err
	}
	return component.Props.Value, nil
}

func (c *client) SetData(id int64, elem interface{}) bool {
	var ok bool
	if c.info == nil {
		return false
	}
	if _, ok = c.ComponentList[id]; !ok {
		return false
	}
	switch c.ComponentList[id].Props.Value.(type) {
	case string:
		elem, ok = elem.(string)
		if !ok {
			return false
		}
	case int:
		elem, ok = elem.(int)
		if !ok {
			return false
		}
	case int64:
		elem, ok = elem.(int64)
		if !ok {
			return false
		}
	case float64:
		elem, ok = elem.(float64)
		if !ok {
			return false
		}
	case float32:
		elem, ok = elem.(float32)
		if !ok {
			return false
		}
	case bool:
		elem, ok = elem.(bool)
		if !ok {
			return false
		}
	default:
	}
	c.ComponentList[id].Props.Value = elem
	return true
}

func (c *client) GetComponent(id int64) (*Component, error) {
	if c.info == nil {
		return nil, errors.New("未初始化相关数据")
	}
	if id > int64(len(c.info.Components)) {
		return nil, errors.New("请求数据超出范围")
	}
	return c.ComponentList[id], nil
}

func (c *client) GetAppStructure() map[string]structure {
	return c.appStructure
}

func (c *client) init() (err error) {
	c.initComponent()
	err = c.initFunc()
	if err != nil {
		return err
	}
	var initStructure []structure
	initStructure, err = c.initStructure(&c.info.Layout)
	if err != nil {
		return err
	}
	structureMap := make(map[string]structure)
	for _, v := range initStructure {
		structureMap[v.elem.Props.Label] = v
	}
	c.appStructure = structureMap
	return
}

func (c *client) initComponent() {
	c.ComponentList = make(map[int64]*Component)
	for i, v := range c.info.Components {
		c.ComponentList[v.ID] = &c.info.Components[i]
	}
}

func (c *client) initFunc() (err error) {
	for i, elem := range c.info.Dependencies {
		for _, index := range elem.Targets {
			if elem.Js == "" {
				c.ComponentList[index].Func = Func{
					Dependency: &c.info.Dependencies[i],
					FnIndex:    i,
				}
			}
		}
	}
	return
}

func (c *client) initStructure(layout *Layout) (s []structure, err error) {
	var arr []structure
	var res []structure
	var component *Component
	for i, _ := range layout.Children {
		arr, err = c.initStructure(&layout.Children[i])
		if err != nil {
			return nil, err
		}
		res = append(res, arr...)
	}
	component = c.ComponentList[layout.ID]
	if component != nil && (component.Props.Label != "" || (component.Type == "button" && component.Props.Value != nil)) {
		mp := make(map[string]structure)
		for _, comp := range res {
			mp[comp.elem.Props.Label] = comp
		}
		if component.Props.Label == "" {
			component.Props.Label = component.Props.Value.(string)
		}
		//把边上的空格去了，太坑爹了
		component.Props.Label = strings.TrimSpace(component.Props.Label)
		return []structure{
			{
				elem:     component,
				children: mp,
			},
		}, err
	} else {
		s = res
		return
	}
}

func (c *client) GetInfo() (err error) {
	if c.url == "" {
		err = errors.New("未配置 url")
		return
	}
	var get *http.Response
	get, err = http.Get(GetSpaceUrl(c.url))
	if err != nil {
		logrus.Errorf("连接空间失败: %s", err.Error())
		return
	}
	defer get.Body.Close()
	var b []byte
	b, err = io.ReadAll(get.Body)
	if err != nil {
		logrus.Errorf("读取返回消息时失败: %s", err.Error())
		return
	}
	compile := regexp2.MustCompile("(?<=<script>window.gradio_config = ).*(?=;</script>)", 0)
	var match *regexp2.Match
	match, err = compile.FindStringMatch(string(b))
	if match == nil {
		if err == nil {
			err = errors.New("未找到对应字段")
		}
		logrus.Errorf("字段查找失败: %s", err.Error())
		return
	}
	var conf SpaceConf
	err = json.Unmarshal([]byte(match.String()), &conf)
	if err != nil {
		logrus.Errorf("数据反序列化失败: %s", err.Error())
		return
	}
	c.info = &conf
	err = c.init()
	if err != nil {
		return err
	}
	return
}

func (c *client) InvokeFunc(id int64, linkMode int) error {
	if c.info == nil {
		return errors.New("未初始化相关数据")
	}
	var data []interface{}
	var msg interface{}
	var err error
	f := c.ComponentList[id].Func
	if f.Dependency == nil {
		return errors.New("该组件没有可调用的函数")
	}
	fnIndex := f.FnIndex
	for _, index := range f.Dependency.Inputs {
		getData, err := c.GetData(index)
		if err != nil {
			return err
		}
		data = append(data, getData)
	}
	switch linkMode {
	case LinkModeWs:
		msg, err = c.SendWsMsg(fnIndex, data)
		if err != nil {
			return err
		}
	case LinkModePoll:
		msg, err = c.SendPollMsg(fnIndex, data)
		if err != nil {
			return err
		}
	default:
		msg, err = c.SendWsMsg(fnIndex, data)
		if err != nil {
			return err
		}
	}
	data, _ = msg.([]interface{})
	for i, res := range data {
		c.SetData(f.Dependency.Outputs[i], res)
	}
	return nil
}

func readFromWs(ws *websocket.Conn, resChan chan response) {
	var err error
	for {
		var res response
		err = ws.ReadJSON(&res)
		if err != nil {
			return
		}
		resChan <- res
	}
}

func (c *client) SendWsMsg(fnIndex int, data interface{}) (interface{}, error) {
	var err error
	var ws *websocket.Conn
	var hash string
	resChan := make(chan response)

	ws, _, err = websocket.DefaultDialer.Dial(GetWsUrl(c.url), nil)
	if err != nil {
		return nil, errors.New("WebSocket 失败！")
	}

	defer func() {
		ws = nil
	}()
	if data == nil {
		return nil, errors.New("data 字段为空！")
	}
	go readFromWs(ws, resChan)
	hash, err = GenerateSessionHash()
	if err != nil {
		return nil, err
	}
	for {
		select {
		case res := <-resChan:
			switch res.Msg {
			case "send_hash":
				err = ws.WriteJSON(message{
					FnIndex:     fnIndex,
					SessionHash: hash,
				})
				if err != nil {
					return nil, err
				}
			case "send_data":
				err = ws.WriteJSON(message{
					FnIndex:     fnIndex,
					SessionHash: hash,
					Data:        data,
				})
				if err != nil {
					return nil, err
				}
			case "process_completed":
				if res.Success {
					return res.Output.Data, nil
				}
				return nil, errors.New("结果获取失败")
			}
		case <-time.After(c.timeout):
			return nil, errors.New("连接超时")
		}
	}
}

func readFromPoll(url string, hash string, resChan chan response) {
	var err error
	var reqJson []byte
	var respJson []byte
	var post *http.Response
	var res response
	defer func() {
		if post.Body != nil {
			post.Body.Close()
		}
	}()
	for {
		reqJson, err = json.Marshal(message{
			Hash: hash,
		})
		if err != nil {
			continue
		}
		post, err = http.Post(url, "application/json", bytes.NewBuffer(reqJson))
		if err != nil {
			continue
		}
		respJson, err = io.ReadAll(post.Body)
		if err != nil {
			continue
		}
		err = json.Unmarshal(respJson, &res)
		if err != nil {
			continue
		}
		if res.Status == "COMPLETE" {
			resChan <- res
			return
		}
	}
}

func (c *client) SendPollMsg(fnIndex int, data interface{}) (interface{}, error) {
	if data == nil {
		return nil, errors.New("data 字段为空！")
	}
	var hash string
	var err error
	var reqJson []byte
	var respJson []byte
	var post *http.Response
	var resp response
	resChan := make(chan response)
	pollUrl := GetPollUrl(c.url)
	hash, err = GenerateSessionHash()
	reqJson, err = json.Marshal(message{
		Action:      "predict",
		FnIndex:     fnIndex,
		SessionHash: hash,
		Data:        data,
	})
	if err != nil {
		return nil, err
	}
	post, err = http.Post(pollUrl+"/push/", "application/json", bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	defer post.Body.Close()
	respJson, err = io.ReadAll(post.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(respJson, &resp)
	if err != nil {
		return nil, err
	}
	go readFromPoll(pollUrl+"/status/", resp.Hash, resChan)
	select {
	case res := <-resChan:
		return res.Data.Data, nil
	case <-time.After(c.timeout):
		return nil, errors.New("连接超时")
	}
}
