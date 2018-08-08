# shortly
url shortening service written in go

## Design
* For the shortened URLs using base64 encoding, a 6 letter long key has been chosen granting 64^6 possible values (over 68 billion)
* This is a userless service so there will be no delete functionality and a second user requesting the same link will receive the same value as the first

## TODO
* add access via api key so that we are not vulnerable to malicious url creating attacks whereby our whole short url space is filled up
* add optional alias to create api
* add a Delete endpoint
* Monitoring of popularity of URLs
* Use logging library that emits line numbers
