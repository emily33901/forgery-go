package valve

import (
	"io/ioutil"
	"strings"

	"github.com/emily33901/lambda-core/lib/vpk"

	"github.com/emily33901/lambda-core/core/filesystem"
	"github.com/emily33901/lambda-core/core/logger"
	"github.com/emily33901/lambda-core/core/resource"
	"github.com/emily33901/lambda-core/lib/gameinfo"
)

// NewFileSystem builds a new filesystem from a game directory root.
// It loads a gameinfo.txt and attempts to find listed resourced
// in it.
func NewFileSystem(gameDir string) filesystem.IFileSystem {
	gameInfo, err := gameinfo.LoadConfig(gameDir)
	if err != nil {
		logger.Panic(err)
	}

	// Register GameInfo.txt referenced resource paths
	// Filesystem module needs to know about all the possible resource
	// locations it can search.
	fs := filesystem.CreateFilesystemFromGameInfoDefinitions(gameDir, gameInfo)

	// Make sure to also load the platform dir
	fs.RegisterLocalDirectory(gameDir + "/../platform")

	// Now try and load all of the vpks that are in those directories
	for _, x := range fs.EnumerateResourcePaths() {
		files, err := ioutil.ReadDir(x)

		if err != nil {
			// panic(err)
			continue
		}

		for _, f := range files {
			if strings.HasSuffix(f.Name(), "_dir.vpk") {
				nameNoSuffix := x + "/" + f.Name()[:len(f.Name())-8]
				v, err := vpk.OpenVPK(nameNoSuffix)

				if err != nil {
					panic(err)
				}

				fs.RegisterVpk(nameNoSuffix, v)
			}
		}

		// TODO: we also need to load all the vpks that arent part of a dir pack
	}

	// Explicity define fallbacks for missing resources
	// Defaults are defined, but if HL2 assets are not readable, then
	// the default may not be readable

	// TODO: these need to be set in settings
	resource.Manager().SetErrorModelName("models/props/de_dust/du_antenna_A.mdl")
	resource.Manager().SetErrorTextureName("materials/error.vtf")

	return fs.(*filesystem.FileSystem)
}

func DumpAllKnownMaterials(fs filesystem.IFileSystem) {
	for _, f := range fs.AllPaths() {
		if strings.Index(f, "materials") != -1 {
			if strings.Index(f, "error") != -1 {
				logger.Notice(f)
			}
		}
	}
}
