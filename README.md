# shortly
url shortening service written in go

## Design
* For the shortened URLs using base64 encoding, a 6 letter long key has been chosen granting 64^6 possible values (over 68 billion)
* This is a userless service so there will be no delete functionality and a second user requesting the same link will receive the same value as the first
* Our data consists of many small files, it is non-relational and read heavy. Dynamodb offers a low-effort managed solution which fits these criteria, thus it has been chosen as our datastore. We can always set a cache up in front of this if performance is insufficient.

## TODO
* add access via api key so that we are not vulnerable to malicious url creating attacks whereby our whole short url space is filled up
* add optional alias to create api
* Monitoring of popularity of URLs
* Use logging library that emits line numbers
* Inject dynamodb creds into docker container
* Add a basic HTML create frontend which accepts original urls and displays the resultant short url for the user
* Rate limiting of clients
* Handle collision of url hashes
