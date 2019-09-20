package world

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

type Solid struct {
	Id     int
	Sides  []Side
	Editor *Editor
}

type Side struct {
	Id              int
	Plane           Plane
	Material        string
	UAxis           UVTransform
	VAxis           UVTransform
	Rotation        float32
	LightmapScale   float32
	SmoothingGroups bool
}

type UVTransform struct {
	Transform mgl32.Vec4
	Scale     float32
}

type Editor struct {
	color             mgl32.Vec3
	visgroupShown     bool
	visGroupAutoShown bool

	logicalPos mgl32.Vec2 // only exists on brush entities?
}

type Plane [3]mgl32.Vec3

func NewSolid(id int, sides []Side, editor *Editor) *Solid {
	return &Solid{
		Id:     id,
		Sides:  sides,
		Editor: editor,
	}
}

func NewSide(id int, plane Plane, material string, uAxis UVTransform, vAxis UVTransform, rotation float32, lightmapScale float32, smoothingGroups bool) *Side {
	return &Side{
		Id:              id,
		Plane:           plane,
		Material:        material,
		UAxis:           uAxis,
		VAxis:           vAxis,
		Rotation:        rotation,
		LightmapScale:   lightmapScale,
		SmoothingGroups: smoothingGroups,
	}
}

func NewEditor(color mgl32.Vec3, visgroupShown bool, visgroupAutoShown bool) *Editor {
	return &Editor{
		color:             color,
		visgroupShown:     visgroupShown,
		visGroupAutoShown: visgroupAutoShown,
	}
}

func NewPlane(a mgl32.Vec3, b mgl32.Vec3, c mgl32.Vec3) *Plane {
	p := Plane([3]mgl32.Vec3{a, b, c})
	return &p
}

func NewPlaneFromString(marshalled string) *Plane {
	var v1, v2, v3 = float32(0), float32(0), float32(0)
	var v4, v5, v6 = float32(0), float32(0), float32(0)
	var v7, v8, v9 = float32(0), float32(0), float32(0)
	fmt.Sscanf(marshalled, "(%f %f %f) (%f %f %f) (%f %f %f)", &v1, &v2, &v3, &v4, &v5, &v6, &v7, &v8, &v9)

	return NewPlane(
		mgl32.Vec3{v1, v2, v3},
		mgl32.Vec3{v4, v5, v6},
		mgl32.Vec3{v7, v8, v9})
}

func NewUVTransform(transform mgl32.Vec4, scale float32) *UVTransform {
	return &UVTransform{
		Transform: transform,
		Scale:     scale,
	}
}

func NewUVTransformFromString(marshalled string) *UVTransform {
	var v1, v2, v3, v4 = float32(0), float32(0), float32(0), float32(0)
	var scale = float32(0)
	fmt.Sscanf(marshalled, "[%f %f %f %f] %f", &v1, &v2, &v3, &v4, &scale)
	return NewUVTransform(mgl32.Vec4{v1, v2, v3, v4}, scale)
}
