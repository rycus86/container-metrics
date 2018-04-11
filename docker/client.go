package docker

import (
	"context"
	"encoding/json"

	"fmt"
	dockerTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/rycus86/container-metrics/model"
	"time"
)

type Client struct {
	client *dockerClient.Client
	latest *map[string]model.Container
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

func (c *Client) GetEngineStats() (*model.EngineStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dockerContainers, err := c.client.ContainerList(ctx, dockerTypes.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	containers := make([]model.Container, len(dockerContainers))
	mapped := map[string]model.Container{}

	for idx, item := range dockerContainers {
		containers[idx] = model.Container{
			Id:     item.ID,
			Name:   item.Names[0],
			Image:  item.Image,
			Labels: item.Labels,
		}

		mapped[item.ID] = containers[idx]
	}

	c.latest = &mapped

	return containers, nil
}

func (c *Client) GetStats(container *model.Container) (*model.Stats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	return convertStats(&dockerStats), nil
}

func (c *Client) ListenForEvents(channel chan<- []model.Container) {
	messages, errors := c.client.Events(context.Background(), dockerTypes.EventsOptions{})

	for {
		select {
		case message := <-messages:
			if message.Status == "start" || message.Status == "destroy" {
				waitFor := make(chan interface{})

				func() {
					previous := *c.latest
					newContainers, err := c.GetContainers()

					containers := make([]model.Container, len(newContainers), len(newContainers))

					for idx, container := range newContainers {
						if existing, ok := previous[container.Id]; ok {
							containers[idx] = existing
						} else {
							containers[idx] = container
						}
					}

					if err != nil {
						fmt.Println("Failed to reload containers", err)
					} else {
						fmt.Println("Reloading with", len(containers), "containers")
						channel <- containers
					}

					close(waitFor)
				}()

				<-waitFor
			}

		case <-errors:
			fmt.Println("Stop listening for events")
			return
		}
	}
}
