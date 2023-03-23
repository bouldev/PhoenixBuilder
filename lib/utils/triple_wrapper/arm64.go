//go:build arm64
// +build arm64

package triple_wrapper

func init() {
	triple.arch.arm64 = true
}
