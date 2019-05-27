package endpoints

import "github.com/monzo/typhon"

func handleRobots(req typhon.Request) typhon.Response {
	return req.Response("User-agent: *\nDisallow: /uploads/\n")
}
