package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/chongyangshi/yronwood/config"
	"github.com/monzo/typhon"
)

func Service() typhon.Service {
	router := typhon.Router{}
	router.GET("/", handleIndex)
	router.GET("/index.html", handleIndex)
	router.POST("/authenticate", authenticate)
	router.PUT("/upload", uploadImage)
	router.GET("/uploads/:accesstype/:filename", viewImage)
	router.POST("/list", listImages)
	router.GET("/robots.txt", handleRobots)

	svc := router.Serve().Filter(typhon.ErrorFilter).Filter(typhon.H2cFilter).Filter(ClientErrorFilter).Filter(CORSFilter)

	return svc
}

// ClientErrorFilter strips sensitive error info before returning error to client, leaving
// only code and message; on a best-effort basis. Very ugly written.
func ClientErrorFilter(req typhon.Request, svc typhon.Service) typhon.Response {
	rsp := svc(req)
	if rsp.Error != nil {
		var basicErr = basicError{}
		bodyBytes, err := rsp.BodyBytes(false)
		if err != nil {
			rsp.Body = ioutil.NopCloser(bytes.NewReader(basicErr.toFailbackBytes()))
			return rsp
		}

		err = json.Unmarshal(bodyBytes, &basicErr)
		if err != nil {
			rsp.Body = ioutil.NopCloser(bytes.NewReader(basicErr.toFailbackBytes()))
			return rsp
		}

		seralized, err := basicErr.toSerialized()
		if err != nil {
			rsp.Body = ioutil.NopCloser(bytes.NewReader(basicErr.toFailbackBytes()))
			return rsp
		}

		rsp.Body = ioutil.NopCloser(bytes.NewReader(seralized))
		return rsp
	}

	return rsp
}

func CORSFilter(req typhon.Request, svc typhon.Service) typhon.Response {
	if req.Method == http.MethodOptions {
		rsp := typhon.NewResponse(req)
		rsp.Header.Set("Access-Control-Allow-Origin", config.ConfigCORSAllowedOrigin)
		rsp.Header.Set("Access-Control-Allow-Methods", "GET, PUT, POST")
		rsp.Body = ioutil.NopCloser(bytes.NewReader([]byte("ok")))
		rsp.StatusCode = http.StatusOK
		return rsp
	}

	rsp := svc(req)
	rsp.Header.Set("Access-Control-Allow-Origin", config.ConfigCORSAllowedOrigin)
	rsp.Header.Set("Access-Control-Allow-Methods", "GET, PUT, POST")

	return rsp
}

type basicError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (b basicError) toFailbackBytes() []byte {
	return []byte(fmt.Sprintf("Error (%s): %s", b.Code, b.Message))
}

func (b basicError) toSerialized() ([]byte, error) {
	seralized, err := json.Marshal(b)
	return seralized, err
}
