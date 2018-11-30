package docker

import (
	"context"
	"io"
	"io/ioutil"

	dockertypes "github.com/docker/docker/api/types"
	dockercli "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/pkg/errors"
)

type Client struct {
	*dockercli.Client
}

func New() (*Client, error) {
	cli, err := dockercli.NewClientWithOpts(dockercli.FromEnv, dockercli.WithVersion("1.38"))
	if err != nil {
		return nil, errors.Wrap(err, "new docker client")
	}
	return &Client{cli}, nil
}

func (d *Client) RunContainer(ctx context.Context, id string, stdout io.Writer, stderr io.Writer) error {
	if err := d.ContainerStart(ctx, id, dockertypes.ContainerStartOptions{}); err != nil {
		return errors.Wrap(err, "container start")
	}
	logs, err := d.ContainerLogs(ctx, id, dockertypes.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return errors.Wrap(err, "container logs stdout")
	}
	defer logs.Close()

	_, err = stdcopy.StdCopy(stdout, stderr, logs)
	return err
}

func (d *Client) PullImage(ref string) error {
	rc, err := d.ImagePull(context.Background(), ref, dockertypes.ImagePullOptions{})
	if err != nil {
		// Retry
		rc, err = d.ImagePull(context.Background(), ref, dockertypes.ImagePullOptions{})
		if err != nil {
			return err
		}
	}
	if _, err := io.Copy(ioutil.Discard, rc); err != nil {
		return err
	}
	return rc.Close()
}
