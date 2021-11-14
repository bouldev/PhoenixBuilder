.PHONY: all current
all: current

current: plugin_example.so
	CGO_ENABLED=1 go build -buildmode=plugin -trimpath -ldflags "-s -w" -o plugin_example.so