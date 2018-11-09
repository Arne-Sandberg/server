sourcefiles = $(wildcard **/*.go)

freecloud-server: $(sourcefiles)
	go build -o freecloud-server ./cmd/freecloud-server

run: freecloud-server
	./freecloud-server --host=0.0.0.0 --port=8080

depensure:
	dep ensure

testunit:
	go test ./...

generateserver:
	./generate_swagger.sh

cleardb:
	rm freecloud.db
