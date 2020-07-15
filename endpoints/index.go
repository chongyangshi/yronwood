package endpoints

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/monzo/typhon"

	"github.com/chongyangshi/yronwood/config"
)

func handleIndex(req typhon.Request) typhon.Response {
	response := fmt.Sprintf("Click <a href='%s'>here</a> if you are not being redirectd automatically.</a>")

	rsp := typhon.NewResponse(req)
	rsp.Header.Set("Location", config.ConfigIndexRedirect)
	rsp.Header.Set("Content-Type", "text/plain")
	rsp.StatusCode = http.StatusMovedPermanently
	rsp.Body = ioutil.NopCloser(bytes.NewReader([]byte(response)))

	return rsp
}
