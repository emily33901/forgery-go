package world

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

func NewVec3FromString(marshalled string) mgl32.Vec3 {
	var x, y, z float32
	fmt.Sscanf(marshalled, "[%f %f %f]", &x, &y, &z)

	return mgl32.Vec3{x, y, z}
}
