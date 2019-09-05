package convert

import (
	"fmt"
	"math/rand"

	"github.com/emily33901/go-forgery/valve/world"
	"github.com/emily33901/gosigl"
	"github.com/emily33901/lambda-core/core/material"
	lambdaMesh "github.com/emily33901/lambda-core/core/mesh"
	lambdaModel "github.com/emily33901/lambda-core/core/model"
)

func SolidToModel(solid *world.Solid) *lambdaModel.Model {
	meshes := make([]lambdaMesh.IMesh, 0)

	solidColor := []float32{rand.Float32()/4 + 0.1, rand.Float32()/2 + 0.5, rand.Float32()/4 + 0.75, 1.0}

	for idx := range solid.Sides {
		mesh := SideToMesh(&solid.Sides[idx])
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
	verts := make([]float32, 0)
	{
		// a plane represents 3 vertices- bottom-left, top-left and top-right
		// Triangle 1
		verts = append(verts, float32(side.Plane[0].X()), float32(side.Plane[0].Y()), float32(side.Plane[0].Z()))
		verts = append(verts, float32(side.Plane[1].X()), float32(side.Plane[1].Y()), float32(side.Plane[1].Z()))
		verts = append(verts, float32(side.Plane[2].X()), float32(side.Plane[2].Y()), float32(side.Plane[2].Z()))
		// Triangle 2
		verts = append(verts, float32(side.Plane[0].X()), float32(side.Plane[0].Y()), float32(side.Plane[0].Z()))
		verts = append(verts, float32(side.Plane[2].X()), float32(side.Plane[2].Y()), float32(side.Plane[2].Z()))

		// Compute new vertex
		vert4 := side.Plane[2].Sub(side.Plane[1].Sub(side.Plane[0]))
		verts = append(verts, float32(vert4.X()), float32(vert4.Y()), float32(vert4.Z()))

		mesh.AddVertex(verts...)
	}

	// Normals
	normals := make([]float32, 0)
	{
		normal := (side.Plane[1].Sub(side.Plane[0])).Cross(side.Plane[2].Sub(side.Plane[0]))
		normals = append(normals, float32(normal.X()), float32(normal.Y()), float32(normal.Z()))
		normals = append(normals, float32(normal.X()), float32(normal.Y()), float32(normal.Z()))
		normals = append(normals, float32(normal.X()), float32(normal.Y()), float32(normal.Z()))
		normals = append(normals, float32(normal.X()), float32(normal.Y()), float32(normal.Z()))
		normals = append(normals, float32(normal.X()), float32(normal.Y()), float32(normal.Z()))
		normals = append(normals, float32(normal.X()), float32(normal.Y()), float32(normal.Z()))

		mesh.AddNormal(normals...)
	}

	// Texture coordinates
	{
		for i := 0; i < len(verts); i += 3 {
			// @TODO width & height must be known
			mesh.AddUV(uvForVertex(verts[i:i+3], &side.UAxis, &side.VAxis, 128, 128)...)
		}
	}

	// Tangents
	mesh.GenerateTangents()

	gosigl.FinishMesh()

	return mesh
}

func uvForVertex(vertex []float32, u *world.UVTransform, v *world.UVTransform, width int, height int) (uvs []float32) {
	cu := ((float32(u.Transform[0]) * vertex[0]) +
		(float32(u.Transform[1]) * vertex[1]) +
		(float32(u.Transform[2]) * vertex[2])) / float32(u.Scale) / float32(width)

	cv := ((float32(v.Transform[0]) * vertex[0]) +
		(float32(v.Transform[1]) * vertex[1]) +
		(float32(v.Transform[2]) * vertex[2])) / float32(v.Scale) / float32(height)

	return []float32{cu, cv}
}
