package qqGroupLink

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var CQCodeTypes = map[string]string{
	"face":    "表情",
	"record":  "语音",
	"at":      "@某人",
	"share":   "链接分享",
	"music":   "音乐分享",
	"image":   "图片",
	"reply":   "回复",
	"redbag":  "红包",
	"forward": "合并转发",
	"xml":     "XML消息",
	"json":    "json消息",
}

type User struct {
	Nickname string `json:"nickname"`
}

type UniversalMessage struct {
	Message     string `json:"message"`
	GameRawText string
	MessageType string `json:"message_type"`
}

type MetaPost struct {
	Time          int64  `json:"time"`
	PostType      string `json:"post_type"`
	SelfID        int    `json:"self_id"`
	MetaEventType string `json:"meta_event_type"`
}

func ParseMetaPost(data []byte) (MetaPost, error) {
	post := MetaPost{}
	err := json.Unmarshal(data, &post)
	return post, err
}

type PrivateMessage struct {
	UniversalMessage
	MetaPost
	UserId int64 `json:"user_id"`
	Sender User  `json:"sender"`
}

type GroupMessage struct {
	PrivateMessage
	GroupID int64 `json:"group_id"`
}

type QMessage struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	// struct{
	// 		UserID string `json:"user_id"`
	// 		Message string `json:"message"`
	// }
	Echo string `json:"echo"`
}

//func (msg UniversalMessage) GetMessage() string {
//	return msg.Message
//}
//
//func (msg UniversalMessage) GetUser() int64 {
//	return -1
//}
//
//func (msg PrivateMessage) GetUser() int64 {
//	return msg.UserId
//}

const (
	IDUniversalMessage = iota
	IDPrivateMessage
	IDGroupMessage
)

func (msg UniversalMessage) ID() byte {
	return IDUniversalMessage
}

func (msg PrivateMessage) ID() byte {
	return IDPrivateMessage
}
func (msg GroupMessage) ID() byte {
	return IDGroupMessage
}

type IMessage interface {
	ID() byte
	Unmarshal(data []byte) (IMessage, error)
}

func GetMessageData(data []byte) (IMessage, error) {
	msg := map[string]interface{}{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	msgType := msg["message_type"].(string)
	// fmt.Println(msgType)
	switch msgType {
	case "private":
		return PrivateMessage{}.Unmarshal(data)
	case "group":
		return GroupMessage{}.Unmarshal(data)
	default:
		return UniversalMessage{}.Unmarshal(data)
	}
}

func (msg UniversalMessage) Unmarshal(data []byte) (IMessage, error) {
	err := json.Unmarshal(data, &msg)
	return msg, err
}
func (msg PrivateMessage) Unmarshal(data []byte) (IMessage, error) {
	err := json.Unmarshal(data, &msg)
	return msg, err
}

func (msg GroupMessage) Unmarshal(data []byte) (IMessage, error) {
	err := json.Unmarshal(data, &msg)
	return msg, err
}

//
//// GetSource 返回当前信息的来源. source为在group_id_list中定义的群昵称. 如果没有定义 则以群号代替. 若为私聊消息, 则为空值.
//func (msg UniversalMessage) GetSource() string {
//	return ""
//}
//
//func (msg GroupMessage) GetSource() string {
//	for _, g := range msg.groupInfo {
//		if msg.GroupID == g.GID {
//			return g.GName
//		}
//	}
//	return strconv.FormatInt(msg.GroupID, 10)
//}
//
//// GetRawTextFromCQMessage 将图片等CQ码转为文字.

func GetRawTextFromCQMessage(msg string) string {
	for k, v := range CQCodeTypes {
		format := fmt.Sprintf(`\[CQ:%s.*?\]`, k)
		rule := regexp.MustCompile(format)
		msg = rule.ReplaceAllString(msg, fmt.Sprintf("[%s]", v))
	}
	return msg
}
