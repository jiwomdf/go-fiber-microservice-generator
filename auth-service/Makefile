export GO111MODULE=on
NAME=auth-service

all:
	make init clean build

init:
	go mod tidy

build:
	go mod tidy
	go build -o ${NAME} app/main.go

clean:
	if [ -f ${NAME} ] ; then rm ${NAME} ; fi

run:
	go mod tidy
	go run app/main.go -c config.local.yaml
