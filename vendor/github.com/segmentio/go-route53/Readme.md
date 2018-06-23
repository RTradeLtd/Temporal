
# go-route53

 Higher-level Route53 client built on [mitchellh/goamz](https://github.com/mitchellh/goamz).

 View the [docs](http://godoc.org/pkg/github.com/segmentio/go-route53).

## Example

```go
package main

import "github.com/segmentio/go-route53"
import "github.com/mitchellh/goamz/aws"
import "encoding/json"
import "os"

func check(err error) {
  if err != nil {
    panic(err)
  }
}

func main() {
  auth, err := aws.EnvAuth()
  check(err)

  dns := route53.New(auth, aws.USWest2)

  res, err := dns.Zone("Z3T864J4ZMBODE").Add("A", "foo.test.io", "0.0.0.0")
  check(err)

  b, err := json.MarshalIndent(res, "", "  ")
  check(err)

  os.Stdout.Write(b)
}
```

# License

 MIT