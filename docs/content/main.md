---
weight: 10
title: API Reference
---

# Introduction

Welcome to the freecloud API! You can use our API to manage all your files, make up- and downloads and all of the other cool stuff you can do in the web client (it is based on the API).

We provide code samples in Go, cURL and JavaScript with [axios](https://github.com/axios/axios).

This API documentation was created with [DocuAPI](https://github.com/bep/docuapi/), a multilingual documentation theme for the static site generator [Hugo](http://gohugo.io/). Huge thanks to bep for the theme, spf13 and all contributors for Hugo.

## API request scheme

All API requests will use this scheme: `<your_freecloud_host>/api/<api_version>/<endpoint>`.
The current API version is `v1`, therefore, make all requests to `/api/v1/<endpoint>`.