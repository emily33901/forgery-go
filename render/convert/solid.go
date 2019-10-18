package convert

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/emily33901/go-forgery/valve/world"
	"github.com/emily33901/lambda-core/core/filesystem"
	materialoader "github.com/emily33901/lambda-core/core/loader/material"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/emily33901/lambda-core/core/material"
	lambdaMesh "github.com/emily33901/lambda-core/core/mesh"
	lambdaModel "github.com/emily33901/lambda-core/core/model"
	"github.com/golang-source-engine/vmt"
)

func SolidToModel(solid *world.Solid, fs filesystem.IFileSystem) *lambdaModel.Model {
	meshes := make([]lambdaMesh.IMesh, 0)

	for idx := range solid.Sides {
		mesh := sideToMesh(&solid.Sides[idx], fs)

		mesh.SetMeta("solid", solid.Id)

		// Color for each vertex
		mesh.AddColor(solid.Editor.Color[0], solid.Editor.Color[1], solid.Editor.Color[2], 1.0)
		mesh.AddColor(solid.Editor.Color[0], solid.Editor.Color[1], solid.Editor.Color[2], 1.0)
		mesh.AddColor(solid.Editor.Color[0], solid.Editor.Color[1], solid.Editor.Color[2], 1.0)
		mesh.AddColor(solid.Editor.Color[0], solid.Editor.Color[1], solid.Editor.Color[2], 1.0)
		mesh.AddColor(solid.Editor.Color[0], solid.Editor.Color[1], solid.Editor.Color[2], 1.0)
		mesh.AddColor(solid.Editor.Color[0], solid.Editor.Color[1], solid.Editor.Color[2], 1.0)
		meshes = append(meshes, mesh)
	}

	return lambdaModel.NewModel(fmt.Sprintf("solid_%d", solid.Id), meshes...)
}

func sideToMesh(side *world.Side, fs filesystem.IFileSystem) *lambdaMesh.Mesh {
	mesh := lambdaMesh.NewMesh()

	// Material
	mesh.SetMaterial(material.NewMaterial(side.Material, vmt.NewProperties()))
	mesh.SetMeta("side", side.Id)

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
			mat := materialoader.LoadSingleMaterial(side.Material, fs)

			width, height := 128, 128

			if mat != nil {
				mesh.SetMaterial(mat)
				width = mat.Width()
				height = mat.Height()
			} else {
				// TODO use the event manager to catch when materials are loaded properly
				// and then update the uvs there and rebind them
				logger.Notice("mat == nil so uvs will be wrong")
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
