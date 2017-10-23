## Upload.ly Service

The upload.ly service provides all the REST APIs that integrate with Google Cloud Platform.

### Pre-requisites

Following are the pre-requisites to build and run on a local Mac.

* Brew

```
/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```

* Go 1.8 or higher

```
brew install go
```

* An IDE of choice (maybe [GoGland](https://www.jetbrains.com/go/download/))
* Set a suitable GOPATH. Ex: `export GOPATH=$HOME/go`
* Pull the code locally:

```
go get github.com/vjsamuel/uploadly
```

* Install govendor

```
go get -u github.com/kardianos/govendor
```

* Pull in all vendored dependencies:

```
cd $GOPATH/src/github.com/vjsamuel/uploadly
govendor sync
```

* Install memcache and start it up

```
brew install memcached
memcached &
```

* Now a token file needs to be obtained from Google Cloud APIs console. Goto the [API console](https://console.developers.google.com/apis/credentials).
 * Click on "Create credentials" and on the drop down "service account key".
 * Select the service account and click on create. This will download a json file which can be saved in `service/token.json`

* Export the following environment variables:

```
export GOOGLE_APPLICATION_CREDENTIALS=token.json 
export BUCKET=<bucket name>
export PROJECT_ID=<project id>
```

* Start the application using:

```
go run main.go

```
