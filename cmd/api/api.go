package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	_ "github.com/jhonjoao/remote-containers/docs"
	communication "github.com/jhonjoao/remote-containers/internal/communication"
	"github.com/libp2p/go-libp2p/core/network"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var stream network.Stream
var apiChan <-chan communication.ResponseData

// @title Gin Swagger Remote Containers API
// @version 1.0
// @description Manage docker containers in remote machine
// @termsOfService http://swagger.io/terms/

// @host localhost:8080
// @BasePath /
// @schemes http
func StartApi(s network.Stream, channel <-chan communication.ResponseData) {

	stream = s
	apiChan = channel

	r := gin.Default()

	port := 8080

	url := ginSwagger.URL(fmt.Sprintf("http://localhost:%v/swagger/doc.json", port))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	r.GET("/", HealthCheck)
	r.GET("/containers/list", listContainers)
	r.POST("/containers/create", createContainer)

	r.GET("/containers/:id", inspectContainer)
	r.DELETE("/containers/:id", deleteContainer)

	status, err := Check(port)

	if !status || err != nil {
		port, _ = GetFreePort()
	}

	r.Run(fmt.Sprintf(":%v", port))
}

type TransactionRequest struct {
	Method string              `json:"Method"`
	Uri    string              `json:"Uri"`
	Header map[string][]string `json:"Header"`
	Body   []byte              `json:"Body"`
	Params *gin.Params         `json:"Params"`
}

func ginContextToBytes(c *gin.Context) ([]byte, error) {

	bodyBytes, err := c.GetRawData()
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}

	requestData := TransactionRequest{
		Method: c.Request.Method,
		Uri:    c.Request.RequestURI,
		Header: c.Request.Header,
		Body:   bodyBytes,
		Params: &c.Params,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %v", err)
	}

	return jsonData, nil
}

// @Summary Show the status of server.
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func HealthCheck(c *gin.Context) {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	c.JSON(http.StatusOK, res)
}

// @Summary lists all Docker containers
// @Accept  */*
// @Produce  json
// @Success 200	{object} map[string]interface{}  "ok"
// @Router /containers/list [get]
func listContainers(c *gin.Context) {

	requestData, err := ginContextToBytes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = communication.WriteData(stream, requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprint("Error sending request to another node:", err.Error())})
		return
	}

	response := <-apiChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []types.Container

	json.Unmarshal(response.Data, &result)

	c.JSON(http.StatusOK, gin.H{"containers": result})
}

// @Summary creates a new Docker container
// @Accept json
// @Produce json
// @Param data body docker.CreateRequest true "body data"
// @Success 200	{object} map[string]interface{}  "ok"
// @Router /containers/create [post]
func createContainer(c *gin.Context) {

	requestData, err := ginContextToBytes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = communication.WriteData(stream, requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprint("Error sending request to another node:", err.Error())})
		return
	}

	response := <-apiChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Err.Error()})
		return
	}

	var result container.CreateResponse

	json.Unmarshal(response.Data, &result)

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// @Summary inspects a Docker container by ID
// @Accept  */*
// @Produce  json
// @Param id path string true "id"
// @Success 200	{object} map[string]interface{}  "ok"
// @Router /containers/:id [get]
func inspectContainer(c *gin.Context) {

	requestData, err := ginContextToBytes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = communication.WriteData(stream, requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprint("Error sending request to another node:", err.Error())})
		return
	}

	response := <-apiChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Err.Error()})
		return
	}

	var result types.ContainerJSON

	json.Unmarshal(response.Data, &result)

	c.JSON(http.StatusOK, gin.H{"result": result})
}

// @Summary deletes a Docker container by ID
// @Accept  */*
// @Produce  json
// @Param id path string true "id"
// @Success 200	{object} map[string]interface{}  "ok"
// @Router /containers/:id [delete]
func deleteContainer(c *gin.Context) {

	containerID := c.Param("id")

	requestData, err := ginContextToBytes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = communication.WriteData(stream, requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprint("Error sending request to another node:", err.Error())})
		return
	}

	response := <-apiChan

	if response.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": response.Err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Container %s deleted", containerID)})
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func Check(port int) (status bool, err error) {
	host := ":" + strconv.Itoa(port)
	server, err := net.Listen("tcp", host)

	if err != nil {
		return false, err
	}

	server.Close()

	return true, nil
}
