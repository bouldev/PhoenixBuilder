package move

import (
	//"fmt"
	"math"
	"time"
	"github.com/go-gl/mathgl/mgl32"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

var ConnectTime time.Time
var Position mgl32.Vec3
var Pitch, Yaw float32
var HeadYaw float32
var Connection *minecraft.Conn
var RuntimeID uint64
var MoveP float32
var Target mgl32.Vec3
var TargetRuntimeID uint64 = 0

func calculateTick() uint64 {
	return uint64(time.Now().Sub(ConnectTime).Milliseconds() / 50)
}

func Move(x,y,z float32) {
	//fmt.Printf("Decided: %f, %f\n",x,z)
	moveVector3:=mgl32.Vec3{x,y/*+256*/,z}
	moveVector:=mgl32.Vec2{x,z}
	Position=Position.Add(moveVector3)
	Connection.WritePacket(&packet.PlayerAuthInput {
		Pitch: Pitch,
		Yaw: Yaw,
		Position: Position,
		MoveVector: moveVector,
		HeadYaw: HeadYaw,
		InputData: 0|packet.InputFlagAutoJumpingInWater,
		InputMode: packet.InputModeTouch,
		PlayMode: packet.PlayModeScreen,
		Tick: calculateTick(),
		Delta: moveVector3,
	})
}

func Jump() {
	Connection.WritePacket(&packet.PlayerAction {
		EntityRuntimeID: RuntimeID,
		ActionType: protocol.PlayerActionJump,
	})
	moveVector3:=mgl32.Vec3{0,256,0}
	moveVector:=mgl32.Vec2{0,0}
	Position=Position.Add(moveVector3)
	Connection.WritePacket(&packet.PlayerAuthInput {
		Pitch: Pitch,
		Yaw: Yaw,
		Position: Position,
		MoveVector: moveVector,
		HeadYaw: HeadYaw,
		InputData: packet.InputFlagJumping,
		InputMode: packet.InputModeTouch,
		PlayMode: packet.PlayModeScreen,
		Tick: calculateTick(),
		Delta: moveVector3,
	})
}

func getz(v float32) float32 {
	if v>0 {
		return v
	}
	return -v
}

var nextAttack = 0

func Auto() {
	delta:=Target.Sub(Position)
	deltax:=delta[0]
	deltaz:=delta[2]
	if math.IsNaN(float64(deltax)) {
		deltax=0
	}
	if math.IsNaN(float64(deltaz)) {
		deltaz=0
	}
	zdx:=getz(deltax)
	zdz:=getz(deltaz)
	if zdx<6 && zdz<6 {
		if nextAttack!=0 {
			nextAttack--
		}else{
			Connection.WritePacket(&packet.InventoryTransaction {
				TransactionData: &protocol.UseItemOnEntityTransactionData {
					TargetEntityRuntimeID: TargetRuntimeID,
					ActionType: 0,
					HotBarSlot: 0,
					HeldItem: protocol.ItemInstance {
						StackNetworkID: 0,
						Stack: protocol.ItemStack {
							ItemType: protocol.ItemType {
								NetworkID: 0,
								MetadataValue: 0,
							},
							BlockRuntimeID: 0,
							Count: 0,
						},
					},
					Position: Target,
					ClickedPosition: Target.Sub(mgl32.Vec3{0,-1,0}),
				},
			})
			nextAttack=10
		}
	}
	maxItem:=zdx
	if zdz>maxItem {
		maxItem=zdz
	}
	//fmt.Printf("Target: %v, delta: %v, max: %f, ",Target,delta,maxItem)
	Move(2*(zdx/maxItem)*(zdx/deltax),0,2*(zdz/maxItem)*(zdz/deltaz))
}