# MrAndreID / Go Application Programming Interface (API)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The `MrAndreID/GoAPI` is a skeleton uses the Go Programming Language (GoLang) with The Echo Framework for The Application Programming Interface (API).

## Table of Contents

* [Requirements](#requirements)
* [Installation](#installation)
* [Migration](#migration)
* [Seeder](#seeder)
* [Unit Test](#unit-test)
* [Usage](#usage)
* [Versioning](#versioning)
* [Authors](#authors)
* [Contributing](#contributing)
* [Official Documentation for Go Language](#official-documentation-for-go-language)
* [License](#license)

## Requirements

To use The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- [Go](https://golang.org/) >= 1.24

## Installation

To use The `MrAndreID/GoAPI`, you must follow the steps below:
- Clone a Repository
```git
# git clone https://github.com/MrAndreID/goapi.git
```
- Get Dependancies
```go
# go mod download
# go mod tidy
```
- Create .env file from .env.example (Linux)
```sh
# cp .env.example .env
```
- Configuring .env file

## Migration

To Run Migration for The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Run Migration for The `MrAndreID/GoAPI`
```go
# go run databases/migrations/main.go --migrate=default
```
- Run Migration for The `MrAndreID/GoAPI` with Drop All Tables
```go
# go run databases/migrations/main.go --migrate=fresh
```

## Seeder

To Run Seeder for The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Run Seeder for The `MrAndreID/GoAPI`
```go
# go run databases/seeders/main.go --seed=default
```

## Unit Test

To Run Unit Test for The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Create .env file in tests folder from .env.example (Linux)
```sh
# cp .env.example tests/.env
```
- Configuring .env file
- Run Unit Test for The `MrAndreID/GoAPI`
```go
# go test -v -cover -coverpkg=./internal/handlers ./tests
```

## Usage

To use The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Directory Structure The `MrAndreID/GoAPI`
| Name                    | Description                                               |
| :---------------------- | :-------------------------------------------------------- |
| `application`           | Initialization of Echo Framework, Middleware, and Routes. |
| `caches`                | Configuration for Cache                                   |
| `configs`               | Condiguration from Env File                               |
| `databases`             | Configuration for Database                                |
| `internal/handlers`     | HTTP Handlers                                             |
| `internal/services`     | Main Business Logic                                       |
| `internal/repositories` | Connector to Database or API External                     |
| `internal/types`        | Struct Data                                               |
| `messagebrokers`        | Configuration for Message Broker                          |
| `objectstorages`        | Configuration for Object Storage                          |
| `tests`                 | Unit Test                                                 |
- Run The `MrAndreID/GoAPI`
```go
# go run main.go
```
- Run The `MrAndreID/GoAPI` with Docker
```docker
# docker build --no-cache -t goapi:1.0.0 .
# docker run --name goapi --restart=always -d -p -v /path/to/folder:/app/storages -v /path/to/folder:/app/tests/storages 10001:10001 goapi:1.0.0
```
- Set The `MrAndreID/GoAPI` to Maintenance Mode in Storages Folder
```sh
# touch storages/maintenance.flag
```

## Versioning

I use [Semanting Versioning](https://semver.org/). For the versions available, see the tags on this repository. 

## Authors

- **Andrea Adam** - [MrAndreID](https://github.com/MrAndreID)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
Please make sure to update tests as appropriate.

## Official Documentation for Go Language

Documentation for Go Language can be found on the [Go Package website](https://pkg.go.dev/).

## License

The `MrAndreID/GoAPI` is released under the [MIT License](https://opensource.org/licenses/MIT). See the `LICENSE` file for more information.
