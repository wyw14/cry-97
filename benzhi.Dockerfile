FROM golang:1.26.2
ENV GOPROXY=off GOSUMDB=off CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor -o /usr/local/bin/biotreat ./cmd/biotreat
EXPOSE 19697
CMD ["/usr/local/bin/biotreat"]
