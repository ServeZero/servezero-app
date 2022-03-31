FROM golang:alpine

WORKDIR /go/src/app

# RUN apk update \
#   && apk add git \
#   && go install github.com/cosmtrek/air@latest \
#   && GO111MODULE=on go get golang.org/x/tools/gopls@latest

RUN apk update
RUN apk add git
RUN go install github.com/cosmtrek/air@latest
RUN go install golang.org/x/tools/gopls@latest

# for building optional Analysis Tools
RUN apk add build-base

# set timezone
RUN apk add --update --no-cache tzdata \
  && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
  && echo "Asia/Tokyo" > /etc/timezone \
  && apk del tzdata

# install sqlite
RUN apk add sqlite
