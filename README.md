# freecloud

## Installing

We support Go version 1.9 upwards, 1.7 and 1.8 *should* work, but are untested (come on, they are over a year old!).
Once you have grabbed a recent version of Go, just run the following:

```
mkdir -p $GOPATH/src/github.com/freecloudio/
cd $GOPATH/src/github.com/freecloudio/
git clone --recursive https://github.com/freecloudio/freecloud
go install github.com/freecloudio/freecloud
```

## API reference

If an error occurred during a request, you'll receive (at least) this JSON payload:

```json
{
  message: "The error message goes here"
}
```

and an appropriate HTTP status code. Note that those messages are not i18n-ed and should only be displayed to users if there is no way around it.


If a request was successful, you should get at least a success flag like this:

```json
{
  success: true
}
```

Extra data in a response payload will be stored in-line in an effort to minify those responses.