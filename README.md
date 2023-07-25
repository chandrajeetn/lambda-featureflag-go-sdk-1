## Lambda-Featureflag-Go-Sdk

local evaluation sdks for golang.

### Local Setup Instructions
- Clone this repo inside `$GOPATH/bin/src/github.com/LambdaTestIncPrivate`
- Install dependencies. 
- Add below config variable and dependent key through env variable(os.Setenv(key,value)).
```
LOCAL_EVALUATION_CONFIG_DEBUG = false (enables debug logs for amplitude).
LOCAL_EVALUATION_CONFIG_SERVER_URL = "https://api.lab.amplitude.com/" (amplitude server url or evaluation proxy server url).
LOCAL_EVALUATION_CONFIG_POLL_INTERVAL = 30 (poller interval for flag rules from amplitude).
LOCAL_EVALUATION_CONFIG_POLLER_REQUEST_TIMEOUT = 10 (poller request timeout).
LOCAL_EVALUATION_DEPLOYMENT_KEY = "" (server side deployment key).
```
- run locally.
```
$ go mod vendor; go build
```