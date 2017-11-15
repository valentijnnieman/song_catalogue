FROM golang:latest 
LABEL Name=song_catalogue_api Version=0.0.1 
ADD . /go/src/github.com/valentijnnieman/song_catalogue/api
WORKDIR /go/src/github.com/valentijnnieman/song_catalogue/api/ 
RUN go get
RUN go build -o main .
EXPOSE 8080 
CMD ["/go/src/github.com/valentijnnieman/song_catalogue/api/main"]