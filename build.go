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
	OrganisationVariable     string `yaml:"organisation_variable,omitempty"`
}

func runExternalProgram(
	program string,
	args []string,
	env []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) error {
	_, _ = stdout.Write([]byte(fmt.Sprintf("\033[0;32m⚙ Running %s...\u001B[0m\n", program)))
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
		Stdout: stdout,
		Stderr: stderr,
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

func writeOutput(
	version string,
	registry string,
	tag string,
	stdout *bytes.Buffer,
	err error,
) {
	output := ""
	prefix := "\033[0;32m✅ "
	if err != nil {
		prefix = "\033[0;31m❌ "
	}
	output += fmt.Sprintf(
		"::group::%sversion=%s registry=%s tag=%s\n",
		prefix,
		version,
		registry,
		tag,
	)
	output += stdout.String()
	if err != nil {
		output += fmt.Sprintf("\033[0;31m%s\033[0m\n", err.Error())
	}
	output += "::endgroup::\n"
	if _, err := os.Stdout.Write([]byte(output)); err != nil {
		panic(err)
	}
}

func buildVersion(
	version string,
	tags []string,
	date string,
	registries map[string]registry,
	push bool,
	githubToken string,
) error {
	var newTags []string
	for _, tag := range tags {
		newTags = append(newTags, tag)
		newTags = append(newTags, fmt.Sprintf("%s-%s", tag, date))
	}

	for registryName, registry := range registries {
		for _, tag := range newTags {
			stdout := &bytes.Buffer{}
			env := []string{
				fmt.Sprintf("CONTAINERSSH_VERSION=%s", version),
				fmt.Sprintf("CONTAINERSSH_TAG=%s", tag),
				fmt.Sprintf("GITHUB_TOKEN=%s", githubToken),
			}

			registryPrefix := fmt.Sprintf("%s/containerssh", registryName)
			if registry.OrganisationVariable != "" {
				organisation := os.Getenv(registry.OrganisationVariable)
				if organisation == "" {
					return fmt.Errorf(
						"cannot push: no organisation set in the %s environment variable",
						registry.OrganisationVariable,
					)
				}
				registryPrefix = fmt.Sprintf("%s/%s/containerssh", registryName, organisation)
			}
			env = append(env, fmt.Sprintf("REGISTRY=%s/", registryPrefix))

			if err := runExternalProgram(
				"docker",
				[]string{
					"compose",
					"build",
				},
				env,
				nil,
				stdout,
				stdout,
			); err != nil {
				err := fmt.Errorf(
					"build failed for version %s registry %s tag %s (%w)",
					version,
					registryPrefix,
					tag,
					err,
				)
				writeOutput(version, registryPrefix, tag, stdout, err)
				return err
			}

			if err := runExternalProgram(
				"docker",
				[]string{
					"compose",
					"up",
					"--abort-on-container-exit",
					"--exit-code-from=sut",
				},
				env,
				nil,
				stdout,
				stdout,
			); err != nil {
				err := fmt.Errorf(
					"tests failed for version %s registry %s tag %s (%w)",
					version,
					registryPrefix,
					tag,
					err,
				)
				writeOutput(version, registryPrefix, tag, stdout, err)
				return err
			}

			if err := runExternalProgram(
				"docker",
				[]string{
					"compose",
					"down",
				},
				env,
				nil,
				stdout, stdout,
			); err != nil {
				err := fmt.Errorf(
					"cleanup failed for version %s registry %s tag %s (%w)",
					version,
					registryPrefix,
					tag,
					err,
				)
				writeOutput(version, registryPrefix, tag, stdout, err)
				return err
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
					stdout,
					stdout,
				); err != nil {
					err := fmt.Errorf(
						"push failed for version %s tag %s registry %s (%w)",
						version,
						tag,
						registryPrefix,
						err,
					)
					writeOutput(version, registryPrefix, tag, stdout, err)
					return err
				}
				if err := runExternalProgram(
					"docker",
					[]string{
						"buildx",
						"build",
						"--push",
						"--platform", "linux/amd64,linux/arm64",
						"--build-arg", fmt.Sprintf("CONTAINERSSH_VERSION=%s", version),
						"--build-arg", fmt.Sprintf("CONTAINERSSH_TAG=%s", tag),
						"-t", fmt.Sprintf("%s/containerssh:%s", registryPrefix, tag),
						"containerssh",
					},
					env,
					nil,
					stdout,
					stdout,
				); err != nil {
					err := fmt.Errorf(
						"push failed for version %s tag %s registry %s (%w)",
						version,
						tag,
						registryPrefix,
						err,
					)
					writeOutput(version, registryPrefix, tag, stdout, err)
					return err
				}
				if err := runExternalProgram(
					"docker",
					[]string{
						"buildx",
						"build",
						"--push",
						"--platform", "linux/amd64,linux/arm64",
						"--build-arg", fmt.Sprintf("CONTAINERSSH_VERSION=%s", version),
						"--build-arg", fmt.Sprintf("CONTAINERSSH_TAG=%s", tag),
						"-t", fmt.Sprintf("%s/containerssh-test-authconfig:%s", registryPrefix, tag),
						"containerssh-test-authconfig",
					},
					env,
					nil,
					stdout,
					stdout,
				); err != nil {
					err := fmt.Errorf(
						"push failed for version %s tag %s registry %s (%w)",
						version,
						tag,
						registryPrefix,
						err,
					)
					writeOutput(version, registryPrefix, tag, stdout, err)
					return err
				}

			}
			writeOutput(version, registryPrefix, tag, stdout, nil)
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
