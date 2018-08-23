# shortly
url shortening service written in go

There is a hosted instance of this project managed by myself (aultimus) available at sh.foobarcat.com

## Design
* For the shortened URLs using base64 encoding, a 6 letter long key has been chosen granting 64^6 possible values (over 68 billion)
* This is a userless service so there will be no delete functionality and a second user requesting the same link will receive the same value as the first
* Our data consists of many small files, it is non-relational and read heavy. Dynamodb offers a low-effort managed solution which fits these criteria, thus it has been chosen as our datastore. We can always set a cache up in front of this if performance is insufficient.
* A URL, when MD5summed and base64 encoded results in a string of length 24 (144 bit), each character having 64 possible values. Thus there are 64^24 possible values for an md5sum hash. In truncating this string to six characters (32 bit) we are reducing the hash space to 64^6 possible values. Assuming an equal distribution of urls to hash buckets, when we get 30,084 entries we have a collision probability of 1 in 10 and when we have 77163 entries in our db we have a collision probability of 1 in 2. This collision factor is likely unworkable for large numbers of users, the alternative would be to use A) Longer shortened URLs or B) a Map Reduce job to iterate through and store all possible keys with a cache for each application providing a subset, that approach is considered as a potential extension to this project.

![Graph of server response latency incurred by collisions in pure hashing implementation](collision_latency.png?raw=true "Graph of server response latency incurred by collisions in pure hashing implementation")

Test averaged 50 requests per collision level to instance running in aws. It looks like collisions are only going to be a problem for a large number of users so this can be a later optimisation after we have the front end and other features up.

## Running

```
go run shortly/main.go
```
Runs a webserver listening on port 8080

Alternatively you can use build.sh and run.sh scripts to build and run the app in a docker image

# Deployment

After building, run push.sh to run the docker image to ECR, and then upload Dockerrun.aws.json from the root directory of this project to the aws EB console.

## API

A website is served at the domain root for use by humans. A JSON API exists for programatically interfacing with the service, that JSON API is detailed below

/v1/create endpoint
```
curl localhost:8080/v1/create -d '{"original_url": "http://foobarcat.blogspot.com"}'
{"shortened_url":"7RxfRd","error":""}
```

/v1/redirect/{url} endpoint

equivalent to /{url} endpoint but provides parsable json response.
```
curl http://localhost:8080/v1/redirect/foo
{"original_url":"","error":"could not find key foo"}
```

## TODO
* add access via api key so that we are not vulnerable to malicious url creating attacks whereby our whole short url space is filled up
* add optional alias to create api
* Monitoring of popularity of URLs
* Inject dynamodb creds into docker container
* Rate limiting of clients
* Add handler for calls to favicon.ico
* Vendor deps
