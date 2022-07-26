package embed_binary

const (
	WINDOWS_x86_64 = "windows_x86_64"
	MACOS_x86_64   = "macos_x86_64"
	Linux_x86_64   = "linux_x86_64"
	Android_arm64  = "android_arm64"
)

func GetCqHttpBinary() []byte {
	return embedding_cqhttp
}

func GetPlantform() string {
	return PLANTFORM
}
