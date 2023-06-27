package fbauth

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gorilla/websocket"
)

type Client struct {
	privateKey   *ecdsa.PrivateKey
	rsaPublicKey *rsa.PublicKey

	salt     []byte
	wsClient *websocket.Conn

	peerNoEncryption bool
	encryptor        *encryptionSession
	serverResponse   chan map[string]interface{}

	readCtx       context.Context
	waitEncrypted chan struct{}
	noEcrypted    bool
	lastReadErr   error
}

func NewClient(ctx context.Context) *Client {
	return &Client{
		readCtx: ctx,
	}
}

func (c *Client) initReadLoop() {
	var err error
	defer func() {
		c.lastReadErr = err
		if c.lastReadErr == nil {
			c.lastReadErr = fmt.Errorf("unknown error")
		}
	}()
	var msg []byte
	for {
		_, msg, err = c.wsClient.ReadMessage()
		if err != nil {
			break
		}
		if err = c.readCtx.Err(); err != nil {
			break
		}
		var message map[string]interface{}
		var outbuf bytes.Buffer
		var inbuf bytes.Buffer
		inbuf.Write(msg)
		reader, _ := gzip.NewReader(&inbuf)
		reader.Close()
		io.Copy(&outbuf, reader)
		msg = outbuf.Bytes()
		if c.encryptor != nil {
			c.encryptor.decrypt(msg)
		}
		if err = json.Unmarshal(msg, &message); err != nil {
			break
		}
		if action, ok := message["action"].(string); !ok {
			select {
			case c.serverResponse <- message:
			default:
			}
		} else if action == "encryption" {
			spub := new(ecdsa.PublicKey)
			keyb64, _ := message["publicKey"].(string)
			keydata, _ := base64.StdEncoding.DecodeString(keyb64)
			spp, _ := x509.ParsePKIXPublicKey(keydata)
			ek, _ := spp.(*ecdsa.PublicKey)
			*spub = *ek
			c.encryptor = &encryptionSession{
				serverPrivateKey: c.privateKey,
				clientPublicKey:  spub,
				salt:             c.salt,
			}
			if err = c.encryptor.init(); err != nil {
				break
			}
			close(c.waitEncrypted)
			continue
		} else if action == "no_encryption" {
			c.noEcrypted = true
			if err = c.sendMessage([]byte(`{"action":"accept_no_encryption"}`)); err != nil {
				break
			}
			close(c.waitEncrypted)
		}
	}
}

func (c *Client) SendMessage(data []byte) (err error) {
	if c.encryptor == nil && c.noEcrypted == false {
		return fmt.Errorf("should wait until channel is ecrypted")
	}
	if !c.noEcrypted && c.encryptor != nil {
		c.encryptor.encrypt(data)
	}
	err = c.sendMessage(data)
	if err != nil {
		return err
	}
	return c.lastReadErr
}

func (c *Client) sendMessage(data []byte) (err error) {
	if c.readCtx.Err() != nil {
		return fmt.Errorf("connection closed: %v (read loop: %v)", c.readCtx.Err(), c.lastReadErr)
	}
	var inbuf bytes.Buffer
	wr := gzip.NewWriter(&inbuf)
	_, err = wr.Write(data)
	if err != nil {
		return err
	}
	err = wr.Close()
	if err != nil {
		return err
	}
	err = c.wsClient.WriteMessage(websocket.BinaryMessage, inbuf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) sendMessageAndGetResponse(data []byte) (resp <-chan map[string]interface{}, err error) {
	if err = c.SendMessage(data); err != nil {
		return
	}
	responseC := make(chan map[string]interface{}, 1)
	go func() {
		r := <-c.serverResponse
		responseC <- r
	}()
	return responseC, nil
}

func (c *Client) SendMessageAndGetResponseWithDeadline(data []byte, deadline context.Context) (resp map[string]interface{}, err error) {
	responseC, err := c.sendMessageAndGetResponse(data)
	if err != nil {
		return nil, err
	}
	select {
	case <-deadline.Done():
		return nil, fmt.Errorf("auth server no response before deadline")
	case resp = <-responseC:
		return resp, nil
	}
}

func (c *Client) sendEncryptRequest() error {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&c.privateKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyBytes)
	if err = c.sendMessage([]byte(`{"action":"enable_encryption_v2","publicKey":"` + string(publicKeyStr) + `"}`)); err != nil {
		return err
	}
	return nil
}

func (c *Client) EstablishConnectionToAuthServer(connectContext context.Context, authServerAddr string) (err error) {
	if c.privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader); err != nil {
		return
	}
	c.waitEncrypted = make(chan struct{})
	c.salt = []byte("2345678987654321")
	c.serverResponse = make(chan map[string]interface{})
	var cancelFn context.CancelFunc
	c.readCtx, cancelFn = context.WithCancel(c.readCtx)
	if c.wsClient, _, err = websocket.DefaultDialer.DialContext(connectContext, authServerAddr, nil); err != nil {
		cancelFn()
		return fmt.Errorf("cannot connect to auth server")
	}
	go func() {
		c.initReadLoop()
		err = fmt.Errorf("fbauth server: read loop closed with error: %v", c.lastReadErr)
		fmt.Println(err)
		cancelFn()
	}()
	if err = c.sendEncryptRequest(); err != nil {
		return err
	}
	select {
	case <-connectContext.Done():
		return connectContext.Err()
	case <-c.readCtx.Done():
		return c.readCtx.Err()
	case <-c.waitEncrypted:
		return nil
	}
}

func (c *Client) Closed() <-chan struct{} {
	return c.readCtx.Done()
}
