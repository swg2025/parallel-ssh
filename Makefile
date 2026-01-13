KVERSION := $(shell uname -r)
PWD := $(shell pwd)



build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ansible.linux . 
	CGO_ENABLED=0 go build -o ./bin/ansible.darwin.arm64 .
