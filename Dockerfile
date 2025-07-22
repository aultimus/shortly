FROM golang:1.10.3-alpine

RUN make build
ADD shortly/shortly /go/bin/shortly
# should probably put in a nicer location but this is quick for now
# don't serve arbitrary files from this location or we could serve our own binary lol
WORKDIR /data
ADD static /data/static
ADD templates /data/templates
EXPOSE 8080
CMD ["/go/bin/shortly"]
