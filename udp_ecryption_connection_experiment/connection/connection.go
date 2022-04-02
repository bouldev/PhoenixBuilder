package connection

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	kcp "github.com/xtaci/kcp-go/v5"
)

// not stream, but ensure a frame is send/recv
type ReliableConnetion interface {
	SendFrame([]byte) error
	RecvFrame() ([]byte, error)
}

type ReliableConnetionServerHandler interface {
	Listen(address string) error
	SetOnNewConnection(func(ReliableConnetion))
	SetOnAcceptNewConnectionFail(func(error))
	SetOnServerDown(func(interface{}))
}

// KCP Server and Client

type KCPConnetionServerHandler struct {
	onAcceptNewConnectonFail func(error)
	onNewConnection          func(ReliableConnetion)
	OnNewKCPConnection       func(*kcp.UDPSession)
	onServerDown             func(interface{})
}

func (s *KCPConnetionServerHandler) Listen(address string) error {
	listener, err := kcp.ListenWithOptions(address, nil, 10, 3)
	if err != nil {
		return fmt.Errorf("udp: listen fail (err=%v)", err)
	}
	go func() {
		defer func() {
			r := recover()
			s.onServerDown(r)
		}()
		for {
			proxyConn, err := listener.AcceptKCP()
			if err != nil {
				fmt.Printf("Transfer: accept new connection fail\n\t(err=%v)\n", err)
				if s.onAcceptNewConnectonFail != nil {
					s.onAcceptNewConnectonFail(err)
				}
				continue
			}
			remoteDescription := proxyConn.RemoteAddr().String()
			fmt.Printf("Transfer: accept new connection @ %v\n", remoteDescription)
			baseConn := &StreamChannelWrapper{reader: proxyConn, writer: proxyConn}
			encryptionConn := &EncryptedChannel{
				connection: baseConn,
				isInitator: false,
			}
			if encryptionConn.Init() != nil {
				fmt.Printf("Transfer: encryption init fail\n")
				s.onAcceptNewConnectonFail(fmt.Errorf("encryption init fail: %v", err))
				continue
			}
			s.onNewConnection(encryptionConn)
		}

	}()
	return nil
}

func (s *KCPConnetionServerHandler) SetOnNewConnection(fn func(ReliableConnetion)) {
	s.onNewConnection = fn
}
func (s *KCPConnetionServerHandler) SetOnAcceptNewConnectionFail(fn func(error)) {
	s.onAcceptNewConnectonFail = fn
}
func (s *KCPConnetionServerHandler) SetOnServerDown(fn func(r interface{})) {
	s.onServerDown = fn
}

func KCPDial(address string) (ReliableConnetion, error) {
	conn, err := kcp.DialWithOptions(address, nil, 10, 3)
	if err != nil {
		return nil, err
	}
	baseConn := &StreamChannelWrapper{reader: conn, writer: conn}
	encryptedConn := &EncryptedChannel{
		connection: baseConn,
		isInitator: true,
	}
	if err := encryptedConn.Init(); err != nil {
		return nil, err
	}
	return encryptedConn, nil
}

// Frame -> Stream -> Frame

type ByteReaderWrapper struct {
	reader io.Reader
	buf    []byte
}

func (brw *ByteReaderWrapper) ReadByte() (byte, error) {
	if brw.buf == nil {
		brw.buf = make([]byte, 1)
	}
	_, err := brw.reader.Read(brw.buf)
	if err != nil {
		return 0, err
	}
	return brw.buf[0], nil
}

type StreamChannelWrapper struct {
	reader   io.Reader
	writer   io.Writer
	isClosed bool
}

const MAX_STD_HEADER_LEN = 1 << 15

func (scw *StreamChannelWrapper) RecvFrame() ([]byte, error) {
	// get length info
	length := uint64(0)
	header2Byte := make([]byte, 2)
	n, err := scw.reader.Read(header2Byte)
	if err != nil || n != 2 {
		scw.isClosed = true
		return nil, err
	}
	shortLength := binary.LittleEndian.Uint16(header2Byte)
	if shortLength > MAX_STD_HEADER_LEN {
		// use var_uint (extend header)
		uvarint, err := binary.ReadUvarint(&ByteReaderWrapper{reader: scw.reader})
		if err != nil {
			scw.isClosed = true
			return nil, err
		}
		length = uvarint
	} else {
		// ok
		length = uint64(shortLength)
	}
	// fmt.Println(length)
	fullData := make([]byte, length)
	offset := uint64(0)
	for offset < length {
		n, err := scw.reader.Read(fullData[offset:])
		if err != nil {
			scw.isClosed = true
			return nil, err

		}
		offset += uint64(n)
	}
	return fullData, nil
}

func (scw *StreamChannelWrapper) SendFrame(data []byte) error {
	length := len(data)
	headerBytes := make([]byte, binary.MaxVarintLen64+2) // 2 is the length of short(std) head
	headerBytesLen := 0
	if length > MAX_STD_HEADER_LEN {
		// mark it is followed by an varint (extend header)
		binary.LittleEndian.PutUint16(headerBytes, uint16(MAX_STD_HEADER_LEN+1))
		n := binary.PutUvarint(headerBytes[2:], uint64(length))
		headerBytesLen = n + 2
	} else {
		binary.LittleEndian.PutUint16(headerBytes, uint16(length))
		headerBytesLen = 2
	}
	offset := 0
	for offset < headerBytesLen {
		n, err := scw.writer.Write(headerBytes[offset:headerBytesLen])
		if err != nil {
			return err
		}
		offset += n
	}
	offset = 0
	for offset < length {
		n, err := scw.writer.Write(data[offset:])
		if err != nil {
			return err
		}
		offset += n
	}
	return nil
}

// encryption channel
type EncryptedChannel struct {
	connection ReliableConnetion
	isInitator bool
	encryptor  *EncryptionSession
	isClosed   bool
}

func (i *EncryptedChannel) initiateEncryptSession() error {
	initiatorPrivateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	salt := make([]byte, 16)
	rand.Read(salt)
	encodedInitiatorPublicKey, _ := x509.MarshalPKIXPublicKey(&initiatorPrivateKey.PublicKey)
	keyForInit := base64.RawStdEncoding.EncodeToString(encodedInitiatorPublicKey) + "-" + base64.RawStdEncoding.EncodeToString(salt)
	err := i.connection.SendFrame([]byte(keyForInit))
	if err != nil {
		return err
	}
	data, err := i.connection.RecvFrame()
	if err != nil {
		return fmt.Errorf("cannot get public key from responder: %v", err)
	}
	responderPubKeyData, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return fmt.Errorf("error parsing public key: %v", err)
	}
	responderPublicKey := new(ecdsa.PublicKey)
	ecdsaKey, ok := responderPubKeyData.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected ECDSA public key, but got %v", responderPubKeyData)
	}
	*responderPublicKey = *ecdsaKey
	i.encryptor = &EncryptionSession{PublicKey: responderPublicKey, PrivateKey: initiatorPrivateKey, Salt: salt}
	return i.encryptor.Init()
}

func (r *EncryptedChannel) waitForEncryptSession() error {
	initiatorPubkeyAndSaltData, err := r.connection.RecvFrame()
	if err != nil {
		return fmt.Errorf("cannot get public key from Initiator: %v", err)
	}
	responderPrivateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	encodedResponderPublicKey, _ := x509.MarshalPKIXPublicKey(&responderPrivateKey.PublicKey)
	err = r.connection.SendFrame(encodedResponderPublicKey)
	if err != nil {
		return err
	}
	b64InitiatorPubkeyAndSalt := strings.Split(string(initiatorPubkeyAndSaltData), "-")
	if len(b64InitiatorPubkeyAndSalt) != 2 {
		return fmt.Errorf("incorrect initiator PublicKey and Salt Data format")
	}
	data, err := base64.StdEncoding.DecodeString(b64InitiatorPubkeyAndSalt[0])
	if err != nil {
		return fmt.Errorf("error 64decode public key: %v", err)
	}
	initiatorPublicKeyData, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return fmt.Errorf("error parsing public key: %v", err)
	}
	initiatorPublicKey := new(ecdsa.PublicKey)
	ecdsaKey, ok := initiatorPublicKeyData.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected ECDSA public key, but got %v", initiatorPublicKeyData)
	}
	*initiatorPublicKey = *ecdsaKey
	salt, err := base64.RawStdEncoding.DecodeString(b64InitiatorPubkeyAndSalt[1])
	if err != nil {
		return fmt.Errorf("error 64decode salt: %v", err)
	}
	r.encryptor = &EncryptionSession{PublicKey: initiatorPublicKey, PrivateKey: responderPrivateKey, Salt: salt}
	return r.encryptor.Init()
}

func (e *EncryptedChannel) Init() error {
	if e.isInitator {
		return e.initiateEncryptSession()
	} else {
		return e.waitForEncryptSession()
	}
}

func (e *EncryptedChannel) SendFrame(data []byte) error {
	encyptedData := data[:]
	e.encryptor.Encrypt(encyptedData)
	return e.connection.SendFrame(encyptedData)
}

func (e *EncryptedChannel) RecvFrame() ([]byte, error) {
	encyptedData, err := e.connection.RecvFrame()
	if err != nil {
		e.isClosed = true
		return nil, err
	}
	e.encryptor.Decrypt(encyptedData)
	return encyptedData, nil
}
