.PHONY: docket

docker:
	-rm webook
	go mod tidy
	GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	docker rmi -f webook:v0.0.1
	docker build -t webook:v0.0.1 .