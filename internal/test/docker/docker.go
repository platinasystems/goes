// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package docker

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"github.com/platinasystems/go/internal/test"
	"gopkg.in/yaml.v2"
)

type Router struct {
	Hostname string
	Cmd      string
	Intfs    []struct {
		Name    string
		Address string
		Vlan    string
	}
	id string
}

type Config struct {
	Image   string
	Volume  string
	Mapping string
	Routers []Router
	cli     *client.Client
}

func Check(t *testing.T) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatalf("Unable to get docker client: %v")
		return err
	}

	ver := cli.ClientVersion()
	t.Logf("Docker client version %v", ver)
	_, err = cli.Info(context.Background())
	if err != nil {
		return err
	}
	_, err = cli.Ping(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func LaunchContainers(t *testing.T, source []byte) (config *Config) {
	assert := test.Assert{t}
	assert.Helper()

	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatalf("Unable to get docker client: %v")
		return
	}

	assert.Nil(yaml.Unmarshal(source, &config))

	config.cli = cli

	if !isImageLocal(t, config) {
		t.Log("no local container, trying to pull from remote")
		err := pullImage(t, config)
		if err != nil {
			t.Log(err)
			t.Fail()
			return
		}
		t.Log("Image %v pulled from remote\n", config.Image)
	} else {
		t.Logf("Image %v found local\n", config.Image)
	}

	path := "PATH=/usr/local/sbin"
	path += ":/usr/local/bin"
	path += ":/usr/sbin"
	path += ":/usr/bin"
	path += ":/sbin"
	path += ":/bin"
	env := []string{path}

	pwd, err := syscall.Getwd()
	if err != nil {
		t.Fatalf("Unable to find cwd: %v", err)
	}
	vdir := pwd + config.Volume

	// Common container config
	cc := &container.Config{}
	cc.Image = config.Image
	cc.Tty = true
	cc.Env = env
	cc.Volumes = map[string]struct{}{config.Mapping: {}}

	// Common host config
	ch := &container.HostConfig{}
	ch.Privileged = true
	ch.NetworkMode = "none"

	// router specific cc & ch config
	for i, router := range config.Routers {
		cc.Hostname = router.Hostname
		cc.Cmd = []string{router.Cmd}

		bind := vdir + "volumes/" + router.Hostname + ":" + config.Mapping
		ch.Binds = []string{bind}

		cresp, err := startContainer(t, config, cc, ch)
		if err != nil {
			t.Fatalf("Failed to start container %v", router.Hostname)
		}
		config.Routers[i].id = cresp.ID
		for _, intf := range router.Intfs {
			if intf.Vlan != "" {
				newIntf := intf.Name + "." + intf.Vlan
				assert.Program(test.Self{},
					"ip", "link", "set", "up", intf.Name)
				assert.Program(test.Self{},
					"ip", "link", "add", "link", intf.Name,
					"name", newIntf, "type", "vlan",
					"id", intf.Vlan)
				assert.Program(test.Self{},
					"ip", "link", "show", newIntf)
				assert.Program(test.Self{},
					"ip", "link", "set", "up", newIntf)
				moveIntfContainer(t, router.Hostname, newIntf,
					intf.Address)
			} else {
				moveIntfContainer(t, router.Hostname, intf.Name,
					intf.Address)
			}
		}
	}
	time.Sleep(1 * time.Second)
	return
}

func FindHost(config *Config, host string) (router Router, err error) {
	for _, r := range config.Routers {
		if r.Hostname == host {
			router = r
			return
		}
	}
	return
}

func ExecCmd(t *testing.T, ID string, config *Config, cmd []string) (out string,
	err error) {

	execOpts := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Detach:       false,
	}

	cli := config.cli
	ctx := context.Background()

	execResp, err := cli.ContainerExecCreate(ctx, ID, execOpts)
	if err != nil {
		t.Logf("Error creating exec: %v", err)
		return
	}

	hresp, err := cli.ContainerExecAttach(ctx, execResp.ID, execOpts)
	if err != nil {
		t.Logf("Error attaching exec: %v", err)
		return
	}
	defer hresp.Close()

	content, err := ioutil.ReadAll(hresp.Reader)
	if err != nil {
		t.Logf("Error reading output: %v", err)
		return
	}
	out = string(content)

	ei, err := cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		t.Logf("Error exec Inspect: %v", err)
		return
	}
	if ei.Running {
		t.Logf("exec still running", ei)
	}
	if ei.ExitCode != 0 {
		err = fmt.Errorf("[%v] exit code %v", cmd, ei.ExitCode)
		return
	}

	return
}

func TearDownContainers(t *testing.T, config *Config) {
	assert := test.Assert{t}
	for _, r := range config.Routers {
		for _, intf := range r.Intfs {
			if intf.Vlan != "" {
				newIntf := intf.Name + "." + intf.Vlan
				moveIntfDefault(t, r.Hostname, newIntf)
				assert.Program(test.Self{},
					"ip", "link", "del", newIntf)
			} else {
				moveIntfDefault(t, r.Hostname, intf.Name)
			}
		}
		err := stopContainer(t, config, r.Hostname, r.id)
		if err != nil {
			t.Logf("Error: stopping %v: %v", r.Hostname, err)
		}
	}
	config.cli.Close()
}

func isImageLocal(t *testing.T, config *Config) bool {

	images, err := config.cli.ImageList(context.Background(),
		types.ImageListOptions{})
	if err != nil {
		t.Fail()
		return false
	}

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == config.Image {
				return true
			}
		}
	}
	return false
}

func isContainerRunning(t *testing.T, config *Config, name string) bool {

	conts, err := config.cli.ContainerList(context.Background(),
		types.ContainerListOptions{All: true})
	if err != nil {
		t.Fail()
		return false
	}

	for _, cont := range conts {
		for _, imagename := range cont.Names {
			if imagename[1:] == name {
				return true
			}
		}
	}
	return false
}

func pullImage(t *testing.T, config *Config) error {
	repo := "docker.io/library/" + config.Image
	out, err := config.cli.ImagePull(context.Background(), repo,
		types.ImagePullOptions{})
	if err != nil {
		t.Fail()
		return err
	}
	defer out.Close()
	// io.Copy(os.Stdout, out)
	return nil
}

func startContainer(t *testing.T, config *Config, cc *container.Config,
	ch *container.HostConfig) (cresp container.ContainerCreateCreatedBody,
	err error) {

	cli := config.cli

	if isContainerRunning(t, config, cc.Hostname) {
		t.Fatalf("Container %v already running", cc.Hostname)
	}
	t.Logf("Starting container %v\n", cc.Hostname)

	ctx := context.Background()

	cresp, err = cli.ContainerCreate(ctx, cc, ch, nil, cc.Hostname)
	if err != nil {
		t.Logf("Error creating container: %v", err)
		return
	}

	err = cli.ContainerStart(ctx, cresp.ID, types.ContainerStartOptions{})
	if err != nil {
		t.Logf("Error starting container: %v", err)
		return
	}

	pid, err := getPid(cc.Hostname)
	if err != nil {
		t.Logf("Error getting pid for %v: %v", cc.Hostname, err)
	}
	src := "/proc/" + pid + "/ns/net"
	dst := "/var/run/netns/" + cc.Hostname
	test.Assert{t}.Program(test.Self{}, "ln", "-s", src, dst)
	return
}

func stopContainer(t *testing.T, config *Config, name string, ID string) error {

	t.Logf("Stopping container %v", name)

	cli := config.cli
	ctx := context.Background()

	err := cli.ContainerStop(ctx, ID, nil)
	if err != nil {
		t.Logf("Error stoping %v %v: %v", name, ID, err)
		return err
	}

	err = cli.ContainerRemove(ctx, ID,
		types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		t.Logf("Error removing volume %v: %v", name, err)
		return err
	}
	link := "/var/run/netns/" + name
	test.Assert{t}.Program("rm", link)

	return nil
}

func getPid(ID string) (pid string, err error) {

	cmd := []string{"/usr/bin/docker", "inspect", "-f", "'{{.State.Pid}}'",
		ID}
	bytes, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		return
	}
	pid = string(bytes)
	pid = strings.Replace(pid, "\n", "", -1)
	pid = strings.Replace(pid, "'", "", -1)
	return
}

func moveIntfContainer(t *testing.T, container string, intf string,
	addr string) error {

	assert := test.Assert{t}

	t.Logf("moving %v to container %v with address %v", intf, container, addr)

	assert.Program(test.Self{},
		"ip", "link", "set", intf, "netns", container)
	assert.Program(test.Self{},
		"ip", "-n", container, "link", "set", "up", "lo")
	assert.Program(test.Self{},
		"ip", "-n", container, "link", "set", "down", intf)
	// ISIS fails with default mtu 9216
	assert.Program(test.Self{},
		"ip", "-n", container, "link", "set", "mtu", "1500", "dev", intf)
	assert.Program(test.Self{},
		"ip", "-n", container, "link", "set", "up", intf)
	assert.Program(test.Self{},
		"ip", "-n", container, "addr", "add", addr, "dev", intf)
	return nil
}

func moveIntfDefault(t *testing.T, container string, intf string) error {
	t.Logf("moving %v from %v to default", intf, container)
	assert := test.Assert{t}
	assert.Program(test.Self{},
		"ip", "-n", container, "link", "set", "down", intf)
	assert.Program(test.Self{},
		"ip", "-n", container, "link", "set", intf, "netns", "1")
	return nil
}
