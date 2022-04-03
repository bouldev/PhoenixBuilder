package fbauth
import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"github.com/gorilla/websocket"
	"crypto/x509"
	"encoding/base64"
	"compress/gzip"
	"bytes"
	"io"
	"fmt"
	"phoenixbuilder/bridge/bridge_fmt"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/environment"
)

type Client struct {
	privateKey *ecdsa.PrivateKey
	rsaPublicKey *rsa.PublicKey
	
	salt []byte
	client *websocket.Conn
	
	encryptor *encryptionSession
	serverResponse chan map[string]interface{}
	
	closed bool
	
	env *environment.PBEnvironment
}

func CreateClient(env *environment.PBEnvironment) *Client {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		panic(err)
	}
	salt := []byte("2345678987654321")
	authclient := &Client {
		privateKey:privateKey,
		salt:salt,
		serverResponse:make(chan map[string]interface{}),
		closed:false,
		env: env,
	}
	cl,_,err:=websocket.DefaultDialer.Dial(args.AuthServer(),nil)
	if err != nil {
		panic(err)
	}
	authclient.client=cl
	encrypted := make(chan struct{})
	go func() {
		defer func() {
			authclient.closed=true
		}()
		//defer panic("Core feature works incorrectly")
		for {
			_, msg, err:=cl.ReadMessage()
			if err != nil {
				break
			}
			var message map[string]interface{}
			var outbuf bytes.Buffer
			var inbuf bytes.Buffer
			inbuf.Write(msg)
			reader,_:=gzip.NewReader(&inbuf)
			reader.Close()
			io.Copy(&outbuf,reader)
			msg=outbuf.Bytes()
			if authclient.encryptor!= nil {
				authclient.encryptor.decrypt(msg)
			}
			json.Unmarshal(msg,&message)
			msgaction,_:=message["action"].(string)
			if msgaction=="encryption" {
				spub:=new(ecdsa.PublicKey)
				keyb64,_:=message["publicKey"].(string)
				keydata, _:=base64.StdEncoding.DecodeString(keyb64)
				spp,_:=x509.ParsePKIXPublicKey(keydata)
				ek,_ := spp.(*ecdsa.PublicKey)
				*spub=*ek
				authclient.encryptor=&encryptionSession {
					serverPrivateKey:privateKey,
					clientPublicKey:spub,
					salt:authclient.salt,
				}
				authclient.encryptor.init()
				close(encrypted)
				continue
			}else if msgaction=="world_chat" {
				chat_msg,_:=message["msg"].(string)
				chat_sender,_:=message["username"].(string)
				select {
				case env.WorldChatChannel<-[]string{chat_sender,chat_msg}:
					continue
				default:
					continue
				}
			}
			select{
			case authclient.serverResponse<-message:
				continue
			default:
				continue
			}
		}
	}()
	pubb,err:=x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err!=nil {
		panic(err)
	}
	pub_str:=base64.StdEncoding.EncodeToString(pubb)
	var inbuf bytes.Buffer
	wr:=gzip.NewWriter(&inbuf)
	wr.Write([]byte(`{"action":"enable_encryption_v2","publicKey":"`+string(pub_str)+`"}`))
	wr.Close()
	cl.WriteMessage(websocket.BinaryMessage,inbuf.Bytes())
	for {
		select {
		case <-encrypted:
			return authclient
		}
	}
	return authclient
}

func (client *Client) CanSendMessage() bool {
	return client.encryptor!=nil&&!client.closed
}

func (client *Client) SendMessage(data[] byte){
	if client.encryptor==nil {
		panic("早すぎる")
	}
	if client.closed {
		bridge_fmt.Println("Error: SendMessage: Connection closed")
		panic("Message after auth close")
	}
	client.encryptor.encrypt(data)
	var inbuf bytes.Buffer
	wr:=gzip.NewWriter(&inbuf)
	wr.Write(data)
	wr.Close()
	client.client.WriteMessage(websocket.BinaryMessage,inbuf.Bytes())
}

type AuthRequest struct {
	Action string `json:"action"`
	ServerCode string `json:"serverCode"`
	ServerPassword string `json:"serverPassword"`
	Key string `json:"publicKey"`
	FBToken string
}

func (client *Client) Auth(serverCode string,serverPassword string,key string,fbtoken string) (string,int,error) {
	authreq:=&AuthRequest {
		Action:"phoenix::login",
		ServerCode:serverCode,
		ServerPassword:serverPassword,
		Key:key,
		FBToken:fbtoken,
	}
	msg,err:=json.Marshal(authreq)
	if err!=nil {
		panic("Failed to encode json")
	}
	client.SendMessage(msg)
	resp,_:=<-client.serverResponse
	code,_:=resp["code"].(float64)
	if code!=0 {
		err,_:=resp["message"].(string)
		trans,hasTranslation:=resp["translation"].(float64)
		if(hasTranslation) {
			err=I18n.T(uint16(trans))
		}
		return "",int(code),fmt.Errorf("%s",err)
	}
	uc_username, _ := resp["username"].(string)
	u_uid, _ := resp["uid"].(string)
	client.env.FBUCUsername=uc_username
	client.env.Uid=u_uid
	str,_:=resp["chainInfo"].(string)
	return str,0,nil
}

type RespondRequest struct {
	Action string `json:"action"`
}

func (client *Client) ShouldRespondUser() string {
	rspreq:=&RespondRequest {
		Action:"phoenix::get-user",
	}
	msg,err:=json.Marshal(rspreq)
	if err!=nil {
		panic("Failed to encode json")
		//return true
		//Torrekie 22/07/21 13.12: Don't understand why this, but LNSSPsd let me made this edit
		return ""
	}
	client.SendMessage(msg)
	resp,_:=<-client.serverResponse
	code,_:=resp["code"].(float64)
	if code != 0 {
		//This should never happen
		bridge_fmt.Println("UNK_1")
		panic("??????")
		//return true
		//Torrekie 22/07/21 13.12: and this
		return ""
	}
	shouldRespond,_:=resp["username"].(string)
	return shouldRespond
}

type FTokenRequest struct {
	Action string `json:"action"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (client *Client) GetToken(username string,password string) string {
	rspreq:=&FTokenRequest {
		Action:"phoenix::get-token",
		Username:username,
		Password:password,
	}
	msg,err:=json.Marshal(rspreq)
	if err!=nil {
		panic("Failed to encode json")
	}
	client.SendMessage(msg)
	resp,_:=<-client.serverResponse
	code,_:=resp["code"].(float64)
	if code != 0 {
		return ""
	}
	usertoken,_:=resp["token"].(string)
	return usertoken
}

type FEncRequest struct {
	Action string `json:"action"`
	Content string `json:"content"`
	Uid string `json:"uid"`
}

func (client *Client) TransferData(content string, uid string) string {
	rspreq:=&FEncRequest {
		Action:"phoenix::transfer-data",
		Content: content,
		Uid: uid,
	}
	msg,err:=json.Marshal(rspreq)
	if err!=nil {
		panic("Failed to encode json")
	}
	client.SendMessage(msg)
	resp,_:=<-client.serverResponse
	code,_:=resp["code"].(float64)
	if code != 0 {
		panic("Failed to transfer start type")
	}
	data,_:=resp["data"].(string)
	return data
}

type WorldChatRequest struct {
	Category string `json:"category"`
	Action string `json:"action"`
	Message string `json:"message"`
}

func (client *Client) WorldChat(message string) {
	req:=&WorldChatRequest {
		Category: "gaming",
		Action:   "world_chat",
		Message:  message,
	}
	msg, err:=json.Marshal(req)
	if(err!=nil) {
		panic("Failed to encode json 254")
	}
	client.SendMessage(msg)
	return
}