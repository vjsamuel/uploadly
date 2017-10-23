# Dockerfile extending the generic Go image with application files for a
# single application.
FROM gcr.io/google-appengine/golang

COPY service /go/src/github.com/vjsamuel/uploadly/service
COPY webapp /go/src/github.com/vjsamuel/uploadly/webapp
WORKDIR /go/src/github.com/vjsamuel/uploadly/service
RUN go-wrapper install -tags appenginevm
