SHELL:= $(shell which bash)

.touch_env:
	sudo touch /etc/docker-environment

build:
	GOOS=linux GOARCH=amd64 go build main.go
	docker build -t khamoud/game .

run:
	docker run -it -p 9090:9090 khamoud/game

push:
	docker push khamoud/game
