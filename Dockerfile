FROM golang:latest 
RUN mkdir /golang 
ADD . /golang/ 
WORKDIR /golang 
RUN go build -o yronwood . 
CMD ["/golang/yronwood"]
