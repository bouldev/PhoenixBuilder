all:
	go build -trimpath -ldflags "-s -w" -o phoenixbuilder
clean:
	rm phoenixbuilder