package convert

import (
	"fmt"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/emily33901/go-forgery/valve/world"
	"github.com/emily33901/lambda-core/core/material"
	lambdaMesh "github.com/emily33901/lambda-core/core/mesh"
	lambdaModel "github.com/emily33901/lambda-core/core/model"
	"github.com/emily33901/lambda-core/core/resource"
)

func SolidToModel(solid *world.Solid) *lambdaModel.Model {
	meshes := make([]lambdaMesh.IMesh, 0)

	solidColor := []float32{rand.Float32()/4 + 0.1, rand.Float32()/2 + 0.5, rand.Float32()/4 + 0.75, 1.0}

	for idx := range solid.Sides {
		mesh := SideToMesh(&solid.Sides[idx])

		// Color for each vertex
		mesh.AddColor(solidColor...)
		mesh.AddColor(solidColor...)
		mesh.AddColor(solidColor...)
		mesh.AddColor(solidColor...)
		mesh.AddColor(solidColor...)
		mesh.AddColor(solidColor...)
		mesh.AddColor(solidColor...)
		meshes = append(meshes, mesh)
	}

	return lambdaModel.NewModel(fmt.Sprintf("solid_%d", solid.Id), meshes...)
}

func SideToMesh(side *world.Side) *lambdaMesh.Mesh {
	mesh := lambdaMesh.NewMesh()

	// Material
	mesh.SetMaterial(material.NewMaterial(side.Material))

	// Vertices
	verts := make([]mgl32.Vec3, 0)
	{
		// a plane represents 3 vertices- bottom-left, top-left and top-right
		// Triangle 1
		verts = append(verts, side.Plane[0])
		verts = append(verts, side.Plane[1])
		verts = append(verts, side.Plane[2])

		// Triangle 2
		verts = append(verts, side.Plane[0])
		verts = append(verts, side.Plane[2])

		// Compute new vertex
		vert4 := side.Plane[2].Sub(side.Plane[1].Sub(side.Plane[0]))
		verts = append(verts, vert4)

		mesh.AddVertex(verts...)
	}

	// Normals
	normals := make([]mgl32.Vec3, 0)
	{
		normal := (side.Plane[1].Sub(side.Plane[0])).Cross(side.Plane[2].Sub(side.Plane[0]))
		normals = append(normals, normal)
		normals = append(normals, normal)
		normals = append(normals, normal)
		normals = append(normals, normal)
		normals = append(normals, normal)
		normals = append(normals, normal)

		mesh.AddNormal(normals...)
	}

	// Texture coordinates
	{
		for i := range verts {
			// TODO width & height must be known
			mat := resource.Manager().Material(side.Material)

			width, height := 128, 128

			if mat != nil {
				width = mat.Width()
				height = mat.Height()
			}

			mesh.AddUV(uvForVertex(verts[i], &side.UAxis, &side.VAxis, width, height))
		}
	}

	// Tangents
	mesh.GenerateTangents()

	return mesh
}

func uvForVertex(vertex mgl32.Vec3, u *world.UVTransform, v *world.UVTransform, width int, height int) (uvs mgl32.Vec2) {
	cu := ((float32(u.Transform[0]) * vertex[0]) +
		(float32(u.Transform[1]) * vertex[1]) +
		(float32(u.Transform[2]) * vertex[2])) / float32(u.Scale) / float32(width)

	cv := ((float32(v.Transform[0]) * vertex[0]) +
		(float32(v.Transform[1]) * vertex[1]) +
		(float32(v.Transform[2]) * vertex[2])) / float32(v.Scale) / float32(height)

	return mgl32.Vec2{cu, cv}
}
