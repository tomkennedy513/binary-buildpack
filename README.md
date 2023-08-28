# Binary Buildpack Sample

## How to use this buildpack

### Initial Setup
1. Fork this repo 
2. Install prerequisites
   1. [pack](https://buildpacks.io/docs/tools/pack/)
   2. [jam](https://github.com/paketo-buildpacks/jam)
3. Update `buildpack.toml`
   1. change the `buildpack.id` field to a new id (ex. `my-buildpack/ytt`)'\
   2. change the `buildpack.homepage` field to be your forked git repo url
   3. update `metadata.dependencies` to contain the binary you would like to be installed
4. Update `constants.go`
   1. change `BinaryName` to be the same as the `id` of the binary dependency provided to `metadata.dependencies`

### Packaging
1. make sure dependency is at proper version
   1. if not, follow [Updating binary dependency](#updating-binary-dependency)
2. run `./scripts/package.sh --image <IMAGE_TO_PUSH> --version <BUILDPACK_VERSION> [--publish]` where:
   1. `IMAGE_TO_PUSH` is image to write the packaged buildpack to 
   2. `BUILDPACK_VERSION` is the version of the buildpack to publish
   3. `--publish` is whether to push the packaged buildpack to the remote registry instead of the docker daemon

### Updating binary dependency 
1. In `buildpack.toml`, update the `metadata.dependencies` entry to contain the info for the binary you would like to 
pull in. If type is a tar.gz, you can add `strip-components=` if needed. 