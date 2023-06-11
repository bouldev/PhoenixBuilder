package core

type Option string

var OptionDebug = Option("debug")

func checkOption(options []Option, option Option) bool {
	for _, opt := range options {
		if opt == option {
			return true
		}
	}
	return false
}
