module phoenixbuilder

go 1.18

require (
	github.com/cheggaaa/pb v1.0.29
	github.com/df-mc/goleveldb v1.1.9
	github.com/hashicorp/go-version v1.6.0
	rogchap.com/v8go v0.7.0
)

require (
	github.com/atomicgo/cursor v0.0.1 // indirect
	github.com/df-mc/atomic v1.10.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gookit/color v1.4.2 // indirect
	github.com/klauspost/cpuid v1.3.1 // indirect
	github.com/klauspost/reedsolomon v1.9.9 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mmcloughlin/avo v0.0.0-20200803215136-443f81d77104 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/templexxx/cpu v0.0.7 // indirect
	github.com/templexxx/xorsimd v0.4.1 // indirect
	github.com/tjfoc/gmsm v1.3.2 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/image v0.5.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/tools v0.1.12 // indirect
)

replace rogchap.com/v8go v0.7.0 => ./depends/v8go@v0.7.0

require (
	github.com/Tnze/go-mc v1.17.0
	github.com/andybalholm/brotli v1.0.3
	github.com/blakesmith/ar v0.0.0-20190502131153-809d4375e1fb
	github.com/disintegration/imaging v1.6.2
	github.com/go-gl/mathgl v1.0.0
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/klauspost/compress v1.13.6
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/muhammadmuzzammil1998/jsonc v0.0.0-20211230184646-baf1f7156737
	github.com/pterm/pterm v0.12.29
	github.com/sandertv/go-raknet v1.12.0
	github.com/xtaci/kcp-go/v5 v5.6.1
	github.com/yuin/gopher-lua v1.1.0
	go.kuoruan.net/v8go-polyfills v0.5.0
	go.uber.org/atomic v1.9.0
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/net v0.7.0
	golang.org/x/term v0.5.0
	golang.org/x/text v0.7.0
	gopkg.in/square/go-jose.v2 v2.6.0
)

replace github.com/Tnze/go-mc v1.17.0 => ./depends/go-mc@v1.17.0
