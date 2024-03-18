package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	api "github.com/jhonjoao/remote-containers/cmd/api"
	internalApi "github.com/jhonjoao/remote-containers/cmd/internalApi"
	communication "github.com/jhonjoao/remote-containers/internal/communication"
	"github.com/jhonjoao/remote-containers/internal/docker"
	p2p "github.com/jhonjoao/remote-containers/internal/libp2p"
	"github.com/libp2p/go-libp2p/core/network"
)

var stream network.Stream
var responseChan chan communication.ResponseData
var apiChan chan communication.ResponseData

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := p2p.NewHost(ctx)

	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	dest := Input("Do you like to connect to another machine? (No - empty)")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\nExiting....")
		h.Close()
		// <-ctx.Done()
		// code to kill connection
		os.Exit(1)
	}()

	responseChan = make(chan communication.ResponseData, 1)
	apiChan = make(chan communication.ResponseData, 1)

	if dest == "" {
		p2p.StartPeer(ctx, h, handleStream)

		select {}

	} else {
		s, err := p2p.StartPeerAndConnect(ctx, h, dest)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}

		stream = *s

		dockerClient := docker.New()

		go communication.HearStream(*s, responseChan)
		go internalApi.ProcessInternalData(*s, dockerClient, responseChan, apiChan)

		api.StartApi(stream, apiChan)
	}

}
func handleStream(s network.Stream) {

	log.Println("Got a new stream!")

	stream = s

	dockerClient := docker.New()

	go communication.HearStream(s, responseChan)
	go internalApi.ProcessInternalData(s, dockerClient, responseChan, apiChan)

	api.StartApi(stream, apiChan)

}

func Input(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}

	return strings.TrimSpace(s)
}
