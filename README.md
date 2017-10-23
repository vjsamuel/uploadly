# upload.ly

|Name | Detail|
|---|---|
| University | [San Jose State University](http://www.sjsu.edu/)|
| Couse | [Cloud Technologies](http://info.sjsu.edu/web-dbgen/catalog/courses/CMPE281.html)|
|Professor| [Sanjay Garje](https://www.linkedin.com/in/sanjaygarje/) |


## Introduction

Have you ever wanted a simple interface to manage your files on the cloud? Have you ever wanted to play with several cloud technologies? If you have answered *YES* to both then lets talk about **upload.ly**. Upload.ly is a tool that leverages various cloud technologies on Google Cloud Platform to provide a simple, yet powerful experience.

### Features:
---


* **API Driven** - Upload.ly provides simple REST APIs through which files can be managed.
* **Secure** - We believe that HTTPS is the way to go and all of upload.ly services are secured with HTTPS.
* **No account creation required** - Do you have a Google account? If yes, you can start using it today.
* **On Cloud** - The entire application is deployed on Google Cloud. Google has years of exerience in running applications at scale and it can help upload.ly scale as well. 
* **Self Healing** - Upload.ly uses the power of Kubernetes to be self managed. Kubernetes can bring up workloads that have died and ensures that the workloads are always alive with the minimum number of replicas.
* **High performant** - Upload.ly is built to scale. It uses concepts of CDN and in-memory caches to offer a very low latency user experience. It can be scaled automatically when traffic increase/decreases.
* **Responsive** - The user experience is built on angular.js and provides a responsive experience out of the box. Hence upload.ly is also mobile friendly

### Technologies Used:
---
*  **Google Cloud Platform** - GCP provides the entire infrastructure for powering upload.ly
 *  **Google Container Engine** - Google's implementation of Kubernetes on which upload.ly is deployed. 
 *  **Google AppEngine Flex** - The newer version of AppEngine built on containers (we use this only to build our containers to work well with GCP offerings).
 *  **Google Container Registry** - GCR is used to store all our built docker containers.
 *  **Google Cloud Storage** - Cloud Storage is used to store all of the users files
 *  **Datastore** -  Datastore is used to store metadata associated with users and their files.
 *  **Pub/Sub** - PubSub is used to write incoming files to a message bus which can be consumed.
 *  **Functions** - Serverless technology used to read off of the message bus and write to Cloud Storage.
 *  **VPC Network** - VPC network is used to allocate our global static IP addresses.
 *  **StackDriver** - Stackdriver provides all our logging and monitoring capabilities.
 *  **Cloud DNS** - Cloud DNS manages provides the nameservers on which A records are defined.
 *  **Cloud CDN** - CDN provides edge caching of files being uploaded to upload.ly.
* **Google Domains** - Domains is used to purchase required domain for the service.
* **Google Signin** - Signin is used to provide the auth capabilities of Google accounts.
* **Helm** - Helm is used to deploy standard application stacks on to Kubernetes.
* **Memcache** - Memcache is used to provide in-memory caches of request responses. 

## Pre-requisites

### Cloud set-up
---
Following are the pre-requisites for deploying upload.ly on your own:

* Create a Google Cloud Platform account [here](https://cloud.google.com/).
* Create a project name of choice
* Enable all of the above GCP technologies on your GCP Console. (It is sufficient to just enable for now).
* Create a Google API Console project and client ID using this [reference](https://developers.google.com/identity/sign-in/web/devconsole-project).
 * In the Authorized Javascript origins section of your sign-in console provide all origins of choice. It is best to also provide `http://localhost:8080` so that local development is easier.
* On the API console enable APIs of all of the above mentioned cloud technologies. 
* Create a bucket on Cloud storage with a name of choice.
* Using the same bucket name, create a topic on PubSub.
* Follow the procedure in the [functions README](function/README.md) to have the function deployed.


### Local development
--- 
Local development and testing is covered in the [service README](service/README.md).

### Cloud deployment
---
Following are the steps required to be followed to deploy on to Google Cloud Platform:

#### Required Tools


* Install `gcloud` CLI and initialize it using:

```
curl https://sdk.cloud.google.com | bash
gcloud init
```
* Install `kubectl` using:

```
brew install kubectl
```

* Install helm using and get helm running using:

```
brew install kubernetes-helm
helm init
```

#### Cloud configuration and deployment

* Create a Kubernetes cluster of Google Container Engine using the steps documented [here](https://cloud.google.com/container-engine/docs/quickstart).
* The following attributes need to be set while configuring the Kubernetes cluster during creation:  
 * Node Image - Container-Optimized
 * Machine Type - f1-micro (costs lesser)
 * Minimal Size - 3
 * Maximal Size - <desired size to autoscale>
* Create a static global IP called `uploadly` using the following command:

```
gcloud compute addresses create uploadly --global
```   

* Build and push the docker container using:

```
cd $GOPATH/src/github.com/vjsamuel/uploadly
gcloud container builds submit --tag gcr.io/<project-id>/uploadly:1.0.12 .
```
* Create the namespace using:

```
kubectl create -f specs/namespace.yml
```

* Install memcache using:

```
helm install memcache stable/memcached --namespace=project
```

* Deploy the required configurations:

``` 
kubectl create -f specs/configmap.yml
```

* Deploy the application and service of uploadly:

```
kubectl create -f specs/deployment.yml
kubectl create -f specs/service.yml
```

* To run the application in insecure mode(on port 80) the Ingress spec for the same can be deployed as:

```
kubectl create -f specs/ingress-insecure.yml
```

* If upload.ly needs to be deployed as over SSL, then the required certificate needs to be procured from a suitable provide in .pem format as `tls.crt` and `tls.key` and uploaded to Kubernetes as:

```
kubectl create secret generic uploadly-cert --from-file=tls.crt --from-file=tls.key -n project
```

(Hint: A free and easy way to get a certificate would be to use `certbot`)

(Command: `sudo certbot certonly --manual -d <uploadly subdomain>`)

* Once the secret is in place the ingress can be replaced with:

```
kubectl replace -f specs/ingress.yml
```

* Goto [Google CDN](https://console.cloud.google.com/net-services/cdn) and click on "Add Origin" to add the newly provision load balancer so that all requests being served as cached based on the API that is being hit.

### DNS Configuration
Now the entire application has been deployed and the DNS needs to be configured.  Following are the steps required to create the DNS A records:

* Goto [Google Domains](https://domains.google/#/) and purchase a domain of choice.
* In [Google Cloud DNS console](https://console.cloud.google.com/net-services/dns/zones) create a zone and provide the name of the domain in the configuration.
* Click on "Setup Registrar" and copy the name servers on the list and add them as custom nameservers in Google Domains.
* Click on "Add record set" in the new Cloud DNS Zone and add the static IP that was created to a domain name of choice like `uploadly.<domain>` as an **A record**. Save it and wait for teh DNS to do the required replication.

This will help you be all set to start using upload.ly on your own cloud account!

## API Spec

Endpoint: [https://uploadly.vjsamuel.me/api/v1](https://uploadly.vjsamuel.me/api/v1)

All requests must pass a header `X-CloudProject-Token`. It is a google account access token that can be generated [here](https://developers.google.com/oauthplayground/)

* Type in `profile` in Authorize APIs and click the button
* Sign in in the following prompt
* Click on "Exchange Authorization code for tokens"
* Copy `id_token` in the obtained response and pass it as the value of `X-CloudProject-Token` header. 

### Upload/Update a file

```
Path: /files
Method: POST|PUT
Content-Type: multipart/form-data

Accepted form inputs:
file: file
description: text
```

|Response Code | Comment|
|---|---|
| 202| Input file was accepted|
|403| Unauthorized. Please provide an X-CloudProject-Token with the request headers|
|500| Internal server error. Please try again|
|400| File size is greater than 10 MB|

Sample Response: N/A

### Get list of files

```
Path: /files
Method: GET
Content-Type: application/json

```

|Response Code | Comment|
|---|---|
| 200| Success|
|403| Unauthorized. Please provide an X-CloudProject-Token with the request headers|
|500| Internal server error. Please try again|

Sample Response: 

```
[
	{
		"file": "Eiffel.jpg",
		"upload_time": "2017-10-21T00:26:23.695627Z",
		"last_modified": "2017-10-23T03:59:48.858025Z",
		"version": 2,
		"size": 60317,
		"type": "image/jpeg",
		"description": "updated"
	},
	{
		"file": "decoded.jpeg",
		"upload_time": "2017-10-23T16:49:10.259336Z",
		"last_modified": "2017-10-23T16:49:10.259336Z",
		"version": 1,
		"size": 60326,
		"type": "image/jpeg",
		"description": "this is a test"
	}

```

### Get File

```
Path: /file/{file}
Method: GET

```

|Response Code | Comment|
|---|---|
| 200| Success|
|403| Unauthorized. Please provide an X-CloudProject-Token with the request headers|
|500| Internal server error. Please try again|

Sample Response: File requested


### Get File Info

```
Path: /file/{file}/info
Method: GET
Content-Type: application/json

```

|Response Code | Comment|
|---|---|
| 200| Success|
|403| Unauthorized. Please provide an X-CloudProject-Token with the request headers|
|500| Internal server error. Please try again|

Sample Response: 

```
{
	"file": "decoded.jpeg",
	"upload_time": "2017-10-23T16:49:10.259336Z",
	"last_modified": "2017-10-23T16:49:10.259336Z",
	"version": 1,
	"size": 60326,
	"type": "image/jpeg",
	"description": "this is a test"
}
```

### Delete File

```
Path: /file/{file}
Method: DELETE

```

|Response Code | Comment|
|---|---|
| 200| Successful deletion|
|403| Unauthorized. Please provide an X-CloudProject-Token with the request headers|
|500| Internal server error. Please try again|

Sample Response: N/A


