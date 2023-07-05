package rental_server_impactor

import (
	"errors"
	"phoenixbuilder/fastbuilder/i18n"
)

var ErrFBUserCenterLoginFail = errors.New(I18n.T(I18n.Auth_InvalidUser))
var ErrRentalServerDisconnected = errors.New("Connection unexpectedly closed")
var ErrFBServerConnectionTimeOut = errors.New("Connection timed out")
var ErrGetTokenTimeOut = errors.New("Login timed out")
var ErrFailToConnectFBServer = errors.New("Connection to authentication server timed out")
var ErrRentalServerConnectionTimeOut = errors.New("Connection to server timed out")
var ErrFailToConnectRentalServer = errors.New("Failed to connect to server.")
var ErrFBTransferDataTimeout = errors.New("Netease authentication data calculation timed out")
var ErrFBTransferDataFail = errors.New("Netease authentication data calculation failed")
var ErrFBTransferCheckNumTimeOut = errors.New("Netease authentication data calculation timed out")
var ErrFBTransferCheckNumFail = errors.New("Netease authentication data calculation failed")
var ErrBotOpPrivilegeRemoved = errors.New("Privilege lost")
