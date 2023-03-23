package packet

import "fastbuilder-core/lib/minecraft/gophertunnel/protocol"

const (
	SimulationTypeGame byte = iota
	SimulationTypeEditor
	SimulationTypeTest
	SimulationTypeInvalid
)

// SimulationType is an in-progress packet. We currently do not know the use case.
type SimulationType struct {
	// SimulationType is the simulation type selected.
	SimulationType byte
}

// ID ...
func (*SimulationType) ID() uint32 {
	return IDSimulationType
}

// Marshal ...
func (pk *SimulationType) Marshal(w *protocol.Writer) {
	w.Uint8(&pk.SimulationType)
}

// Unmarshal ...
func (pk *SimulationType) Unmarshal(r *protocol.Reader) {
	r.Uint8(&pk.SimulationType)
}
