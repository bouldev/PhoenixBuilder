// +build ios

package platform_helper

/*
#cgo CFLAGS: -fobjc-arc

void NetworkRequest();
void playBackgroundMusic();
void stopBackgroundMusic();
*/
import "C"

func DoNetworkRequest() {
	C.NetworkRequest()
}

func RunBackground() {
	C.playBackgroundMusic()
}

func StopBackground() {
	C.stopBackgroundMusic()
}