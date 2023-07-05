package transfer

const (
	DefaultPubSubAccessPoint      = "tcp://*:24016"
	DefaultCtrlAccessPoint        = "tcp://*:24015"
	DefaultDirectPubSubModeEnable = true
	DefaultDirectSendModeEnable   = true
)

type EndPointOption struct {
	PubAccessPoint  string
	CtrlAccessPoint string
	DirectSendMode  bool
	DirectSubMode   bool
}

func MakeDefaultEndPointOption() *EndPointOption {
	return &EndPointOption{
		PubAccessPoint:  DefaultPubSubAccessPoint,
		CtrlAccessPoint: DefaultCtrlAccessPoint,
		DirectSendMode:  DefaultDirectSendModeEnable,
		DirectSubMode:   DefaultDirectPubSubModeEnable,
	}
}
