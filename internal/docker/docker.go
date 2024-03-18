package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	Client *client.Client
}

func New() DockerClient {

	client, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		fmt.Println("docker client error", err.Error())
	}

	return DockerClient{
		Client: client,
	}

}

func (myDocker *DockerClient) ListContainers(options container.ListOptions) []types.Container {

	containers, err := myDocker.Client.ContainerList(context.Background(), options)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return containers
}

type CreateRequest struct {
	Image string   `json:"image" binding:"required"`
	Name  string   `json:"name"`
	Cmd   []string `json:"cmd"`
}

func (myDocker *DockerClient) CreateContainer(request CreateRequest) *container.CreateResponse {

	resp, err := myDocker.Client.ContainerCreate(context.Background(), &container.Config{
		Image: request.Image,
		Cmd:   request.Cmd,
	}, nil, nil, nil, request.Name)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return &resp
}

func (myDocker DockerClient) InspectContainer(containerID string) *types.ContainerJSON {
	containerJSON, err := myDocker.Client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return &containerJSON
}

func (myDocker DockerClient) DeleteContainer(containerID string) {

	err := myDocker.Client.ContainerRemove(context.Background(), containerID, container.RemoveOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}
