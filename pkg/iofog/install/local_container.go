/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

// LocalContainer struct to encapsulate utilities around docker
type LocalContainer struct {
	client *client.Client
}

// ExecResult contains the output of a command ran into docker exec
type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

type port struct {
	Host      string
	Container *LocalContainerPort
}

type Credentials struct {
	User     string
	Password string
}

type LocalContainerConfig struct {
	Host          string
	Ports         []port
	ContainerName string
	Image         string
	DefaultImage  string
	Privileged    bool
	Binds         []string
	NetworkMode   string
	Credentials   Credentials
}

type LocalControllerConfig struct {
	ContainerMap map[string]*LocalContainerConfig
}

type LocalContainerPort struct {
	Protocol string
	Port     string
}

type LocalAgentConfig struct {
	LocalContainerConfig
	Name string
}

func GetLocalContainerName(t string, isSystem bool) string {
	names := map[string]string{
		"controller": sanitizeContainerName("iofog-controller"),
		"agent":      sanitizeContainerName("iofog-agent"),
	}
	name, ok := names[t]
	if !ok {
		return ""
	}
	if isSystem {
		return name + "-system"
	}
	return name
}

func sanitizeContainerName(name string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_.-]")
	return r.ReplaceAllString(name, "-")
}

// NewAgentConfig generates a static agent config
func NewLocalAgentConfig(name, image string, ctrlConfig *LocalContainerConfig, credentials Credentials, isSystem bool) *LocalAgentConfig {
	if image == "" {
		image = util.GetAgentImage()
	}

	return &LocalAgentConfig{
		LocalContainerConfig: LocalContainerConfig{
			Host: "0.0.0.0",
			Ports: []port{
				{Host: "54321", Container: &LocalContainerPort{Protocol: "tcp", Port: "54321"}},
				{Host: "8081", Container: &LocalContainerPort{Protocol: "tcp", Port: "22"}},
			},
			ContainerName: GetLocalContainerName("agent", isSystem),
			Image:         image,
			Privileged:    true,
			Binds:         []string{"/var/run/docker.sock:/var/run/docker.sock:rw"},
			NetworkMode:   "host",
			Credentials:   credentials,
		},
		Name: name,
	}
}

// NewLocalControllerConfig generats a static controller config
func NewLocalControllerConfig(image string, credentials Credentials) *LocalContainerConfig {
	if image == "" {
		image = util.GetControllerImage()
	}
	return &LocalContainerConfig{
		Host: "0.0.0.0",
		Ports: []port{
			{Host: iofog.ControllerPortString, Container: &LocalContainerPort{Port: iofog.ControllerPortString, Protocol: "tcp"}},
			{Host: iofog.ControllerHostECNViewerPortString, Container: &LocalContainerPort{Port: iofog.DefaultHTTPPortString, Protocol: "tcp"}},
		},
		ContainerName: GetLocalContainerName("controller", false),
		Image:         image,
		Privileged:    false,
		Binds:         []string{},
		NetworkMode:   "bridge",
		Credentials:   credentials,
	}
}

// NewLocalContainerClient returns a LocalContainer struct
func NewLocalContainerClient() (*LocalContainer, error) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	if err := client.FromEnv(cli); err != nil {
		return nil, err
	}
	return &LocalContainer{
		client: cli,
	}, nil
}

// GetLogsByName returns the logs of the container specified by name
func (lc *LocalContainer) GetLogsByName(name string) (stdout, stderr string, err error) {
	ctx := context.Background()
	r, err := lc.client.ContainerLogs(ctx, name, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return
	}
	defer r.Close()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	_, err = stdcopy.StdCopy(stdoutBuf, stderrBuf, r)
	if err != nil {
		return
	}

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	return
}

func (lc *LocalContainer) GetContainerByName(name string) (types.Container, error) {
	ctx := context.Background()
	// List containers
	containers, err := lc.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return types.Container{}, err
	}

	// Find by name
	for idx := range containers {
		container := &containers[idx]
		for _, containerName := range container.Names {
			if containerName == "/"+name { // Docker prefixes names with /
				return *container, nil
			}
		}
	}
	return types.Container{}, util.NewInputError(fmt.Sprintf("Could not find container %s", name))
}

func (lc *LocalContainer) ListContainers() ([]types.Container, error) {
	ctx := context.Background()
	return lc.client.ContainerList(ctx, types.ContainerListOptions{})
}

// CleanContainer stops and remove a container based on a container name
func (lc *LocalContainer) CleanContainer(name string) error {
	ctx := context.Background()

	container, err := lc.GetContainerByName(name)
	if err != nil {
		return err
	}
	// Stop container if running (ignore error if there is no running container)
	if err := lc.client.ContainerStop(ctx, container.ID, nil); err != nil {
		return err
	}

	// Force remove container
	return lc.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
}

func (lc *LocalContainer) CleanContainerByID(id string) error {
	ctx := context.Background()

	// Stop container if running (ignore error if there is no running container)
	if err := lc.client.ContainerStop(ctx, id, nil); err != nil {
		return err
	}

	// Force remove container
	return lc.client.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true})
}

func (lc *LocalContainer) getPullOptions(config *LocalContainerConfig) (ret types.ImagePullOptions) {
	dockerUser := config.Credentials.User
	dockerPwd := config.Credentials.Password

	if dockerUser != "" {
		authConfig := types.AuthConfig{
			Username: dockerUser,
			Password: dockerPwd,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			panic(err)
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		ret.RegistryAuth = authStr
	}
	return
}

func getImageTag(image string) string {
	if strings.HasPrefix(image, "docker.io/") {
		return image[len("docker.io/"):]
	}
	return image
}

func (lc *LocalContainer) waitForImage(image string, counter int8) error {
	if counter >= 18 { // 180 seconds
		return util.NewInternalError("Could not find newly pulled image: " + image)
	}
	ctx := context.Background()
	imgs, listErr := lc.client.ImageList(ctx, types.ImageListOptions{All: true})
	if listErr != nil {
		return util.NewError(fmt.Sprintf("Could not list local images: %v\n", listErr))
	}
	for idx := range imgs {
		for _, tag := range imgs[idx].RepoTags {
			if tag == image {
				return nil
			}
		}
	}
	time.Sleep(10 * time.Second)
	return lc.waitForImage(image, counter+1)
}

// DeployContainer deploys a container based on an image and a port mappin
func (lc *LocalContainer) DeployContainer(containerConfig *LocalContainerConfig) (string, error) {
	ctx := context.Background()

	portSet := nat.PortSet{}
	portMap := nat.PortMap{}

	// Create port mappings
	for _, port := range containerConfig.Ports {
		natPort, err := nat.NewPort(port.Container.Protocol, port.Container.Port)
		if err != nil {
			return "", err
		}
		portSet[natPort] = struct{}{}
		portMap[natPort] = []nat.PortBinding{
			{
				HostIP:   containerConfig.Host,
				HostPort: port.Host,
			},
		}
	}

	dockerContainerConfig := &dockerContainer.Config{
		Image:        containerConfig.Image,
		ExposedPorts: portSet,
	}
	hostConfig := &dockerContainer.HostConfig{
		PortBindings: portMap,
		Privileged:   containerConfig.Privileged,
		Binds:        containerConfig.Binds,
		NetworkMode:  dockerContainer.NetworkMode(containerConfig.NetworkMode),
	}

	// Pull image
	reader, err := lc.client.ImagePull(ctx, containerConfig.Image, lc.getPullOptions(containerConfig))
	imageTag := getImageTag(containerConfig.Image)
	if err != nil {
		Verbose(fmt.Sprintf("Could not pull image: %v, listing local images...\n", err.Error()))
		imgs, listErr := lc.client.ImageList(ctx, types.ImageListOptions{All: true})
		if listErr != nil {
			Verbose(fmt.Sprintf("Could not list local images: %v\n", listErr))
			return "", err
		}
		found := false
		for idx := range imgs {
			for _, tag := range imgs[idx].RepoTags {
				if tag == imageTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			Verbose(fmt.Sprintf("Could not pull image: %v\n Could not find image [%v] locally, please run docker pull [%v]\n", err, containerConfig.Image, containerConfig.Image))
			return "", err
		}
	} else {
		defer reader.Close()
		_, err := ioutil.ReadAll(reader)
		if err != nil {
			return "", err
		}
		// Wait for image to be discoverable by docker daemon
		err = lc.waitForImage(imageTag, 0)
		if err != nil {
			return "", err
		}
	}

	container, err := lc.client.ContainerCreate(ctx, dockerContainerConfig, hostConfig, nil, nil, containerConfig.ContainerName)
	if err != nil {
		return "", util.NewError(fmt.Sprintf("Failed to create container: %v\n", err))
	}

	// Start container
	err = lc.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", util.NewError(fmt.Sprintf("Failed to start container: %v\n", err))
	}

	return container.ID, err
}

// Returns endpoint to reach controller container from within another container
func (lc *LocalContainer) GetLocalControllerEndpoint() (controllerEndpoint string, err error) {
	host, err := lc.GetContainerIP(GetLocalContainerName("controller", false))
	if err != nil {
		return controllerEndpoint, err
	}
	controllerEndpoint = fmt.Sprintf("http://%s:%s", host, iofog.ControllerPortString)
	return
}

func (lc *LocalContainer) GetContainerIP(name string) (ip string, err error) {
	container, err := lc.GetContainerByName(name)
	if err != nil {
		return
	}

	network, found := container.NetworkSettings.Networks[container.HostConfig.NetworkMode]
	if !found {
		return "", util.NewNotFoundError(fmt.Sprintf("Container %s : Could not find network setting for network %s", name, container.HostConfig.NetworkMode))
	}

	return network.IPAddress, nil
}

func (lc *LocalContainer) WaitForCommand(containerName string, condition *regexp.Regexp, command ...string) error {
	for iteration := 0; iteration < 120; iteration++ {
		output, err := lc.ExecuteCmd(containerName, command)
		if err != nil {
			Verbose(fmt.Sprintf("Container command %v failed with error %v\n", command, err.Error()))
		}
		if condition.MatchString(output.StdOut) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return util.NewInternalError("Timed out waiting for container")
}

func (lc *LocalContainer) ExecuteCmd(name string, cmd []string) (execResult ExecResult, err error) {
	ctx := context.Background()

	container, err := lc.GetContainerByName(name)
	if err != nil {
		return
	}

	// Create command to execute inside container
	execConfig := types.ExecConfig{AttachStdout: true, AttachStderr: true,
		Cmd: cmd}
	execStartCheck := types.ExecStartCheck{}

	execID, err := lc.client.ContainerExecCreate(ctx, container.ID, execConfig)
	if err != nil {
		return
	}

	// Attach command to container
	res, err := lc.client.ContainerExecAttach(ctx, execID.ID, execStartCheck)
	if err != nil {
		return
	}
	defer res.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, res.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return execResult, err
		}
		break

	case <-ctx.Done():
		return execResult, ctx.Err()
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}

	inspect, err := lc.client.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = inspect.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)

	// Run command
	if err = lc.client.ContainerExecStart(ctx, execID.ID, execStartCheck); err != nil {
		return
	}
	return execResult, nil
}

func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	srcLength := len(filepath.ToSlash(src))

	// walk through every file in the folder
	err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file == src {
			// Skip root folder
			return nil
		}
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide relative name. Get everything after the source
		name := string([]rune(filepath.ToSlash(file))[srcLength:])
		header.Name = name

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}

func (lc *LocalContainer) CopyToContainer(name, source, dest string) (err error) {
	ctx := context.Background()

	container, err := lc.GetContainerByName(name)
	if err != nil {
		return
	}

	// content must be a Reader to a TAR
	// tar + gzip
	var content bytes.Buffer
	_ = compress(source, &content)

	// Create dest folder in container if not exists
	if _, err = lc.ExecuteCmd(name, []string{"mkdir", "-p", dest}); err != nil {
		return err
	}

	return lc.client.CopyToContainer(ctx, container.ID, dest, &content, types.CopyToContainerOptions{})
}
