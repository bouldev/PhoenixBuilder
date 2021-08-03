package hotbarmanager

import (
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/builder"
	"sync"
)

var waitingMap sync.Map
var processedAlarm chan bool

func ProcessInPacket(items []protocol.ItemInstance) {
	for i:=0;i<9;i++ {
		waiter, loaded := waitingMap.LoadAndDelete(i)
		if !loaded {
			continue
		}
		inp,_:=waiter.(chan *mctype.Block)
		inp<-&mctype.Block {
			Name: &builder.PEBlockStr[items[i].Stack.ItemType.NetworkID],
			Data: items[i].Stack.ItemType.MetadataValue,
		}
	}
	select {
	case processedAlarm<-true:
		return
	default:
		return
	}
	// Notify if waiting, directly return otherwise.
}

func Init() {
	processedAlarm=make(chan bool)
}

func RegisterWaiter(c chan *mctype.Block) int64 {
	for {
		for i:=0;i<9;i++ {
			_, ok:=waitingMap.LoadOrStore(i,c)
			if !ok {
				return int64(i)
			}
		}
		<-processedAlarm
	}
}