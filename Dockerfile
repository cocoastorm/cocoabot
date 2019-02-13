FROM golang:alpine as builder

RUN apk update && apk add --no-cache build-base git

COPY . $GOPATH/src/git.cocoastorm.com/khoa/cocoabot
WORKDIR $GOPATH/src/git.cocoastorm.com/khoa/cocoabot

RUN go get -d -v

RUN GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/cocoabot

FROM opencoconut/ffmpeg

RUN apk update && apk add --no-cache ca-certificates

COPY --from=builder /go/bin/cocoabot /go/bin/cocoabot

ENTRYPOINT [ "/go/bin/cocoabot" ]
