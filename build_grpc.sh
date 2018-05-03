#!/bin/sh

git clone https://github.com/freecloudio/grpc-services.git

mkdir models

protoc -I grpc-services/ grpc-services/auth.proto --go_out=plugins=grpc:models