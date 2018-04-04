package docker

import (
	"context"
	"encoding/json"

	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/rycus86/container-metrics/container"
	"github.com/rycus86/container-metrics/stats"
)

type Client struct {
	client *dockerClient.Client
}

func NewClient() (*Client, error) {
	cli, err := dockerClient.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &Client{
		client: cli,
	}, nil
}

func (c *Client) GetContainers() ([]container.Container, error) {
	dockerContainers, err := c.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]container.Container, len(dockerContainers))

	for idx, item := range dockerContainers {
		containers[idx] = container.Container{
			Id:     item.ID,
			Name:   item.Names[0],
			Labels: item.Labels,
		}
	}

	return containers, nil
}

func (c *Client) GetStats(container *container.Container) (*stats.Stats, error) {
	response, err := c.client.ContainerStats(context.Background(), container.Id, false)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var dockerStats dockerTypes.StatsJSON
	err = json.NewDecoder(response.Body).Decode(&dockerStats)
	if err != nil {
		return nil, err
	}

	return convertStats(&dockerStats), nil
}

func (c *Client) StreamStats(container *container.Container, out chan<- *stats.Stats) {
	response, err := c.client.ContainerStats(context.Background(), container.Id, true)
	if err != nil {
		close(out)
		return
	}
	defer response.Body.Close()

	for {
		var dockerStats dockerTypes.StatsJSON
		err = json.NewDecoder(response.Body).Decode(&dockerStats)
		if err != nil {
			break
		}

		out <- convertStats(&dockerStats)
	}

	close(out)
}
