package access_helper

import (
	"errors"
	"phoenixbuilder/fastbuilder/i18n"
)

var ErrFBUserCenterLoginFail = errors.New(I18n.T(I18n.Auth_InvalidUser))
var ErrRentalServerDisconnected = errors.New("connection unexpectedly closed")
var ErrFBServerConnectionTimeOut = errors.New("connection timed out")
var ErrGetTokenTimeOut = errors.New("login timed out")
var ErrFailToConnectFBServer = errors.New("connection to authentication server timed out")
var ErrRentalServerConnectionTimeOut = errors.New("connection to server timed out")
var ErrFailToConnectRentalServer = errors.New("failed to connect to server.")
var ErrFBChallengeSolvingTimeout = errors.New("challenge solving timed out")
var ErrFBTransferDataTimeout = errors.New("netease authentication data calculation timed out")
var ErrFBTransferDataFail = errors.New("netease authentication data calculation failed")
var ErrFBTransferCheckNumTimeOut = errors.New("netease authentication data calculation timed out")
var ErrFBTransferCheckNumFail = errors.New("netease authentication data calculation failed")
var ErrBotOpPrivilegeRemoved = errors.New("privilege lost")
