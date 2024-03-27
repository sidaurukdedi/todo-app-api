## TODO App API
This Project is simple todo API with upload attachment feature.


[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com) 
[![forthebadge](https://forthebadge.com/images/badges/built-with-love.svg)](https://forthebadge.com)
[![forthebadge](https://forthebadge.com/images/badges/contains-technical-debt.svg)](https://forthebadge.com)
[![forthebadge](https://forthebadge.com/images/badges/check-it-out.svg)](https://forthebadge.com)

### Requirements
- Golang version 1.21+ (https://golang.org/)


### Installing

A step by step series of examples that tell you have to get a development env running

Say what the step will be with .env.example
- Create ENV file (.env) with this configuration:
```
APP_NAME=go-codebase
PORT=9091
MARIADB_RO_HOST=localhost
MARIADB_RO_RO_PORT=3306
MARIADB_RO_USERNAME=username
MARIADB_RO_PASSWORD=password
MARIADB_RO_DATABASE=database
MARIADB_RO_MAX_OPEN_CONNECTIONS=25
MARIADB_RO_MAX_IDLE_CONNECTIONS=25
KAFKA_BROKERS=localhost
KAFKA_SSL_ENABLE=true
KAFKA_USERNAME=username
KAFKA_PASSWORD=password
KAFKA_SSL_ENABLE=false
BASIC_AUTH_USERNAME=username
BASIC_AUTH_PASSWORD=password
AES_SECRET=yoursecretkey
AES_IV=yoursalt
```
- Then run this command (Development Issues)
```
Give the example
...
$ make run-dev
```