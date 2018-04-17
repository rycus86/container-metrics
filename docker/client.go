package docker

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"

	"github.com/rycus86/container-metrics/model"
)

type Client struct {
	client  *dockerClient.Client
	timeout time.Duration
}

func NewClient(timeout time.Duration) (*Client, error) {
	cli, err := dockerClient.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &Client{
		client:  cli,
		timeout: timeout,
	}, nil
}

func (c *Client) GetEngineStats() (*model.EngineStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	info, err := c.client.Info(ctx)
	if err != nil {
		return nil, err
	}

	return &model.EngineStats{
		Host:              info.Name,
		Images:            info.Images,
		Containers:        info.Containers,
		ContainersRunning: info.ContainersRunning,
		ContainersStopped: info.ContainersStopped,
		ContainersPaused:  info.ContainersPaused,
	}, nil
}

func (c *Client) GetContainers() ([]model.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	dockerContainers, err := c.client.ContainerList(ctx, dockerTypes.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]model.Container, len(dockerContainers))
	mapped := map[string]model.Container{}

	for idx, item := range dockerContainers {
		imageName := item.Image
		if at_index := strings.Index(imageName, "@"); at_index >= 0 {
			imageName = imageName[0:at_index]
		}

		containers[idx] = model.Container{
			Id:     item.ID,
			Name:   item.Names[0][1:],
			Image:  imageName,
			Labels: item.Labels,
		}

		mapped[item.ID] = containers[idx]
	}

	return containers, nil
}

func (c *Client) GetStats(container *model.Container) (*model.Stats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	response, err := c.client.ContainerStats(ctx, container.Id, false)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var dockerStats dockerTypes.StatsJSON
	err = json.NewDecoder(response.Body).Decode(&dockerStats)
	if err != nil {
		return nil, err
	}

	return convertStats(&dockerStats, response.OSType), nil
}

func (c *Client) ListenForEvents(channel chan<- []model.Container) {
	messages, errors := c.client.Events(context.Background(), dockerTypes.EventsOptions{})

	for {
		select {
		case message := <-messages:
			if message.Status == "start" || message.Status == "destroy" {
				waitFor := make(chan interface{})

				func() {
					containers, err := c.GetContainers()
					if err != nil {
						log.Println("Failed to reload containers", err)
					} else {
						channel <- containers
					}

					close(waitFor)
				}()

				<-waitFor
			}

		case <-errors:
			log.Println("Stop listening for events")
			return
		}
	}
}
