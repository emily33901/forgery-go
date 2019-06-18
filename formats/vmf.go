package formats

import (
	"errors"
	"os"
	"strconv"

	"github.com/emily33901/go-forgery/valve/world"
	"github.com/galaco/source-tools-common/entity"
	"github.com/galaco/vmf"
	"github.com/go-gl/mathgl/mgl32"
)

type Vmf struct {
	versionInfo  VersionInfo
	visGroups    VisGroups
	viewSettings ViewSettings
	world        world.World
	entities     entity.List
	cameras      Cameras
	cordon       Cordon
}

func (vmf *Vmf) VersionInfo() *VersionInfo {
	return &vmf.versionInfo
}

func (vmf *Vmf) Visgroups() *VisGroups {
	return &vmf.visGroups
}

func (vmf *Vmf) ViewSettings() *ViewSettings {
	return &vmf.viewSettings
}

func (vmf *Vmf) Worldspawn() *world.World {
	return &vmf.world
}

func (vmf *Vmf) Entities() *entity.List {
	return &vmf.entities
}

func (vmf *Vmf) Cameras() *Cameras {
	return &vmf.cameras
}

func (vmf *Vmf) Cordons() *Cordon {
	return &vmf.cordon
}

type VersionInfo struct {
	EditorVersion int
	EditorBuild   int
	MapVersion    int
	FormatVersion int
	Prefab        bool
}

func NewVersionInfo(version int, build int, mapRevision int, format int, isPrefab bool) *VersionInfo {
	return &VersionInfo{
		EditorVersion: version,
		EditorBuild:   build,
		MapVersion:    mapRevision,
		FormatVersion: format,
		Prefab:        isPrefab,
	}
}

type VisGroups struct {
}

type ViewSettings struct {
	SnapToGrid      bool
	ShowGrid        bool
	ShowLogicalGrid bool
	GridSpacing     int
	Show3DGrid      bool
}

type Cameras struct {
	ActiveCamera int
	CameraList   []Camera
}

func NewCameras(activeCameraIndex int, cameras []Camera) *Cameras {
	return &Cameras{
		ActiveCamera: activeCameraIndex,
		CameraList:   cameras,
	}
}

type Camera struct {
	Position mgl32.Vec3
	Look     mgl32.Vec3
}

func NewCamera(position mgl32.Vec3, look mgl32.Vec3) *Camera {
	return &Camera{
		Position: position,
		Look:     look,
	}
}

type Cordon struct {
	mins   mgl32.Vec3
	maxs   mgl32.Vec3
	active bool
}

func NewVmf(version *VersionInfo, visgroups *VisGroups, worldSpawn *world.World, entities *entity.List, cameras *Cameras) *Vmf {
	return &Vmf{
		versionInfo: *version,
		visGroups:   *visgroups,
		world:       *worldSpawn,
		entities:    *entities,
		cameras:     *cameras,
	}
}

// Public loader function to open and import a vmf file
// Will error out if the file is malformed or cannot be opened
func LoadVmf(filepath string) (*Vmf, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := vmf.NewReader(file)
	importable, err := reader.Read()

	if err != nil {
		return nil, err
	}

	// Create models for different vmf properties
	versionInfo, err := loadVersionInfo(&importable.VersionInfo)
	if err != nil || versionInfo == nil {
		return nil, err
	}
	visGroups, err := loadVisGroups(&importable.VisGroup)
	if err != nil || visGroups == nil {
		return nil, err
	}
	worldspawn, err := loadWorld(&importable.World)
	if err != nil || worldspawn == nil {
		return nil, err
	}

	cameras, err := loadCameras(&importable.Cameras)
	if err != nil || cameras == nil {
		return nil, err
	}

	entities := loadEntities(&importable.Entities)

	return NewVmf(versionInfo, visGroups, worldspawn, entities, cameras), nil
}

// loadVersionInfo creates a VersionInfo model
// from the versioninfo vmf block
func loadVersionInfo(root *vmf.Node) (*VersionInfo, error) {
	if root == nil {
		return nil, errors.New("missing versioninfo")
	}
	editorVersion, err := strconv.ParseInt(root.GetProperty("editorversion"), 10, 32)
	if err != nil {
		return nil, err
	}
	editorBuild, err := strconv.ParseInt(root.GetProperty("editorbuild"), 10, 32)
	if err != nil {
		return nil, err
	}
	mapVersion, err := strconv.ParseInt(root.GetProperty("mapversion"), 10, 32)
	if err != nil {
		return nil, err
	}
	formatVersion, err := strconv.ParseInt(root.GetProperty("formatversion"), 10, 32)
	if err != nil {
		return nil, err
	}
	prefab := false
	if root.GetProperty("prefab") != "0" {
		prefab = true
	}

	return NewVersionInfo(int(editorVersion), int(editorBuild), int(mapVersion), int(formatVersion), prefab), nil
}

// loadVisgroups loads all visgroup information from the
// visgroups block of a vmf
func loadVisGroups(root *vmf.Node) (*VisGroups, error) {
	return &VisGroups{}, nil
}

func loadWorld(root *vmf.Node) (*world.World, error) {
	solidNodes := root.GetChildrenByKey("solid")
	worldSpawn := entity.FromVmfNode(root)

	solids := make([]world.Solid, len(solidNodes))
	for idx, solidNode := range solidNodes {
		solid, err := loadSolid(&solidNode)
		if err != nil {
			return nil, err
		}
		solids[idx] = *solid
	}

	return world.NewWorld(&worldSpawn, solids), nil
}

// loadSolid takes a vmf node tree that represents a solid and turns
// it into a properly defind model structure for the solid with
// proper type definitions.
func loadSolid(node *vmf.Node) (*world.Solid, error) {
	id, err := strconv.ParseInt(node.GetProperty("id"), 10, 32)
	if err != nil {
		return world.NewSolid(-1, nil, nil), err
	}
	sideNodes := node.GetChildrenByKey("side")
	// Create sides for solid
	sides := make([]world.Side, len(sideNodes))
	for idx, sideNode := range sideNodes {
		var id int64
		var plane world.Plane
		var material string
		var u, v world.UVTransform
		var rotation, lmScale float64
		var smoothing bool

		id, err := strconv.ParseInt(sideNode.GetProperty("id"), 10, 32)
		if err != nil {
			return nil, err
		}
		plane = *world.NewPlaneFromString(sideNode.GetProperty("plane"))

		material = sideNode.GetProperty("material")

		u = *world.NewUVTransformFromString(sideNode.GetProperty("uaxis"))
		v = *world.NewUVTransformFromString(sideNode.GetProperty("vaxis"))

		rotation, err = strconv.ParseFloat(sideNode.GetProperty("rotation"), 32)
		if err != nil {
			return nil, err
		}
		lmScale, err = strconv.ParseFloat(sideNode.GetProperty("lightmapscale"), 32)
		if err != nil {
			return nil, err
		}
		smoothing, err = strconv.ParseBool(sideNode.GetProperty("smoothing_groups"))
		if err != nil {
			return nil, err
		}

		sides[idx] = *world.NewSide(int(id), plane, material, u, v, rotation, lmScale, smoothing)
	}

	return world.NewSolid(int(id), sides, nil), nil
}

// loadEntities creates models from the entity data block
// from a vmf
func loadEntities(node *vmf.Node) *entity.List {
	entities := entity.FromVmfNodeTree(*node)

	return &entities
}

// loadCameras creates cameras from the vmf camera list
func loadCameras(node *vmf.Node) (*Cameras, error) {
	activeCamProp := node.GetProperty("activecamera")
	activeCamIdx, _ := strconv.ParseInt(activeCamProp, 10, 32)

	cameras := make([]Camera, 0)

	cameraProps := node.GetChildrenByKey("camera")
	for _, camProp := range cameraProps {
		pos := camProp.GetProperty("position")
		look := camProp.GetProperty("look")

		cameras = append(cameras, *NewCamera(world.NewVec3FromString(pos), world.NewVec3FromString(look)))
	}

	return NewCameras(int(activeCamIdx), cameras), nil
}
