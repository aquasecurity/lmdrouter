# lmdrouter

**Go HTTP router library for AWS API Gateway-invoked Lambda Functions**

## Table of Contents

* [Overview](#overview)
    * [Features](#features)
* [Installation](#installation)
* [Usage](#usage)

## Overview

`lmdrouter` is a simple-to-use library for writing AWS Lambda functions in Go
that listen to events of type API Gateway Proxy Request. It allows creating a
lambda function that can match requests based on their URI, just like an HTTP
server would.

The library provides an interface not unlike the standard `net/http.Mux` type
or community-libraries such as [httprouter](https://github.com/julienschmidt/httprouter)
and [chi](https://github.com/go-chi/chi).

### Features

* Supports all HTTP methods
* Supports middleware at a global and per-resource level
* Provides ability to automatically "unmarshal" an API Gateway request to an
  arbitrary Go struct, with data coming from path parameters and/or query
  parameters.
* Provides ability to automatically "marshal" response structs to an API Gateway
  response (only JSON responses are currently generated).

## Installation

```shell
go get github.com/aquasecurity/lmdrouter
```

## Usage

`lmdrouter` is meant to be used inside Go Lambda functions.

```go
package main

import (
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aquasecurity/lmdrouter"
)

var router *lmdrouter.Router

func init() {
    router = lmdrouter.NewRouter("/api", loggerMiddleware, authMiddleware)
    router.Route("GET", "/", listSomethings)
	router.Route("POST", "/", postSomething, someOtherMiddleware)
	router.Route("GET", "/:id", getSomething)
	router.Route("PUT", "/:id", updateSomething)
	router.Route("DELETE", "/:id", deleteSomething)
}

func main() {
    lambda.Start(router.Handler)
}

// the rest of the code is a redacted example, it should probably reside in a
// separate library inside your project

type listSomethingsInput struct {
    ID            string  `lambda:"path.id"`              // a path parameter declared as :id
    ShowSomething bool    `lambda:"query.show_something"` // a query parameter named "show_something"
}

func listSomethings(ctx context.Context, req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
	// parse input
	var input listSomethingsInput
	err = lmdrouter.UnmarshalRequest(req.PathParameters, req.QueryStringParameters, &input)
	if err != nil {
		return lmdrouter.HandleError(err)
	}

	// call some business logic that generates an output struct
    // ...

	return lmdrouter.MarshalResponse(http.StatusOK, nil, output)
}

func loggerMiddleware(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		// [LEVEL] [METHOD PATH] [CODE] EXTRA
		format := "[%s] [%s %s] [%d]%s"
		level := "INF"
		var code int
		var extra string

		res, err = next(ctx, req)
		if err != nil {
			level = "ERR"
			code = http.StatusInternalServerError
			extra = " " + err.Error()
		} else {
			code = res.StatusCode
			if code >= 400 {
				level = "ERR"
			}
		}

        log.Printf(format, level, req.HTTPMethod, req.Path, code, extra)

		return res, err
	}
}
```
