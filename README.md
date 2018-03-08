# freecloud

[![Build Status](https://travis-ci.org/freecloudio/freecloud.svg?branch=master)](https://travis-ci.org/freecloudio/freecloud)
[![license](https://img.shields.io/github/license/freecloudio/freecloud.svg)](https://github.com/freecloudio/freecloud/blob/master/LICENSE)


[![GitHub issues](https://img.shields.io/github/issues-raw/freecloudio/freecloud.svg)](https://github.com/freecloudio/freecloud/issues?q=is%3Aopen+is%3Aissue)
[![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/freecloudio/freecloud.svg)](https://github.com/freecloudio/freecloud/issues?q=is%3Aissue+is%3Aclosed)


## Installing

We support Go version 1.9 upwards, 1.7 and 1.8 *should* work, but are untested (come on, they are over a year old!).
Once you have grabbed a recent version of Go, just run the following:

```
mkdir -p $GOPATH/src/github.com/freecloudio/
cd $GOPATH/src/github.com/freecloudio/
git clone --recursive https://github.com/freecloudio/freecloud
go install github.com/freecloudio/freecloud
```

If you want to be sure to have the newest frontend-client then install yarn and run the following:

```
cd $GOPATH/src/github.com/freecloudio/freecloud/client/
yarn
yarn build
```

## API reference

The API reference is hosted on the github pages of this project: [https://freecloudio.github.io/freecloud/](https://freecloudio.github.io/freecloud/)

___

[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com)
[![forthebadge](https://forthebadge.com/images/badges/built-with-love.svg)](https://forthebadge.com)

[![forthebadge](https://forthebadge.com/images/badges/gluten-free.svg)](https://forthebadge.com)
[![forthebadge](https://forthebadge.com/images/badges/powered-by-netflix.svg)](https://forthebadge.com)