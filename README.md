# freecloud

[![Build Status](https://travis-ci.com/freecloudio/server.svg?branch=master)](https://travis-ci.org/freecloudio/freecloud)
[![license](https://img.shields.io/github/license/freecloudio/server.svg)](https://github.com/freecloudio/freecloud/blob/master/LICENSE)


[![GitHub issues](https://img.shields.io/github/issues-raw/freecloudio/server.svg)](https://github.com/freecloudio/freecloud/issues?q=is%3Aopen+is%3Aissue)
[![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/freecloudio/server.svg)](https://github.com/freecloudio/freecloud/issues?q=is%3Aissue+is%3Aclosed)


## Installing

We support Go version 1.10 and upwards.
Once you have grabbed a recent version of Go, just run the following:

```
mkdir -p $GOPATH/src/github.com/freecloudio/
cd $GOPATH/src/github.com/freecloudio/
git clone --recursive https://github.com/freecloudio/server
cd server
dep ensure
make run
```

If you want to be sure to have the newest web-client then install npm and run the following:

```
cd $GOPATH/src/github.com/freecloudio/server/client/
git checkout master && git pull
npm install
npm run build
```

## Swagger build

For building the swagger sources use `make generateswagger`.
For it to use install make sure `swagger` is installed!

## API reference

The API reference can be accessed under `localhost:8080/api/v1/docs` after starting the server.

___

[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com)
[![forthebadge](https://forthebadge.com/images/badges/gluten-free.svg)](https://forthebadge.com)

[![forthebadge](https://forthebadge.com/images/badges/contains-technical-debt.svg)](https://forthebadge.com)
[![forthebadge](https://forthebadge.com/images/badges/built-with-love.svg)](https://forthebadge.com)