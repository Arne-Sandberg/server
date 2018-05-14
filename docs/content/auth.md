---
weight: 20
title: Authentication
---

# Authentication

All authenticated API requests (will be denoted on endpoint's documentation) require you to send a session token in the `Authorization` header.

> Set the authorization header like this:

```go

client := http.Client{}
req := http.NewRequest("GET", "/api/<version>/<endpoint>", nil)
req.Header.Set("Authorization", "YourToken123")
client.Do(req)

```

```shell
curl "api_endpoint_here"
  -H "Authorization: YourToken123"
```

```javascript
axios.defaults.headers.common.Authorization = "YourToken123"
```


## Creating a session `/auth/login`

## Logout `/auth/logout`
{{< endpointStats auth="true" >}}

## Creating a user `/auth/signup`