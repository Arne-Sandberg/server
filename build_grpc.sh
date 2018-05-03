#!/bin/bash

# cd into the directory of the script
cd "$(dirname "$0")"

GIT_DIR=grpc-services

if [ -d "$GIT_DIR" ]; then
  cd $GIT_DIR
	git checkout master
	git pull
	cd ..
else
	git clone https://github.com/freecloudio/grpc-services.git $GIT_DIR
fi

rm models/*.pb.go

for F in $GIT_DIR/*.proto; do
    protoc -I $GIT_DIR/ $F --go_out=plugins=grpc:models
		FILENAME=$(basename "$F" | cut -f 1 -d '.')
		protoc-go-inject-tag -input=models/$FILENAME.pb.go
done