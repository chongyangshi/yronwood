package endpoints

import "github.com/monzo/typhon"

func handleIndex(req typhon.Request) typhon.Response {
	return req.Response("Yronwood is a private image storage and sharing service.")
}
