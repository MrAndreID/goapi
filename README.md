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
- [Go](https://golang.org/) >= 1.26

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
- Set `USE_DATABASE=true` and Configuring The `DATABASE_*` Value in .env file
- Run Migration for The `MrAndreID/GoAPI`
```go
# go run ./internal/application/database/migration --migrate=default
```
- Run Migration for The `MrAndreID/GoAPI` with Drop All Tables
```go
# go run ./internal/application/database/migration --migrate=fresh
```

## Seeder

To Run Seeder for The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Run Migration for The `MrAndreID/GoAPI` Before Run Seeder
- Run Seeder for The `MrAndreID/GoAPI`
```go
# go run ./internal/application/database/seeder --seed=default
```

## Unit Test

To Run Unit Test for The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Create .env.test file from .env.test.example (Linux)
```sh
# cp .env.test.example .env.test
```
- Configuring .env.test file
- Prepare a PostgreSQL Server for The Repository Unit Test
- Run Unit Test for The `MrAndreID/GoAPI`
```go
# go test -v -cover -coverpkg=./internal/feature/... ./internal/feature/... -count=1
```
- Run Unit Test for The `MrAndreID/GoAPI` with Coverage Profile
```go
# go test -v -cover -coverpkg=./internal/feature/... -coverprofile=coverage.out ./internal/feature/... -count=1
# go tool cover -func=coverage.out
```

## Usage

To use The `MrAndreID/GoAPI`, you must ensure that you meet the following requirements:
- Directory Structure The `MrAndreID/GoAPI`

| Name                                       | Description                                               |
| :----------------------------------------- | :-------------------------------------------------------- |
| `cmd/api`                                  | Entry Point for The Application                           |
| `internal/application`                     | Initialization of Echo Framework, Middleware, and Routes. |
| `internal/application/cache`               | Configuration for Cache                                   |
| `internal/application/config`              | Configuration from Env File                               |
| `internal/application/database`            | Configuration for Database                                |
| `internal/application/database/migration`  | Migration Command                                         |
| `internal/application/database/seeder`     | Seeder Command                                            |
| `internal/application/dependency`          | Feature Dependency Validation on Start Up                 |
| `internal/application/message_broker`      | Configuration for Message Broker                          |
| `internal/application/object_storage`      | Configuration for Object Storage                          |
| `internal/entity`                          | Shared Struct Data                                        |
| `internal/feature/v1/user`                 | User Feature and Unit Test                                |
| `storage`                                  | Log File and Maintenance Flag File                         |

- Run The `MrAndreID/GoAPI`
```go
# go run ./cmd/api
```
- Run The `MrAndreID/GoAPI` with Docker
```docker
# docker build --no-cache -t goapi:1.0.0 .
# docker run --name goapi --restart=always -d -p 10001:10001 -v /path/to/.env:/app/.env:ro -v /path/to/folder:/app/storage goapi:1.0.0
```
- Set The `MrAndreID/GoAPI` to Maintenance Mode in Storage Folder
```sh
# touch storage/maintenance.flag
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
