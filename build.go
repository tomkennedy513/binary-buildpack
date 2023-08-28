package buildpack

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type DependencyService interface {
	Resolve(path, name, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

type SBOMGenerator interface {
	Generate(path string) (sbom.SBOM, error)
}

func Build(
	dependencyService DependencyService,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		dependency, err := dependencyService.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), BinaryName, "", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(packit.BuildpackPlanEntry{}, dependency, clock.Now())
		bom := dependencyService.GenerateBillOfMaterials(dependency)

		binaryLayer, err := context.Layers.Get(BinaryName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		launchMetadata := packit.LaunchMetadata{
			BOM: bom,
		}

		cachedChecksum, ok := binaryLayer.Metadata["dependency-checksum"].(string)
		if ok && cargo.Checksum(dependency.Checksum).MatchString(cachedChecksum) {
			logger.Process("Reusing cached layer %s", binaryLayer.Path)
			logger.Break()

			binaryLayer.Launch = true
			binaryLayer.Cache = true

			return packit.BuildResult{
				Layers: []packit.Layer{binaryLayer},
				Build:  packit.BuildMetadata{},
				Launch: launchMetadata,
			}, nil
		}

		logger.Process("Executing build process")

		binaryLayer, err = binaryLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}
		binaryLayer.Launch = true
		binaryLayer.Cache = true

		binDirectory := filepath.Join(binaryLayer.Path, "bin")
		err = os.MkdirAll(binDirectory, 0755)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Subprocess(fmt.Sprintf("Installing %s", BinaryName))

		duration, err := clock.Measure(func() error {
			return dependencyService.Deliver(dependency, context.CNBPath, binDirectory, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}
		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(binaryLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.Generate(binaryLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		binaryLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		binaryLayer.Metadata = map[string]interface{}{
			"dependency-checksum": dependency.Checksum,
		}

		return packit.BuildResult{
			Layers: []packit.Layer{binaryLayer},
			Build:  packit.BuildMetadata{},
			Launch: launchMetadata,
		}, nil
	}
}
