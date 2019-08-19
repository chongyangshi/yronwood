# Yronwood

Yronwood is a simple API server written in Go for hosting and sharing public, unlisted, and private images from a backend file system. It uses no database, and the file system is the only source of truth for attributes.

It is designed to run as a microservice container in Kubernetes, but can also run stand-alone. Although it would be quite awkward to configure in a stand-alone state. 

It serves as a drop-in replacement for my old [Lychee](https://github.com/LycheeOrg/Lychee)-based PHP image gallery. This ~9MB container is much leaner and faster than the old PHP/MySQL-based solution.

![Yronwood](https://images.ebornet.com/uploads/big/17af8d5a4ae2ae708e821308812ccf62.png)

## Usage

See [`Dockerfile`](https://github.com/icydoge/yronwood/tree/master/Dockerfile) for how it is built and ran; and [`k8s-backend.yaml`](https://github.com/icydoge/yronwood/tree/master/k8s-backend.yaml) for how I have configured it in my cluster.

See [`config/config.go`](https://github.com/icydoge/yronwood/tree/master/config/config.go) for how it is configured via environmental variables. In a Kubernetes cluster, an ingress (NGINX in my case) terminates TLS and runs in front of its pod.

See [`types/types.go`](https://github.com/icydoge/yronwood/tree/master/types/types.go) for the API schema. All non-GET requests have JSON request payloads, while GET requests use query string params.

## Web Interface

A simple web interface for showing and uploading the images is available as a separate containerized web service in `/web/yronwood`, with configurations enclosed. It runs as a separate web service in Kubernetes in my setup.

## Credits

I build Go services at [Monzo](https://monzo.com), so this project has extensively relied on some of Monzo's open-source projects I use day-to-day:

* [`typhon`](https://github.com/monzo/typhon) for HTTP service and request schematics
* [`slog`](https://github.com/monzo/slog) for logging
* [`terrors`](https://github.com/monzo/terrors) for internal error schematics

Separation of build and runtime containers for Go applications can be a bit of a pain for simple setups. My Dockerfile is borrowed from [this written by C Hemidy](https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324).
