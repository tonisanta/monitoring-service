FROM golang:alpine AS build
WORKDIR /go/src/myapp
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/myapp cmd/main.go

#this is a multi-stage build. This is used to keep the running docker container small while still be able to build/compile things needing a lot of dependencies.
# copy the compiled code in the build container to the new container
FROM alpine
COPY --from=build /go/bin/myapp /go/bin/myapp
ENTRYPOINT ["/go/bin/myapp"]