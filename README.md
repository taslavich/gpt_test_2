# Introduction

This is a template for doing go-micro development using GitLab. It's based on the
[helloworld](https://github.com/micro/services/blob/master/helloworld) Go Micro
template.

# Reference links

- [GitLab CI Documentation](https://docs.gitlab.com/ee/ci/)
- [Go Micro Introduction](https://micro.dev/introduction)
- [Go Micro Documentation](https://micro.dev/docs)

# Getting started

First thing to do is update `main.go` with your new project path:

```diff
-       proto "gitlab.com/gitlab-org/project-templates/go-micro/proto"
+       proto "gitlab.com/$YOUR_NAMESPACE/$PROJECT_NAME/proto"
```

Note that these are not actual environment variables, but values you should
replace.

## What's contained in this project

- main.go - is the main definition of the service, handler and client
- proto - contains the protobuf definition of the API

## Dependencies

Install the following

- [micro](https://github.com/micro/micro)
- [protoc-gen-micro](https://github.com/micro/micro/tree/master/cmd/protoc-gen-micro)

## Run Service

```shell
go run main.go --server_address localhost:8080
```

## Query Service

```
micro call --address localhost:8080 greeter Greeter.Hello '{"name": "John"}'
```

## Generate code from protobuf file

Make sure you have installed `protoc` and `protoc-gen-go` as described in https://github.com/go-micro/generator.
You can then generate the go code by running:

```shell
 protoc --go_opt=paths=source_relative --micro_opt=paths=source_relative --proto_path=./proto/ --micro_out=./proto/ --go_out=./proto/ ./proto/greeter.proto
```
