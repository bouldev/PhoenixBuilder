// +build ios

package platform_helper

/*
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