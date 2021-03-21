package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type registry struct {
	UserVariable     string `yaml:"user_variable"`
	PasswordVariable string `yaml:"password_variable"`
}

func runExternalProgram(program string, args []string, env []string, stdin io.Reader) error {
	programPath, err := exec.LookPath(program)
	if err != nil {
		return err
	}
	env = append(env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	env = append(env, fmt.Sprintf("TMP=%s", os.Getenv("TMP")))
	env = append(env, fmt.Sprintf("TEMP=%s", os.Getenv("TEMP")))
	cmd := &exec.Cmd{
		Path:   programPath,
		Args:   append([]string{programPath}, args...),
		Env:    env,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  stdin,
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func buildVersion(
	version string,
	tags []string,
	date string,
	registries map[string]registry,
	push bool,
	githubToken string,
) error {
	log.Printf("Building images for version %s...", version)

	var newTags []string
	for _, tag := range tags {
		newTags = append(newTags, tag)
		newTags = append(newTags, fmt.Sprintf("%s-%s", tag, date))
	}

	for registryName, registry := range registries {
		for _, tag := range newTags {
			env := []string{
				fmt.Sprintf("CONTAINERSSH_VERSION=%s", version),
				fmt.Sprintf("CONTAINERSSH_TAG=%s", tag),
				fmt.Sprintf("GITHUB_TOKEN=%s", githubToken),
				fmt.Sprintf("REGISTRY=%s/", registryName),
			}

			if err := runExternalProgram(
				"docker-compose",
				[]string{
					"build",
				},
				env,
				nil,
			); err != nil {
				return fmt.Errorf(
					"build failed for version %s tag %s registry %s (%w)",
					version,
					tag,
					registryName,
					err,
				)
			}

			if err := runExternalProgram(
				"docker-compose",
				[]string{
					"up",
					"--abort-on-container-exit",
					"--exit-code-from=sut",
				},
				env,
				nil,
			); err != nil {
				return fmt.Errorf(
					"tests failed for version %s tag %s registry %s (%w)",
					version,
					tag,
					registryName,
					err,
				)
			}

			if err := runExternalProgram(
				"docker-compose",
				[]string{
					"down",
				},
				env,
				nil,
			); err != nil {
				return fmt.Errorf(
					"cleanup failed for version %s tag %s registry %s (%w)",
					version,
					tag,
					registryName,
					err,
				)
			}

			if push {
				username := os.Getenv(registry.UserVariable)
				if username == "" {
					return fmt.Errorf(
						"cannot push: no username set in the %s environment variable",
						registry.UserVariable,
					)
				}
				password := os.Getenv(registry.PasswordVariable)
				if password == "" {
					return fmt.Errorf(
						"cannot push: no password set in the %s environment variable",
						registry.PasswordVariable,
					)
				}
				if err := runExternalProgram(
					"docker",
					[]string{
						"login",
						registryName,
						"-u",
						os.Getenv(registry.UserVariable),
						"--password-stdin",
					},
					env,
					bytes.NewBuffer([]byte(password)),
				); err != nil {
					return fmt.Errorf(
						"push failed for version %s tag %s registry %s (%w)",
						version,
						tag,
						registryName,
						err,
					)
				}
				if err := runExternalProgram(
					"docker-compose",
					[]string{
						"push",
					},
					env,
					nil,
				); err != nil {
					return fmt.Errorf(
						"push failed for version %s tag %s registry %s (%w)",
						version,
						tag,
						registryName,
						err,
					)
				}
			}
		}
	}

	return nil
}

type config struct {
	Revision   string              `yaml:"revision"`
	Versions   map[string][]string `yaml:"versions"`
	Registries map[string]registry `yaml:"registries"`
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

	for version, tags := range conf.Versions {
		if err := buildVersion(version, tags, conf.Revision, conf.Registries, push, githubToken); err != nil {
			log.Fatal(err)
		}
	}
}
