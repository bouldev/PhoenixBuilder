//go:build linux && !android
// +build linux,!android

package triple_wrapper

func init() {
	triple.system.linux = true
}
