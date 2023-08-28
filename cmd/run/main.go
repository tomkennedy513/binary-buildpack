package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	"buildpack"
)

type Generator struct{}

func (f Generator) Generate(path string) (sbom.SBOM, error) {
	return sbom.Generate(path)
}

func main() {
	dependencyManager := postal.NewService(cargo.NewTransport())
	logEmitter := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	packit.Run(
		buildpack.Detect(),
		buildpack.Build(
			dependencyManager,
			Generator{},
			logEmitter,
			chronos.DefaultClock,
		),
	)
}
