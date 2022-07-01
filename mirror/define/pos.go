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

// 为和国际版MC保持统一，世界范围被定义为 -64~319 ,现在网易的高度也是 -64 ~ 319 了，所以省事了不少
var WorldRange = Range{-64, 319}

// CubePos holds the position of a block. The position is represented of an array with an x, y and z value,
// where the y value is positive.
type CubePos [3]int

func (p CubePos) OutOfYBounds() bool {
	y := p[1]
	return y > WorldRange[1] || y < WorldRange[0]
}

// String converts the Pos to a string in the format (1,2,3) and returns it.
func (p CubePos) String() string {
	return fmt.Sprintf("(%v,%v,%v)", p[0], p[1], p[2])
}

func (p CubePos) Sub(po CubePos) (offset CubePos) {
	offset[0] = p[0] - po[0]
	offset[1] = p[1] - po[1]
	offset[2] = p[2] - po[2]
	return offset
}

func (p CubePos) Add(po CubePos) (offset CubePos) {
	offset[0] = p[0] + po[0]
	offset[1] = p[1] + po[1]
	offset[2] = p[2] + po[2]
	return offset
}

// X returns the X coordinate of the block position.
func (p CubePos) X() int {
	return p[0]
}

// Y returns the Y coordinate of the block position.
func (p CubePos) Y() int {
	return p[1]
}

// Z returns the Z coordinate of the block position.
func (p CubePos) Z() int {
	return p[2]
}

func GetPosFromNBT(nbt map[string]interface{}) (x, y, z int, success bool) {
	if ax, hasK := nbt["x"]; hasK {
		if cx, success := ax.(int32); success {
			x = int(cx)
		} else {
			return 0, 0, 0, false
		}
	} else {
		return 0, 0, 0, false
	}
	if ay, hasK := nbt["y"]; hasK {
		if cy, success := ay.(int32); success {
			y = int(cy)
		} else {
			return 0, 0, 0, false
		}
	} else {
		return 0, 0, 0, false
	}
	if az, hasK := nbt["z"]; hasK {
		if cz, success := az.(int32); success {
			z = int(cz)
		} else {
			return 0, 0, 0, false
		}
	} else {
		return 0, 0, 0, false
	}
	return x, y, z, true
}

func GetCubePosFromNBT(nbt map[string]interface{}) (p CubePos, success bool) {
	if x, y, z, success := GetPosFromNBT(nbt); success {
		return CubePos{x, y, z}, true
	} else {
		return CubePos{0, 0, 0}, false
	}
}
