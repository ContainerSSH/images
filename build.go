package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v2"
)

type gitHubReleaseAsset struct {
	Url                string `json:"url"`
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []gitHubReleaseAsset `json:"assets"`
}

type githubReleaseResponse = []githubRelease

func downloadRelease(repo string, version string, githubToken string, assets map[string]string) error {
	release, err := getRelease(repo, version, githubToken)
	if err != nil {
		return err
	}
	for _, asset := range release.Assets {
		for targetAssetName, targetFile := range assets {
			if asset.Name == targetAssetName {
				if err := downloadAsset(asset.BrowserDownloadUrl, targetFile, githubToken); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func downloadAsset(url string, file string, githubToken string) error {
	fh, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create target file %s (%w)", file, err)
	}
	defer func() {
		_ = fh.Close()
	}()

	httpClient := &http.Client{}
	binaryRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for asset download %s (%w)", url, err)
	}
	if githubToken != "" {
		binaryRequest.Header.Add("authorization", "bearer "+githubToken)
	}
	binaryResponse, err := httpClient.Do(binaryRequest)
	if err != nil {
		return fmt.Errorf("failed to download asset %s (%w)", url, err)
	}
	defer func() {
		_ = binaryResponse.Body.Close()
	}()
	if binaryResponse.StatusCode != 200 {
		return fmt.Errorf(
			"invalid HTTP response code while downloading asset %s (%s)",
			url,
			binaryResponse.Status,
		)
	}

	if _, err := io.Copy(fh, binaryResponse.Body); err != nil {
		return fmt.Errorf("failed to download asset %s (%w)", file, err)
	}
	return nil
}

func getRelease(repo string, version string, githubToken string) (*githubRelease, error) {
	httpClient := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)
	jsonRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request (%w)", err)
	}
	if githubToken != "" {
		jsonRequest.Header.Add("authorization", "bearer "+githubToken)
	}

	jsonResponse, err := httpClient.Do(jsonRequest)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to download information about the latest go-swagger release (%w)",
			err,
		)
	}
	defer func() {
		_ = jsonResponse.Body.Close()
	}()
	if jsonResponse.StatusCode != 200 {
		return nil, fmt.Errorf(
			"invalid HTTP response code for release query (%s)",
			jsonResponse.Status,
		)
	}
	releaseResponse := &githubReleaseResponse{}
	err = json.NewDecoder(jsonResponse.Body).Decode(releaseResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode github release data (%w)", err)
	}

	prefixedVersion := fmt.Sprintf("v%s", version)
	for _, release := range *releaseResponse {
		if release.TagName == version || release.TagName == prefixedVersion {
			return &release, nil
		}
	}
	return nil, fmt.Errorf("version not found")
}

type sourceTarget struct {
	source string
	target string
}

type registry struct {
	UserVariable     string `yaml:"user_variable"`
	PasswordVariable string `yaml:"password_variable"`
}

func getFilesFromTarball(tarball string, fileMap []sourceTarget) error {
	fh, err := os.Open(tarball)
	if err != nil {
		return fmt.Errorf("failed to open file %s (%w)", tarball, err)
	}
	gzipReader, err := gzip.NewReader(fh)
	if err != nil {
		return fmt.Errorf("failed to initialize gzip stream for %s (%w)", tarball, err)
	}
	tarReader := tar.NewReader(gzipReader)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read header (%w)", err)
		}

		switch header.Typeflag {
		case tar.TypeReg:
			for _, sourceTarget := range fileMap {
				if sourceTarget.source == header.Name {
					outFile, err := os.Create(sourceTarget.target)
					if err != nil {
						return fmt.Errorf(
							"failed to create extraction target file %s (%w)",
							sourceTarget.target,
							err,
						)
					}
					if _, err := io.Copy(outFile, tarReader); err != nil {
						_ = outFile.Close()
						return fmt.Errorf(
							"failed to extract to target file %s (%w)",
							sourceTarget.target,
							err,
						)
					}
					_ = outFile.Close()
				}
			}

		}
	}
	return nil
}

func buildImage(ctx context.Context, cli *client.Client, directory string, tags []string, args map[string]*string) error {
	reader, writer := io.Pipe()
	defer func() {
		_ = writer.Close()
	}()
	defer func() {
		_ = reader.Close()
	}()

	go func() {
		if err := tarDirectory(directory, writer); err != nil {
			log.Fatalf("failed to tar build directory (%v)", err)
		}
		if err := writer.CloseWithError(io.EOF); err != nil {
			log.Fatalf("failed to close writer (%v)", err)
		}
	}()

	response, err := cli.ImageBuild(ctx, reader, types.ImageBuildOptions{
		Dockerfile: "/Dockerfile",
		Tags:      tags,
		BuildArgs: args,
	})
	if err != nil {
		return fmt.Errorf("failed to build image (%v)", err)
	}
	defer func() { _ = response.Body.Close() }()

	if _, err := io.Copy(os.Stdout, response.Body); err != nil {
		return fmt.Errorf("failed to stream build output (%w)", err)
	}
	return nil
}

func tarDirectory(src string, writer io.Writer) error {
	gzipWriter := gzip.NewWriter(writer)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		fileName := path[len(src)+1:]
		header, err := tar.FileInfoHeader(info, fileName)
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s (%w)", fileName, err)
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s (%w)", fileName, err)
		}
		fh, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file %s (%w)", path, err)
		}
		defer func() {
			_ = fh.Close()
		}()
		if _, err := io.Copy(tarWriter, fh); err != nil {
			return fmt.Errorf("failed to read file contents %s (%w)", path, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to add files to tarball (%w)", err)
	}
	_ = tarWriter.Close()
	_ = gzipWriter.Close()
	return nil
}

func buildVersion(
	ctx context.Context,
	cli *client.Client,
	version string,
	tags []string,
	date string,
	registries map[string]registry,
	push bool,
	githubToken string,
) error {
	log.Printf("Downloading assets for version %s...", version)
	tempDir := os.TempDir()
	tarball := path.Join(tempDir, "containerssh.tar.gz")
	assets := map[string]string{
		fmt.Sprintf("containerssh_%s_linux_amd64.tar.gz", version): tarball,
	}
	defer func() {
		_ = os.Remove(tarball)
	}()
	if err := downloadRelease("containerssh/containerssh", version, githubToken, assets); err != nil {
		return err
	}
	if err := getFilesFromTarball(
		tarball, []sourceTarget{
			{"containerssh", "containerssh/containerssh"},
			{"LICENSE.md", "containerssh/LICENSE.md"},
			{"NOTICE.md", "containerssh/NOTICE.md"},
		},
	); err != nil {
		return err
	}
	var newTags []string
	for _, tag := range tags {
		for registryName := range registries {
			newTags = append(newTags, fmt.Sprintf("%s/containerssh/containerssh:%s", registryName, tag))
			if tag != "latest" {
				newTags = append(newTags, fmt.Sprintf("%s/containerssh/containerssh:%s-%s", registryName, tag, date))
			}
		}
	}
	log.Printf("Building image for version %s...", version)
	if err := buildImage(
		ctx, cli, "containerssh", newTags, map[string]*string{},
	); err != nil {
		return err
	}

	if push {
		if err := pushImage(ctx, cli, newTags, registries); err != nil {
			return err
		}
	}
	return nil
}

func pushImage(ctx context.Context, cli *client.Client, tags []string, registries map[string]registry) error {
	for _, tag := range tags {
		log.Printf("Pushing image %s...", tag)
		registry := strings.Split(tag, "/")
		authArs := map[string]interface{}{
			"username":      os.Getenv(registries[registry[0]].UserVariable),
			"password":      os.Getenv(registries[registry[0]].PasswordVariable),
			"email":         "",
			"serveraddress": registry[0],
		}
		encodedAuth, err := json.Marshal(authArs)
		if err != nil {
			return fmt.Errorf("failed to encode credentials (%w)", err)
		}
		reader, err := cli.ImagePush(ctx, tag, types.ImagePushOptions{
			RegistryAuth: base64.StdEncoding.EncodeToString(encodedAuth),
		})
		if err != nil {
			return fmt.Errorf("image push for %s failed (%w)", tag, err)
		}
		if _, err := io.Copy(os.Stdout, reader); err != nil {
			_ = reader.Close()
			return fmt.Errorf("image push for %s failed (%w)", tag, err)
		}
		_ = reader.Close()
	}
	log.Printf("Push complete.")
	return nil
}

type config struct {
	Revision   string              `yaml:"revision"`
	Versions   map[string][]string `yaml:"versions"`
	Registries map[string]registry `yaml:"registries"`
}

func getDockerClient(ctx context.Context) (*client.Client, error) {
	log.Printf("Pushing images...")
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return cli, fmt.Errorf("failed to set up Docker client (%w)", err)
	}
	cli.NegotiateAPIVersion(ctx)

	return cli, nil
}

func main() {
	push := false
	if len(os.Args) == 2 && os.Args[1] == "--push" {
		push = true
	}
	githubToken := os.Getenv("GITHUB_TOKEN")

	fh, err := os.Open("build.yaml")
	if err != nil {
		log.Fatal(err)
	}
	data, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	conf := &config{}
	if err := yaml.Unmarshal(data, conf); err != nil {
		log.Fatal(err)
	}
	ctx := context.TODO()
	cli, err := getDockerClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for version, tags := range conf.Versions {
		if err := buildVersion(ctx, cli, version, tags, conf.Revision, conf.Registries, push, githubToken); err != nil {
			log.Fatal(err)
		}
	}
}
