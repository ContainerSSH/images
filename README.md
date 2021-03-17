[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.github.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Container Image Repository</h1>

This repository contains the scripts that build the ContainerSSH container images.

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer repository. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## How this repository works

This repository contains a build script in Go called `build.go`. It can be invoked by running `go run build.go`. This script will read [build.yaml](build.yaml) and build the container image based on that revision. It uses the GitHub API to download release artifacts, so it may need the `GITHUB_TOKEN` environment variable set.