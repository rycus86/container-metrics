package docker

import (
	"context"
	"encoding/json"

	dockerClient "github.com/docker/docker/client"
	dockerTypes "github.com/docker/docker/api/types"
)

type Client struct{
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

func (c *Client) GetContainerIDs() ([]string, error) {
	// TODO context.Background()
	containers, err := c.client.ContainerList(context.Background(), dockerTypes.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(containers))

	for idx, container := range containers {
		ids[idx] = container.ID
	}

	return ids, nil
}

func (c *Client) GetStats(containerID string) (*dockerTypes.StatsJSON, error) {
	// TODO context.Background()
	response, err := c.client.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var stats dockerTypes.StatsJSON  // TODO custom type instead
	err = json.NewDecoder(response.Body).Decode(&stats)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (c *Client) StreamStats(containerID string, out chan<- dockerTypes.StatsJSON) {
	// TODO context.Background()
	response, err := c.client.ContainerStats(context.Background(), containerID, true)
	if err != nil {
		close(out)
		return
	}
	defer response.Body.Close()

	for {
		var stats dockerTypes.StatsJSON  // TODO custom type instead
		err = json.NewDecoder(response.Body).Decode(&stats)
		if err != nil {
			break
		}

		out <- stats
	}

	close(out)
}
