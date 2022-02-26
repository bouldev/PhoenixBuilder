// +build ios

package platform_helper

// void NetworkRequest();
import "C"

func DoNetworkRequest() {
	C.NetworkRequest()
}