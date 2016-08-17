dev:
	go run main.go

run:
	docker run --env-file ovh.env -v /var/run/docker.sock:/var/run/docker.sock krkr/iplb-docker

example-start:
	docker-compose -f example.yml up -d

example-scale:
	docker-compose -f example.yml scale apish=3