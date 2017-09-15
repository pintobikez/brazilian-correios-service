# Correios
Correios service is a small app to deal with Reverse Logistics requests to Brazilian Correios and to Perform Tracking of Posted Items
It exposes an API that allows to create ReverseLogistic request for:
- Postage (SEDEX / PAC / ESEDEX)
- Colection (SEDEX / PAC / ESEDEX)
It exposes an API that allows to perform Tracking of Posted Items


# Correios Reverse
It will response via a callback defined by the requester when:
- The amount of errors in a Request to Correios is equal to a MaxRetries (defined in the config)
- When the person as posted the product within correios (responds with the PostageCode and TrackingCode)
- When Correios accepted the Collection of the item

# Correios Tracking
It will respond via a callback defined by the requester if:
- The amount of objects to track is bigger then 5 (this is because of the time that Correios takes to respond)
If not it will reply in the Tracking request

## Requirements
App requires Golang 1.8 or later, Glide Package Manager and Docker (for building)

## Installation
- Install [Golang](https://golang.org/doc/install)
- Install [Glide](https://glide.sh)
- Install [Docker](htts://docker.com)


## Build
For building binaries please use make, look at the commands bellow:

```
// Build the binary in your environment
$ make build

// Build with another OS. Default Linux
$ make OS=darwin build

// Build with custom version.
$ make APP_VERSION=0.1.0 build

// Build with custom app name.
$ make APP_NAME=correios-service build

// Passing all flags
$ make OS=darwin APP_NAME=correios-service-docker APP_VERSION=0.1.0 build

// Clean Up
$ make clean

// Configure. Install app dependencies.
$ make configure

// Check if docker exists.
$ make depend

// Create a docker image with application
$ make pack

// Pack with custom Docker namespace. Default gfgit
$ make DOCKER_NS=gfgit pack

// Pack with custom version.
$ make APP_VERSION=0.1.0 pack

// Pack with custom app name.
$ make APP_NAME=correios-service pack

// Pack passing all flags
$ make APP_NAME=correios-service-docker APP_VERSION=0.1.0 DOCKER_NS=gfgit pack
```

## Development
```
// Running tests
$ make test

// Running tests with coverage. Output coverage file: coverage.html
$ make test-coverage

// Running tests with junit report. Output coverage file: report.xml
$ make test-report
```

## Run it
```
// Run and launch docker
$ make build; docker build -t correios-service-docker .; docker-compose up;
```

## Usage:

# Create a Postage Request
```
curl -v -X POST http://127.0.0.1:8080/reverse/ -H 'content-type:application/json' -d '{"callback":"http://localhost:8080","order_nr":68802479,"request_type":"POSTAGE","request_service":"PAC","origin_nome":"ANGELITA ALVES PORTELLA CHYBIAK","origin_logradouro":"Rua Bahia","origin_numero":234,"origin_complemento":"casa","origin_cep":"76982138","origin_bairro":"Parque Industrial Novo Tempo","origin_cidade":"Vilhena","origin_uf":"RO","origin_referencia":"prox a Art Moveis","origin_email":"417030351829@mktp.extra.com.br","slip_number":"854555215","destination_nome":"Deluxe","destination_logradouro":"Rua Luiz Maske","destination_numero":248,"destination_complemento":"","destination_cep":"89066650","destination_bairro":"Itoupavazinha","destination_cidade":"Blumenau","destination_uf":"SC","destination_referencia":"","destination_email":"anderson.paulino@befashion4ever.com.br","status":"","error_message":"","postage_code":"","tracking_code":"","created_at":"","updated_at":"","items" :[{"item":"PR667APF92CYT-10297","product_name":"Blusa Be Fashion 4ever Cropped Vermelho"}]}'
```
## Configurable parameters
```
request_type:
	 POSTAGE
	 COLECT

request_service
	PAC
	SEDEX
```

# Update a previous Postage Request
```
curl -v -X PUT http://localhost:8080/reverse/1 -H 'content-type:application/json' -d '{"callback":"http://localhost:8080","order_nr":68802479,"request_type":"POSTAGE","request_service":"PAC","origin_nome":"ANGELITA ALVES PORTELLA CHYBIAK","origin_logradouro":"Rua Bahia","origin_numero":234,"origin_complemento":"casa","origin_cep":"76982138","origin_bairro":"Parque Industrial Novo Tempo","origin_cidade":"Vilhena","origin_uf":"RO","origin_referencia":"prox a Art Moveis","origin_email":"417030351829@mktp.extra.com.br","slip_number":"854555215","destination_nome":"Deluxe","destination_logradouro":"Rua Luiz Maske","destination_numero":248,"destination_complemento":"","destination_cep":"89066650","destination_bairro":"Itoupavazinha","destination_cidade":"Blumenau","destination_uf":"SC","destination_referencia":"","destination_email":"anderson.paulino@befashion4ever.com.br","status":"","error_message":"","postage_code":"","tracking_code":"","created_at":"","updated_at":"","items" :[{"item":"PR667APF92CYT-10297","product_name":"Blusa Be Fashion 4ever Cropped Vermelho"}]}'
```
## Configurable parameters
```
request_type:
	 POSTAGE
	 COLECT

request_service
	PAC
	SEDEX
```

# Get a Request Information
```
curl -v -X GET http://127.0.0.1:8080/reverse/1
```

# Cancel a Request
```
curl -v -X DELETE  http://127.0.0.1:8080/reverse/1
```

# Perform Search
```
curl -v -X POST http://127.0.0.1:8080/reversesearch/ -d '{"from":0,"offset":20}'
```
## Available Search parameters
```
order_by_field: the field that we want to order the search for

order_by_type: 
	DESC
	ASC

from: from value for pagination

offset: offset value for pagination

where : an array that will contain the multiple where clauses (currently it joins them with AND)
  [
    field: the field that we want to search for
    value: the value of the field that we want to search for
    operator: the operator to be used in the search: like, =, >=, <=, <>, IN, NOT IN
  ]
```

# Check tracking of an object
```
 curl -v -X POST http://127.0.0.1:8080/tracking/ -H 'content-type:application/json' -d '{"callback":"http://localhost:8080","tracking_type":"ALL","language":"BR","objects":["PO444714015BR"]}'
```
## Configurable parameters
```
callback: the URL to where it will respond with the status

objects: an array containg the tracking codes (AWB)

tracking_type:
	ALL -> Retrieves all tracking status of the object
	LAST -> Retrieves the last tracking status of the object

language:
	BR -> Brazilian Portuguese
	EN -> English
```