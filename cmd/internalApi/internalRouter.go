package internalapi

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/jhonjoao/remote-containers/cmd/api"
	"github.com/jhonjoao/remote-containers/internal/communication"
	"github.com/jhonjoao/remote-containers/internal/docker"
	"github.com/libp2p/go-libp2p/core/network"
)

type InternalHandler func(*api.TransactionRequest)

type InternalRouter struct {
	Method  string          `json:"Method"`
	Path    string          `json:"Path"`
	Handler InternalHandler `json:"Handler"`
}

var stream network.Stream
var dockerClient docker.DockerClient

func ProcessInternalData(s network.Stream, client docker.DockerClient, channel chan communication.ResponseData, apiChan chan communication.ResponseData) {

	stream = s
	dockerClient = client

	for {
		value, ok := <-channel
		if !ok {
			fmt.Println("Channel closed. Exiting goroutine.")
			return
		}

		var data api.TransactionRequest
		json.Unmarshal(value.Data, &data)

		if data.Method == "" {
			select {
			case apiChan <- value:
			case <-time.After(10 * time.Second):
				fmt.Println("Failed to add response in channel")
			}
			continue
		}

		if value.Err != nil {
			fmt.Println("Error to listen data", value.Err.Error())
			return
		}

		routes := map[string][]InternalRouter{}

		routes["GET"] = make([]InternalRouter, 0)

		routes["GET"] = append(routes["GET"], InternalRouter{
			Path:    "/containers/list",
			Handler: listContainers,
		})

		routes["GET"] = append(routes["GET"], InternalRouter{
			Path:    "/containers/:id",
			Handler: inspectContainer,
		})

		routes["POST"] = make([]InternalRouter, 0)

		routes["POST"] = append(routes["POST"], InternalRouter{
			Path:    "/containers/create",
			Handler: createContainer,
		})

		routes["DELETE"] = make([]InternalRouter, 0)

		routes["POST"] = append(routes["POST"], InternalRouter{
			Path:    "/containers/:id",
			Handler: deleteContainer,
		})

		for _, routes := range routes[data.Method] {

			if data.Params != nil {
				if matchRoute(routes.Path, data.Uri) {
					routes.Handler(&data)
					break
				}
			}

			if routes.Path == data.Uri {
				routes.Handler(&data)
				break
			}
		}

	}

}

func matchRoute(pattern, route string) bool {
	pattern = regexp.QuoteMeta(pattern)
	pattern = replacePlaceholders(pattern)

	re := regexp.MustCompile("^" + pattern + "$")

	return re.MatchString(route)
}

func replacePlaceholders(pattern string) string {
	placeholderPattern := regexp.MustCompile(`:\w+`)
	return placeholderPattern.ReplaceAllString(pattern, `[^/]+`)
}

func listContainers(w *api.TransactionRequest) {

	containers := dockerClient.ListContainers(container.ListOptions{})

	bytes, _ := json.Marshal(containers)

	err := communication.WriteData(stream, bytes)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

}

func createContainer(w *api.TransactionRequest) {

	var request docker.CreateRequest

	json.Unmarshal(w.Body, &request)

	response := dockerClient.CreateContainer(request)

	bytes, _ := json.Marshal(response)

	err := communication.WriteData(stream, bytes)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
}

func inspectContainer(w *api.TransactionRequest) {

	containerId, _ := w.Params.Get("id")

	response := dockerClient.InspectContainer(containerId)

	bytes, _ := json.Marshal(response)

	err := communication.WriteData(stream, bytes)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
}

func deleteContainer(w *api.TransactionRequest) {
	containerId, _ := w.Params.Get("id")

	dockerClient.InspectContainer(containerId)

	bytes, _ := json.Marshal("Ok")

	err := communication.WriteData(stream, bytes)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
}
