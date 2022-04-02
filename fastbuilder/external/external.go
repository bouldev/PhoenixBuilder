// +build do_not_build_this

package external

import (
	"phoenixbuilder/fastbuilder/environment"
	"crypto/elliptic"
	"crypto/ecdsa"
	"crypto/rand"
	"strings"
	"net"
)

type externalSession struct {
	privateKey *ecdsa.PrivateKey
	salt []byte
}

func ListenExt(env *environment.PBEnvironment, address string) {
	listener, err:=net.Listen("udp", address)
	if(err!=nil) {
		fmt.Printf("Failed to listen on address %s: %v\n",address,err)
		return
	}
	fmt.Printf("Listening for external connection on address %s\n",address)
	go func() {
		for {
			connection, err := listener.Accept()
			if(err!=nil) {
				fmt.Printf("accept() failed: %v\n",err)
				continue
			}
			rAddr:=connection.RemoteAddr().String()
			allow_unsafe:=false
			if(!strings.Contains(rAddr,"127.0.0.1")&&!strings.Contains(rAddr,"localhost")&&!strings.Contains(rAddr,"::1")) {
				allow_unsafe=true
			}
			privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			if(err!=nil) {
				panic("external.go: 38")
			}
			salt:=make([]byte,16)
			rand.Reader.Read(salt)
			fmt.Printf("External connection established, peer: %s\n",rAddr)
			go func() {
				
}