//go:build windows
// +build windows

package triple_wrapper

func init() {
	triple.system.windows = true
}
