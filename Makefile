default: buildgo

depensure:
	dep ensure

buildgo:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix nocgo -o freecloud-server ./cmd/freecloud-server

rungo:
	./freecloud-server --host=0.0.0.0 --port=8080

testunit:
	go test ./...

generateserver:
	./generate_swagger.sh