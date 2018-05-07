FROM golang
ENV PATH /usr/local/go/bin/go:$PATH
WORKDIR /go/src
ENTRYPOINT ["/go/docker-run.sh"]
EXPOSE 8187
