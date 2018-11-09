default: buildgo

depensure:
	dep ensure

buildgo:
	go build -o freecloud-server ./cmd/freecloud-server

rungo:
	./freecloud-server --host=0.0.0.0 --port=8080

testunit:
	go test ./...

generateserver:
	./generate_swagger.sh

cleardb:
	rm freecloud.db