[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Container Image Repository</h1>

This repository contains the scripts that build the ContainerSSH container images.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer repository. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## How this repository works

This repository contains a build script in Go called `build.go`. It can be invoked by running `go run build.go`. This script will read [build.yaml](build.yaml) and build the container image based on that revision. It uses the GitHub API to download release artifacts, so it may need the `GITHUB_TOKEN` environment variable set. The optional `--push` flag can be set to push the images to the corresponding registries.

Under the hood the build uses [`docker compose`](https://docs.docker.com/compose/) to build, test, and push the images. The build steps can be performed manually.

Before you begin you must set several environment variables. These are the following:

| Variable | Required | Description|
|----------|----------|------------|
| `CONTAINERSSH_VERSION` | Yes | Sets the ContainerSSH version to download. |
| `CONTAINERSSH_TAG_VERSION` | Yes | Sets the container image tag suffix to create. (See the [Versioning section](#versioning) below.) | 
| `REGISTRY` | No | Sets the registry prefix to push to. For example, `quay.io/`. Defaults to the Docker hub. |
| `GITHUB_TOKEN` | No | Sets the GitHub access token to work around anonymous rate limits. |
| `SOURCE_REPO` | No | Sets the source URL for downloads. Defaults to `https://github.com/ContainerSSH/ContainerSSH`. |

For example, on Linux/MacOS:

```bash
CONTAINERSSH_VERSION="v0.5.2"
CONTAINERSSH_TAG="v0.5.2"
```

On Windows/PowerShell:

```ps1
$env:CONTAINERSSH_VERSION="v0.5.2"
$env:CONTAINERSSH_TAG="v0.5.2"
```

### Build

The build step requires build arguments to function. At the very least it should contain the `CONTAINERSSH_VERSION` variable so that the build knows which ContainerSSH release to download.

Optionally, you can also specify a `GITHUB_TOKEN` to work around GitHub rate limits and `SOURCE_REPO` to point the build to a different source URL.

```bash
docker compose build
``` 

### Test

Testing is done via a container called `sut`. This container will wait for ContainerSSH to come up and then run a simple SSH connection to it to test that it works correctly. This is not a comprehensive test, but checks if the image build was successful.

```
docker compose up --abort-on-container-exit --exit-code-from=sut
```

### Clean up after test

```
docker compose down
```

### Push

Finally, pushing container images can also be done from `docker compose`. After a `docker login` command this can be simply done using the following command:

```
docker compose push
```

## Versioning

ContainerSSH container images are versioned independently of ContainerSSH. This allows for more frequent rebuilds of the image than we have ContainerSSH releases. This is important because we want our users to have frequent security updates. Therefore, the build script creates multiple tags for the image.

Let's take version 0.4.0, for example. Let's say the [build.yaml](build.yaml) contains the following configuration:

```yaml
revision: 20200318
versions:
  0.4.0:
   - latest
   - 0.4
   - 0.4.0
```

In this case the build script would create the following tags:

- latest
- 0.4
- 0.4-20200318
- 0.4.0
- 0.4.0-20200318

Users can safely rely on the tag with the ContainerSSH version, or can specify a very specific build version should they need to pin to an immutable version.