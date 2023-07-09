package mux_pumper

import (
	"context"
	"phoenixbuilder/fastbuilder/lib/utils/sync_wrapper"
	"phoenixbuilder/minecraft/protocol/packet"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type MCPacketNameIDMapping map[string]uint32

var mcPacketNameIDMapping MCPacketNameIDMapping

func initMCPacketNameIDMapping() {
	pool := packet.NewPool()
	mcPacketNameIDMapping = MCPacketNameIDMapping{}
	for id, pkMaker := range pool {
		pk := pkMaker()
		pkName := reflect.TypeOf(pk).Elem().Name()
		mcPacketNameIDMapping[pkName] = id
		// mcPacketNameIDMapping["ID"+pkName] = id
		// mcPacketNameIDMapping[fmt.Sprint(id)] = id
	}
}

func init() {
	initMCPacketNameIDMapping()
}

func stringWantsToIDSet(want []string) map[uint32]bool {
	s := map[uint32]bool{}
	for _, w := range want {
		if w == "any" || w == "all" {
			for _, id := range mcPacketNameIDMapping {
				s[id] = true
			}
			continue
		}
		add := true
		if strings.HasPrefix(w, "!") {
			add = false
			w = w[1:]
		}
		if strings.HasPrefix(w, "ID") {
			w = w[2:]
		}
		if id, found := mcPacketNameIDMapping[w]; found {
			if add {
				s[id] = true
			} else {
				delete(s, id)
			}
		}
	}
	return s
}

// should be no block
type PumperNoBlock func(pk packet.Packet) error

type GamePacketPumperMux struct {
	subPumpers map[uint32]*sync_wrapper.SyncMap[PumperNoBlock]
}

func NewGamePacketPumperMux() *GamePacketPumperMux {
	if len(mcPacketNameIDMapping) == 0 {
		initMCPacketNameIDMapping()
	}
	pm := &GamePacketPumperMux{
		subPumpers: map[uint32]*sync_wrapper.SyncMap[PumperNoBlock]{},
	}
	for _, id := range mcPacketNameIDMapping {
		pm.subPumpers[id] = sync_wrapper.NewInstanceMap[PumperNoBlock]()
	}
	return pm
}

func (p *GamePacketPumperMux) translateStringWantsToIDSet(want []string) map[uint32]bool {
	return stringWantsToIDSet(want)
}

func (p *GamePacketPumperMux) GetMCPacketNameIDMapping() MCPacketNameIDMapping {
	return mcPacketNameIDMapping
}

func (p *GamePacketPumperMux) PumpGamePacket(pk packet.Packet) {
	id := pk.ID()
	if subPumper, found := p.subPumpers[id]; found {
		toRemove := []string{}
		subPumper.Iter(func(k string, pumper PumperNoBlock) (continueIter bool) {
			err := pumper(pk)
			if err != nil {
				toRemove = append(toRemove, k)
			}
			return true
		})
		for _, k := range toRemove {
			subPumper.Delete(k)
		}
	}
}

func (p *GamePacketPumperMux) AddNewPumper(want []string, pumper PumperNoBlock) {
	idSet := p.translateStringWantsToIDSet(want)
	for id := range idSet {
		if subPumpers, found := p.subPumpers[id]; found {
			subPumpers.Set(uuid.New().String(), pumper)
		}
	}
}

func MakeMCPacketNoBlockFeeder(ctx context.Context, pkChan chan packet.Packet) PumperNoBlock {
	pumper := func(pk packet.Packet) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case pkChan <- pk:
		default:
		}
		return nil
	}
	return pumper
}
