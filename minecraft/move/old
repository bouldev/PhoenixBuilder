package move

import (
	"time"
	"github.com/go-gl/mathgl/mgl32"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/command"
	"github.com/google/uuid"
	"fmt"
	"encoding/json"
	"math"
	"math/rand"
)

func decideWhether(chance int) bool {
	rand.Seed(time.Now().UnixNano())
	r:=rand.Intn(100)
	if r<=chance {
		return true
	}
	return false
}

func breakLenAndDegVec2(len float32, deg float32) mgl32.Vec2 {
	// cos(deg)=x/len
	// sin(deg)=z/len
	// x=len*cos(deg)
	// z=len*sin(deg)
	a, b:=math.Sincos(float64(mgl32.DegToRad(deg)))
	return mgl32.Vec2 { len*float32(b), len*float32(a) }
}

func breakLenAndDeg(len float32, deg float32) mgl32.Vec3 {
	vec:=breakLenAndDegVec2(len, deg)
	return mgl32.Vec3 { vec[0], 0, vec[1] }
}

// 0: ok
// 1: fully blocked
// 2: jumpable
// 3: x-3=falldistance
func getBlockedStatus(conn *minecraft.Conn,bpos protocol.BlockPos) byte {
	quid,_:=uuid.NewUUID()
	wc:=make(chan *packet.CommandOutput)
	command.UUIDMap.Store(quid.String(),wc)
	command.SendWSCommand(fmt.Sprintf("gettopsolidblock %d %d %d",bpos[0],bpos[1]+1,bpos[2]),quid,conn)
	res:=<-wc
	close(wc)
	if res.SuccessCount==0 {
		return 1
	}
	var msg map[string]interface{}
	json.Unmarshal([]byte(res.UnknownString),&msg)
	pos:=msg["position"].(map[string]interface{})
	fallposy:=pos["y"].(float64)
	fall:=byte(bpos[1]-int32(fallposy)+3)
	if fall==3 {
		return 0
	}
	return fall
}

func randInt64(min, max int64) int64 {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Int63n(max-min) + min
}

func toBlockPos(pos mgl32.Vec3) protocol.BlockPos {
	return protocol.BlockPos { int32(pos[0]), int32(pos[1]), int32(pos[2]) }
}

func playerFakeMove(conn *minecraft.Conn,pos mgl32.Vec3, yaw float32) {
	command.SendSizukanaCommand(fmt.Sprintf("tp %s %f %f %f 0 %d false", conn.IdentityData().DisplayName, pos[0], pos[1], pos[2], int(yaw)), conn)
}

func InitMoveSystem(conn *minecraft.Conn) {
	return
	pos:=conn.GameData().PlayerPosition
	yaw:=float32(0)
	//bpos:=protocol.BlockPos { int32(pos[0]), int32(pos[1]), int32(pos[2]) }
	//lastblockstate:=byte(0)
	// speednormal: 0.03448275
	// speedinair:  0.01
	// fall: 0.5*9.8*((falltick*0.5)*(falltick*0.5))
	//speed:=0.03448275
	falltick:=0
	fall_len:=float32(0)
	outofsync:=0
	falling:=false
	blockstateupdated:=false
	continuetofall:=false
	continuetojump:=false
	turn:=int64(0)
	doneturn:=float32(0)
	breakugoifneedturn:=false
	jumpv:=float32(0)
	tick:=uint64(0)
	go func() {
		for {
			tick++
			if yaw>=360 {
				yaw-=360
			}
			// move: 0.03448275
			if falling {
				fall:=float32(falltick)*(-jumpv)+0.5*9.8*((float32(falltick)*0.1)*(float32(falltick)*0.1))
				if fall<fall_len {
					if continuetojump {
						movevec:=breakLenAndDegVec2(0.03448275,yaw)
						pos=pos.Add(mgl32.Vec3 { movevec[0], 0, movevec[1] })
						//playerFakeMove(conn, pos.Add(mgl32.Vec3 { movevec[0], -fall, movevec[1] }), yaw)
						conn.WritePacket(&packet.PlayerAuthInput {
							Pitch: 0,
							Yaw: yaw,
							Position: pos.Add(mgl32.Vec3 { movevec[0], -fall, movevec[1] }),
							MoveVector: movevec,
							HeadYaw: yaw,
							InputData: packet.InputFlagDescend|packet.InputFlagChangeHeight|packet.InputFlagJumping|packet.InputFlagJumpDown|packet.InputFlagDown,
							InputMode: packet.InputModeTouch,
							PlayMode: packet.PlayModeNormal,
							Tick: tick,
						})
					}else{
						//playerFakeMove(conn, pos.Sub(mgl32.Vec3 { 0, fall, 0 }), yaw)
						conn.WritePacket(&packet.PlayerAuthInput {
							Pitch: 0,
							Yaw: yaw,
							Position: pos.Sub(mgl32.Vec3 { 0, fall, 0 }),
							MoveVector: mgl32.Vec2 { 0, 0 },
							HeadYaw: yaw,
							InputData: packet.InputFlagDescend|packet.InputFlagChangeHeight|packet.InputFlagDown,
							InputMode: packet.InputModeTouch,
							PlayMode: packet.PlayModeNormal,
							Tick: tick,
						})
					}
					falltick++
				}else{
					//playerFakeMove(conn, pos.Sub(mgl32.Vec3 { 0, fall_len, 0 }), yaw)
					conn.WritePacket(&packet.PlayerAuthInput {
						Pitch: 0,
						Yaw: yaw,
						Position: pos.Sub(mgl32.Vec3 { 0, fall_len, 0 }),
						MoveVector: mgl32.Vec2 { 0, 0 },
						HeadYaw: yaw,
						InputData: packet.InputFlagDescend|packet.InputFlagChangeHeight|packet.InputFlagDown,
						InputMode: packet.InputModeTouch,
						PlayMode: packet.PlayModeNormal,
						Tick: tick,
					})
					continuetojump=false
					pos=pos.Sub(mgl32.Vec3 { 0, fall_len, 0 })
					falltick=0
					falling=false
				}
				time.Sleep(time.Duration(50)*time.Millisecond)
				continue
			}
			svec:=mgl32.Vec2 { pos[0]-float32(int(pos[0])), pos[2]-float32(int(pos[2])) }
			if svec.Len()>=0.4 && !blockstateupdated {
				stat:=getBlockedStatus(conn, toBlockPos(pos.Add(breakLenAndDeg(1,yaw))))
				if stat==1 {
					turn=randInt64(59,141)
					yaw+=10
					outofsync++ // If >= 9, no longer move
				}else if stat>3 {
					willfall:=stat-3
					if willfall<=2 && decideWhether(40) {
						fall_len=float32(willfall)
						continuetofall=true
						blockstateupdated=true
					}else if willfall==3 && decideWhether(30) {
						fall_len=float32(willfall)
						continuetofall=true
						blockstateupdated=true
					}else if willfall==4 && decideWhether(20) {
						fall_len=float32(willfall)
						continuetofall=true
						blockstateupdated=true
					}else if willfall>4 && decideWhether(1) {
						fall_len=float32(willfall)
						continuetofall=true
						blockstateupdated=true
					}else {
						turn=randInt64(60,140)
					}
					blockstateupdated=true
				}else if stat==2 {
					if decideWhether(40) {
						jumpv=2
						fall_len=-1
						continuetojump=true
						falling=true
					}else {
						turn=randInt64(59,141)
					}
					blockstateupdated=true
				}
			}else if svec.Len()<0.4 && blockstateupdated {
				blockstateupdated=false
				if continuetofall {
					falling=true
					continue
				}
			}
			if turn!=0 {
				yawadd:=float32(turn/6)
				yaw+=yawadd
				doneturn+=yawadd
				if breakugoifneedturn {
					//playerFakeMove(conn, pos, yaw)
					conn.WritePacket(&packet.PlayerAuthInput {
						Pitch: 0,
						Yaw: yaw,
						Position: pos,
						HeadYaw: yaw,
						InputData: 0,
						InputMode: packet.InputModeTouch,
						PlayMode: packet.PlayModeNormal,
						Tick: tick,
					})
					time.Sleep(time.Duration(50)*time.Millisecond)
					continue
				}
				if doneturn>=float32(turn) {
					doneturn=0
					turn=0
					blockstateupdated=false
					breakugoifneedturn=true
				}
			}
			if breakugoifneedturn {
				breakugoifneedturn=false
			}
			movevec:=breakLenAndDegVec2(0.03448275,yaw)
			if yaw>=360 {
				yaw-=360
			}
			//playerFakeMove(conn, pos.Add(mgl32.Vec3 { movevec[0], 0, movevec[1] }), yaw)
			conn.WritePacket(&packet.PlayerAuthInput {
				Pitch: 0,
				Yaw: yaw,
				Position: pos.Add(mgl32.Vec3 { movevec[0], 0, movevec[1] }),
				MoveVector: movevec,
				HeadYaw: yaw,
				InputData: 0,
				InputMode: packet.InputModeTouch,
				PlayMode: packet.PlayModeNormal,
				Tick: tick,
			})
			pos=pos.Add(mgl32.Vec3 { movevec[0], 0, movevec[1] })
			time.Sleep(time.Duration(50)*time.Millisecond)
		}
	} ()
}