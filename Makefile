cat := $(if $(filter $(OS),Windows_NT),type,cat)
TAG := $(shell $(cat) VERSION)

build: export GOOS = linux
build: export GOARCH = amd64

build:
	go build -o main.bin.tmp .

pack: build
	docker build -t notify17/k8s-node-watcher-example:$(TAG) .

push: pack
	docker push notify17/k8s-node-watcher-example:$(TAG)