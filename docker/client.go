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
	"regexp"
)

type Client struct {
	client       *dockerClient.Client
	timeout      time.Duration
	labelFilters []string
}

func NewClient(timeout time.Duration, labelFilters []string) (*Client, error) {
	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(""))
	if err != nil {
		return nil, err
	}

	return &Client{
		client:       cli,
		timeout:      timeout,
		labelFilters: labelFilters,
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

		// strip the hash after the @ if present
		if atIndex := strings.Index(imageName, "@"); atIndex >= 0 {
			imageName = imageName[0:atIndex]
		}

		containers[idx] = model.Container{
			Id:     item.ID,
			Name:   getContainerName(item),
			Image:  getContainerImage(item),
			Labels: c.getLabelsFor(item),
		}

		mapped[item.ID] = containers[idx]
	}

	return containers, nil
}

func getContainerName(c dockerTypes.Container) string {
	return c.Names[0][1:]
}

func getContainerImage(c dockerTypes.Container) string {
	imageName := c.Image

	// strip the hash after the @ if present
	if atIndex := strings.Index(imageName, "@"); atIndex >= 0 {
		imageName = imageName[0:atIndex]
	}

	return imageName
}

func (c *Client) getLabelsFor(dc dockerTypes.Container) map[string]string {
	// return all labels if there aren't any filters
	if len(c.labelFilters) == 1 && c.labelFilters[0] == "" {
		return dc.Labels
	}

	filteredLabels := map[string]string{}

	for name, value := range dc.Labels {
		for _, filter := range c.labelFilters {
			// ignore case, match from the start
			if matched, err := regexp.MatchString("(?i)^" + filter, name); err == nil && matched {
				filteredLabels[name] = value
			}
		}
	}

	return filteredLabels
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
