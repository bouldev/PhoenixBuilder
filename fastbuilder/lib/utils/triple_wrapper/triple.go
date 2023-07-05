package triple_wrapper

type Triple struct {
	system struct {
		windows bool
		linux   bool
		darwin  bool
		android bool
	}
	arch struct {
		amd64 bool
		arm64 bool
		i386  bool
	}
}

var triple = &Triple{}

func GetSystemStr() string {
	if triple.system.windows {
		return "windows"
	}
	if triple.system.linux {
		return "linux"
	}
	if triple.system.darwin {
		return "darwin"
	}
	if triple.system.android {
		return "android"
	}
	return "unknown"
}

func GetArchStr() string {
	if triple.arch.amd64 {
		return "amd64"
	}
	if triple.arch.arm64 {
		return "arm64"
	}
	if triple.arch.i386 {
		return "x86"
	}
	return "unknown"
}

func GetSystemArchStr() string {
	return GetSystemStr() + "-" + GetArchStr()
}

func IsWindows() bool {
	return triple.system.windows
}
