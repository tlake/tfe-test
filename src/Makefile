DOCKER_IMAGE=quintilesims/guestbook

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o guestbook
	docker build -t $(DOCKER_IMAGE) .

push:
	docker push $(DOCKER_IMAGE)

release: build push

.PHONY: build push release
