package define

import (
	"fmt"
)

type ChunkPos [2]int32

// X returns the X coordinate of the chunk position.
func (p ChunkPos) X() int32 {
	return p[0]
}

// Z returns the Z coordinate of the chunk position.
func (p ChunkPos) Z() int32 {
	return p[1]
}

// String implements fmt.Stringer and returns (x, z).
func (p ChunkPos) String() string {
	return fmt.Sprintf("(%v, %v)", p[0], p[1])
}

// Pos holds the position of a block. The position is represented of an array with an x, y and z value,
// where the y value is positive.
type Pos [3]int

// String converts the Pos to a string in the format (1,2,3) and returns it.
func (p Pos) String() string {
	return fmt.Sprintf("(%v,%v,%v)", p[0], p[1], p[2])
}

// X returns the X coordinate of the block position.
func (p Pos) X() int {
	return p[0]
}

// Y returns the Y coordinate of the block position.
func (p Pos) Y() int {
	return p[1]
}

// Z returns the Z coordinate of the block position.
func (p Pos) Z() int {
	return p[2]
}
