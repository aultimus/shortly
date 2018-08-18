# shortly
url shortening service written in go

## Design
* For the shortened URLs using base64 encoding, a 6 letter long key has been chosen granting 64^6 possible values (over 68 billion)
* This is a userless service so there will be no delete functionality and a second user requesting the same link will receive the same value as the first
* Our data consists of many small files, it is non-relational and read heavy. Dynamodb offers a low-effort managed solution which fits these criteria, thus it has been chosen as our datastore. We can always set a cache up in front of this if performance is insufficient.
* A URL, when MD5summed and base64 encoded results in a string of length 24 (144 bit), each character having 64 possible values. Thus there are 64^24 possible values for an md5sum hash. In truncating this string to six characters (32 bit) we are reducing the hash space to 64^6 possible values. Assuming an equal distribution of urls to hash buckets, when we get 30,084 entries we have a collision probability of 1 in 10 and when we have 77163 entries in our db we have a collision probability of 1 in 2. This collision factor is likely unworkable for large numbers of users, the alternative would be to use A) Longer shortened URLs or B) a Map Reduce job to iterate through and store all possible keys with a cache for each application providing a subset, that approach is considered as a potential extension to this project.

## Running

```
go run shortly/main.go
```
Runs a webserver listening on port 8080


## API

/{url} endpoint

301 redirects to the original url represented by the short url 'foo' or returns 404 if short url not found.
```
curl http://localhost:8080/foo
```


### JSON endpoints

/create endpoint
```
curl localhost:8080/create -d '{"original_url": "http://foobarcat.blogspot.com"}'
{"shortened_url":"7RxfRd","error":""}
```

/redirect/{url} endpoint

equivalent to /{url} endpoint but provides parsable json response.
```
curl http://localhost:8080/redirect/foo
{"original_url":"","error":"could not find key foo"}
```

## TODO
* add access via api key so that we are not vulnerable to malicious url creating attacks whereby our whole short url space is filled up
* add optional alias to create api
* Monitoring of popularity of URLs
* Use logging library that emits line numbers
* Inject dynamodb creds into docker container
* Add a basic HTML create frontend which accepts original urls and displays the resultant short url for the user
* Rate limiting of clients
* Handle collision of url hashes
