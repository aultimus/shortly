# shortly
url shortening service written in go

## Design
* For the shortened URLs using base64 encoding, a 6 letter long key has been chosen granting 64^6 possible values (over 68 billion)


## TODO
* add access via api key so that we are not vulnerable to malicious url creating attacks whereby our whole short url space is filled up
* add optional expiry date, alias to create api
* add a Delete endpoint
* Monitoring of popularity of URLs
