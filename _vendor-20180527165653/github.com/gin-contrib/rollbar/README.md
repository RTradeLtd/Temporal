# rollbar
[![Build Status](https://travis-ci.org/gin-contrib/rollbar.svg)](https://travis-ci.org/gin-contrib/rollbar)
[![codecov](https://codecov.io/gh/gin-contrib/rollbar/branch/master/graph/badge.svg)](https://codecov.io/gh/gin-contrib/rollbar)
[![Go Report Card](https://goreportcard.com/badge/github.com/gin-contrib/rollbar)](https://goreportcard.com/report/github.com/gin-contrib/rollbar)
[![GoDoc](https://godoc.org/github.com/gin-contrib/rollbar?status.svg)](https://godoc.org/github.com/gin-contrib/rollbar)
[![Join the chat at https://gitter.im/gin-gonic/gin](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/gin-gonic/gin)

Middleware to integrate with [rollbar](https://rollbar.com/) error monitoring. It uses [roll](https://github.com/stvp/roll) client for Go that reports errors and logs messages.

## Usage

### Start using it

Download and install it:

```bash
$ go get github.com/gin-contrib/rollbar
```

Import it in your code:

```go
import "github.com/gin-contrib/rollbar"
```

### Example

```go
package main

import (
	"github.com/gin-contrib/rollbar"
	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
)

func main() {
	roll.Token = "POST_SERVER_ITEM_ACCESS_TOKEN"
	//roll.Environment = "production" // defaults to "development"

	r := gin.Default()
	r.Use(rollbar.Recovery(true))

	r.Run(":8080")
}
```
