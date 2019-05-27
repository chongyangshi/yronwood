package endpoints

import (
	"github.com/monzo/typhon"
)

func Service() typhon.Service {
	router := typhon.Router{}
	router.GET("/", handleIndex)
	router.PUT("/upload", uploadImage)
	router.GET("/uploads/:accesstype/:filename", viewImage)
	router.POST("/list", listImages)
	router.GET("/robots.txt", handleRobots)

	svc := router.Serve().Filter(typhon.ErrorFilter).Filter(typhon.H2cFilter)

	return svc
}
