#!/bin/bash

# cd into the directory of the script
cd "$(dirname "$0")"

GIT_DIR=grpc-services

if [ "$1" != "" ]; then
    GIT_BRANCH=$1
else
    GIT_BRANCH=master
fi


if [ -d "$GIT_DIR" ]; then
    cd ${GIT_DIR}
    git checkout -q ${GIT_BRANCH}
	git pull > /dev/null
	cd ..

	echo "Checked out ${GIT_BRANCH}"
else
	git clone https://github.com/freecloudio/grpc-services.git ${GIT_DIR} > /dev/null
	cd ${GIT_DIR}
	git checkout -q ${GIT_BRANCH}
	cd ..

	echo "Cloned proto into ${GIT_DIR} and checked out ${GIT_BRANCH}"
fi

rm models/*.pb.go

echo "Old generated Files removed"

for F in ${GIT_DIR}/*.proto; do
	echo ""
	FILENAME=$(basename "$F" | cut -f 1 -d '.')
    protoc -I ${GIT_DIR}/ ${F} --go_out=plugins=grpc:models > /dev/null
    echo "${FILENAME}.pb.go: generated"
	protoc-go-inject-tag -input=models/${FILENAME}.pb.go &> /dev/null
	echo "${FILENAME}.pb.go: tags injected"
done