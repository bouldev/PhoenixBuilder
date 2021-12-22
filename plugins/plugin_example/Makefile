.PHONY: all current install clean
all: current install

current: plugin_example.so
plugin_example.so: main.go
	CGO_ENABLED=1 go build -buildmode=plugin -trimpath -ldflags "-s -w" -o plugin_example.so
install: plugin_example.so
	@echo Installing
	cp plugin_example.so ~/.config/fastbuilder/plugins/
clean:
	rm plugin_example.so
