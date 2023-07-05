//go:build android
// +build android

package triple_wrapper

func init() {
	triple.system.android = true
}
