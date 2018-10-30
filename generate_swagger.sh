#!/bin/bash

# cd into the directory of the script
cd "$(dirname "$0")"

if ! type "swagger" > /dev/null; then
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
fi

GIT_DIR=api

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
	git clone -b ${GIT_BRANCH} https://github.com/freecloudio/api.git ${GIT_DIR} > /dev/null

	echo "Cloned proto into ${GIT_DIR} and checked out ${GIT_BRANCH}"
fi

echo "$(pwd)"

swagger validate ./api/freecloud.yml > /dev/null
swagger generate server -A freecloud -P models.User -f ./api/freecloud.yml > /dev/null

