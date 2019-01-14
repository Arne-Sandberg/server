sourcefiles = $(wildcard **/*.go)

freecloud-server: $(sourcefiles)
	go build -o freecloud-server ./cmd/freecloud-server

run: freecloud-server
	./freecloud-server --host=0.0.0.0 --port=8080

depensure:
	dep ensure

testunit:
	go test ./...

validateswagger:
	swagger validate ./api/freecloud.yml

generateswagger: validateswagger
	swagger generate server -A freecloud -P models.Principal -f ./api/freecloud.yml

cleardb:
	rm freecloud.db
