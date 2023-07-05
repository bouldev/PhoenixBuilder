//go:build darwin
// +build darwin

package triple_wrapper

func init() {
	triple.system.darwin = true
}
