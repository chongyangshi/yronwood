package endpoints

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/monzo/typhon"
)

func handleRobots(req typhon.Request) typhon.Response {
	response := fmt.Sprintf("User-agent: *\nDisallow: /uploads/\n")

	rsp := typhon.NewResponse(req)
	rsp.Header.Set("Content-Type", "text/plain")
	rsp.Body = ioutil.NopCloser(bytes.NewReader([]byte(response)))

	return rsp
}
