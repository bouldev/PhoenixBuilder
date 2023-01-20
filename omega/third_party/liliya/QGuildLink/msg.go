package QGuildLink

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// 消息发送者数据结构
type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
	TinyID   string `json:"tiny_id"`
}

// 慢速模式数据结构
type SlowModeInfo struct {
	SlowModeKey    int32  `json:"slow_mode_key"`
	SlowModeText   string `json:"slow_mode_text"`
	SpeakFrequency int32  `json:"speak_frequency"`
	SlowModeCircle int32  `json:"slow_mode_circle"`
}

// 权限组信息数据结构
type RoleInfo struct {
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
}

// API请求数据结构
type CQRequest struct {
	Action string `json:"action"`
	Params any    `json:"params"`
	Echo   string `json:"echo"`
}

// 通用响应数据结构
type CQResponse struct {
	Time          int64  `json:"time"`
	SelfID        int64  `json:"self_id"`
	PostType      string `json:"post_type"`
	MetaEventType string `json:"meta_event_type"`
	Status        string `json:"status"`
	Retcode       int    `json:"retcode"`
	Msg           string `json:"msg"`
	Wording       string `json:"wording"`
	Data          any    `json:"data"`
	Echo          string `json:"echo"`
}

// 事件数据结构-收到频道消息
type GuildMsg struct {
	PostType    string  `json:"post_type"`
	MessageType string  `json:"message_type"`
	SubType     string  `json:"sub_type"`
	GuildID     string  `json:"guild_id"`
	ChannelID   string  `json:"channel_id"`
	UserID      string  `json:"user_id"`
	MessageID   string  `json:"message_id"`
	Sender      *Sender `json:"sender"`
	Message     string  `json:"message"`
}

// 响应数据结构-获取频道列表
type GuildInfoResponse struct {
	GuildID        string `json:"guild_id"`
	GuildName      string `json:"guild_name"`
	GuildDisplayID string `json:"guild_display_id"`
}

// 请求数据结构-获取子频道列表
type GuildChannelListRequest struct {
	GuildID string `json:"guild_id"`
	NoCache bool   `json:"no_cache"`
}

// 响应数据结构-获取子频道列表
type GuildChannelListResponse struct {
	OwnerGuildID    string          `json:"owner_guild_id"`
	ChannelID       string          `json:"channel_id"`
	ChannelType     int32           `json:"channel_type"`
	ChannelName     string          `json:"channel_name"`
	CreateTime      int64           `json:"create_time"`
	CreatorTinyID   string          `json:"creator_tiny_id"`
	TalkPermission  int32           `json:"talk_permission"`
	VisibleType     int32           `json:"visible_type"`
	CurrentSLowMode int32           `json:"current_slow_mode"`
	SlowModes       []*SlowModeInfo `json:"slow_modes"`
}

// 请求数据结构-单独获取频道成员信息
type GetGuildMemberProfileRequest struct {
	GuildID string `json:"guild_id"`
	UserID  string `json:"user_id"`
}

// 响应数据结构-单独获取频道成员信息
type GetGuildMemberProfileResponse struct {
	TinyID    string      `json:"tiny_id"`
	Nickname  string      `json:"nickname"`
	AvatarURL string      `json:"avatar_url"`
	JoinTime  int64       `json:"join_time"`
	Roles     []*RoleInfo `json:"roles"`
}

// 请求数据结构-发送信息到子频道
type SendGuildChannelMsgRequest struct {
	GuildID   string `json:"guild_id"`
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

func ParseRecvPkt(data []byte) (result *CQResponse, err error) {
	result = &CQResponse{}
	err = json.Unmarshal(data, result)
	return result, err
}

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

func GetRawTextFromCQMessage(msg string) string {
	for k, v := range CQCodeTypes {
		format := fmt.Sprintf(`\[CQ:%s.*?\]`, k)
		rule := regexp.MustCompile(format)
		msg = rule.ReplaceAllString(msg, fmt.Sprintf("[%s]", v))
	}
	return msg
}
