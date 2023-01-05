package aiTalk

import (
	"Lealra/config"
	"Lealra/myUtil"
	"Lealra/returnStruct"
	"encoding/json"
	"github.com/gorilla/websocket"
	"strconv"
	"strings"
	"time"
)

var MsgDistributeMap = make(map[string]chan string)
var BotMap = make(map[string]interface{})
var OpenAiMap = make(map[string]*OpenAiPersonal)
var numList = []string{"₀", "₁", "₂", "₃", "₄", "₅", "₆", "₇", "₈", "₉"}

func generatNumber(num int) string {
	var ret string
	for num > 0 {
		ret = numList[num%10] + ret
		num /= 10
	}
	return ret
}

func MsgHandler(ml []string, mjson returnStruct.Message, ws *websocket.Conn) bool {
	name := strconv.FormatInt(mjson.GroupID, 10) + strconv.FormatInt(mjson.UserID, 10)
	if len(ml) >= 2 {
		if _, ok := OpenAiMap[name]; !ok {
			OpenAiMap[name] = &OpenAiPersonal{}
		}
		switch ml[0] {
		case "/talk":
			msg, err := OpenAiMap[name].SendAndReceiveMsg(mjson.Message[6:])
			myUtil.SendGroupMessage(mjson.GroupID, msg)
			if err != nil {
				myUtil.ErrLog.Println("OpenAi talk 请求时出错！error:", err)
			}
			return true
		case "/preset":
			OpenAiMap[name].SetPreset(mjson.Message[8:])
			myUtil.SendGroupMessage(mjson.GroupID, "预设配置成功！")
		case "/edit":
			if len(ml) < 3 {
				return false
			}
			msg, err := OpenAiMap[name].EditMsg(strings.Join(ml[1:len(ml)-1], " "), ml[len(ml)-1])
			myUtil.SendGroupMessage(mjson.GroupID, msg)
			if err != nil {
				myUtil.ErrLog.Println("OpenAi edit 请求时出错！error:", err)
			}
			return true
		}
	}
	if MsgDistributeMap[name] != nil {
		if mjson.Message == "再见" {
			if MsgDistributeMap[name] != nil {
				myUtil.SendGroupMessage(mjson.GroupID, "再见~")
				myUtil.MsgLog.Println("aiTalk已关闭")
				close(MsgDistributeMap[name])
				MsgDistributeMap[name] = nil
			}
			return true
		} else {
			select {
			case MsgDistributeMap[name] <- mjson.Message:
				return true
			default:
				myUtil.MsgLog.Println("消息通道阻塞，丢弃消息：" + mjson.Message)
				return true
			}
		}
	}
	if len(ml) == 2 && ml[0] == "[CQ:at,qq="+strconv.FormatInt(config.Settings.BotName.Id, 10)+"]" {
		if ml[1] == "聊天" {
			msg := "请选择人格，10s后自动失效："
			for i, v := range CharList {
				msg += "\n" + generatNumber(i+1) + v.BotName
			}
			myUtil.SendGroupMessage(mjson.GroupID, msg)
			MsgDistributeMap[name] = make(chan string)
			var num int
			var err error
			select {
			case msg = <-MsgDistributeMap[name]:
				num, err = strconv.Atoi(msg)
				if num > len(CharList) || num <= 0 || err != nil {
					close(MsgDistributeMap[name])
					MsgDistributeMap[name] = nil
					myUtil.SendGroupMessage(mjson.GroupID, "请输入正确的编号！")
					return false
				}
			case <-time.After(10 * time.Second):
				close(MsgDistributeMap[name])
				MsgDistributeMap[name] = nil
				myUtil.SendGroupMessage(mjson.GroupID, "[CQ:at,qq="+strconv.FormatInt(mjson.UserID, 10)+"] 发起的会话已失效！")
				return false
			}
			go EstablishCoversation(ws, num-1, mjson.UserID, mjson.GroupID)
			return true
		}
	}
	if ok, _ := myUtil.IsInArray(config.Settings.Auth.Admin, mjson.UserID); ok && mjson.Message == "刷新人格列表" {
		if GetCharacterList() {
			myUtil.SendGroupMessage(mjson.GroupID, "人格列表已刷新！")
		} else {
			myUtil.SendGroupMessage(mjson.GroupID, "人格列表刷新时出现错误！请检查配置文件格式是否正确")
		}
		return true
	}
	return false
}

func EstablishCoversation(ws *websocket.Conn, botNum int, uid int64, groupId int64) {
	var chat *AiChat
	botId := CharList[botNum].BotId
	name := strconv.FormatInt(groupId, 10) + strconv.FormatInt(uid, 10)
	if BotMap[strconv.FormatInt(uid, 10)+botId] == nil {
		BotMap[strconv.FormatInt(uid, 10)+botId] = &AiChat{}
		chat = BotMap[strconv.FormatInt(uid, 10)+botId].(*AiChat)
		bot, err := chat.SetBot(botId)
		if !bot {
			myUtil.ErrLog.Println("群组", groupId, "的", uid, "建立 aiTalk 初始化失败,error:", err)
			myUtil.SendGroupMessage(groupId, "聊天建立失败，请稍后重试")
			close(MsgDistributeMap[name])
			MsgDistributeMap[name] = nil
			return
		}
		if len(chat.CreatedInfo.Messages) > 0 {
			myUtil.SendGroupMessage(groupId, chat.CreatedInfo.Messages[0].Text)
		} else {
			myUtil.SendGroupMessage(groupId, "该人格暂无问候语，请直接开始对话")
		}
		BotMap[strconv.FormatInt(uid, 10)+botId] = chat
	} else {
		chat = BotMap[strconv.FormatInt(uid, 10)+botId].(*AiChat)
		chat.Renew()
		len := len(chat.LastReply.Replies)
		if len != 0 {
			myUtil.SendGroupMessage(groupId, "请继续说吧，我们上次聊到哪里了？\n("+chat.LastReply.Replies[chat.LastReply.Index%len].Text+")")
		} else {
			BotMap[strconv.FormatInt(uid, 10)+botId] = nil
			myUtil.SendGroupMessage(groupId, "会话连接出现了意外问题！请尝试重新连接")
		}

	}
	for {
		select {
		case text, ok := <-MsgDistributeMap[name]:
			if ok {
				switch text {
				case "刷新":
					bot, err := chat.SetBot(botId)
					if !bot {
						myUtil.ErrLog.Println("群组", groupId, "的", uid, "刷新 aiTalk 失败,error:", err)
					} else {
						if len(chat.CreatedInfo.Messages) > 0 {
							myUtil.SendGroupMessage(groupId, chat.CreatedInfo.Messages[0].Text)
						} else {
							myUtil.SendGroupMessage(groupId, "该人格暂无问候语，请直接开始对话")
						}
					}
				case "重连":
					renew, err := chat.Renew()
					if renew {
						BotMap[strconv.FormatInt(uid, 10)+botId] = chat
						myUtil.SendGroupMessage(groupId, "重连成功！")
					} else {
						myUtil.ErrLog.Println("群组", groupId, "的", uid, "重连 aiTalk 失败,error:", err)
						myUtil.SendGroupMessage(groupId, "重连失败！请稍后重试，仍存在问题请告知开发者")
					}
				case "换一句":
					myUtil.SendGroupMessage(groupId, chat.GetAnotherMsg())
				default:
					_, msg, err := chat.SendMag(text)
					if err != nil {
						myUtil.ErrLog.Println("收取aiTalk回复失败,error:", err)
						msg = "哎呀，出错了！"
					}
					v := returnStruct.SendMsg{Action: "send_msg", Param: returnStruct.Params{Message: CharList[botNum].BotName + "_says," + msg, GroupID: groupId}}
					o, _ := json.Marshal(v)
					myUtil.WsLock.Lock()
					err = ws.WriteMessage(returnStruct.MsgType, o)
					myUtil.WsLock.Unlock()
					if err != nil {
						myUtil.ErrLog.Println("发送aiTalk回复失败,error:", err)
					}
				}
			} else {
				myUtil.MsgLog.Println("aiTalk被关闭，线程退出")
				return
			}
		case <-time.After(360 * time.Second):
			if MsgDistributeMap[name] != nil {
				close(MsgDistributeMap[name])
				MsgDistributeMap[name] = nil
			}
			return
		}
	}
}
