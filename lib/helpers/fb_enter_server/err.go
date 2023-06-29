package fb_enter_server

import "errors"

var ErrFBUserCenterLoginFail = errors.New("无效的 Fastbuilder 用户名或密码")
var ErrRentalServerDisconnected = errors.New("与租赁服的连接已断开")
var ErrFBServerConnectionTimeOut = errors.New("与FB服务器建立连接超时，检查网络")
var ErrGetTokenTimeOut = errors.New("FB用户登陆及获取Token超时")
var ErrFailToConnectFBServer = errors.New("无法与FB服务器建立连接")
var ErrRentalServerConnectionTimeOut = errors.New("与网易租赁服建立连接超时，检查网络")
var ErrFailToConnectRentalServer = errors.New("无法与网易租赁服建立连接")
var ErrFBTransferDataTimeOut = errors.New("FB服务器处理网易租赁服TransferData请求超时")
var ErrFBTransferDataFail = errors.New("FB服务器处理网易租赁服TransferData失败")
var ErrFBTransferCheckNumTimeOut = errors.New("FB服务器处理网易租赁服TransferCheckNum请求超时")
var ErrFBTransferCheckNumFail = errors.New("FB服务器处理网易租赁服TransferCheckNum失败")
var ErrBotOpPrivilegeRemoved = errors.New("机器人OP权限被移除")
