# gin-limit
Stay under the limit with this handy Gin middleware.

## Purpose
By default, [http.ListenAndServe](https://golang.org/pkg/net/http/#ListenAndServe) (which [gin.Run](https://github.com/gin-gonic/gin/blob/9e930b9bdd5a29bac38fb491ffac4e7ea84b825d/gin.go#L221) wraps) will serve an unbounded number of requests.  Limiting the number of simultaneous connections can sometimes greatly speed things up under load.  Inspired by [x/net/netutil.LimitListener](https://godoc.org/golang.org/x/net/netutil#LimitListener).

## Example
```go
package main

import (
  "github.com/aviddiviner/gin-limit"
  "github.com/gin-gonic/gin"
)

func main() {
  r := gin.Default()
  r.Use(limit.MaxAllowed(20))
  // ...
  r.Run(":8080")
}
```

## Lies, damned lies and statistics
Everyone loves synthetic benchmarks, so have some numbers from my 2015 Macbook (on a fast rendering page; single sqlite query, basic templates).

    % wrk -t12 -c400 -d20s http://localhost:4560/
    Running 20s test @ http://localhost:4560/
      12 threads and 400 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency   848.63ms  525.39ms   1.64s    62.81%
        Req/Sec    45.43     61.85   360.00     90.50%
      8908 requests in 20.10s, 21.91MB read
      Socket errors: connect 0, read 219, write 0, timeout 0
    Requests/sec:    443.19
    Transfer/sec:      1.09MB

Now 10x faster with `limit.MaxAllowed(3)` (although that would be higher in the real world).  Hooray!

    % wrk -t12 -c400 -d20s http://localhost:4560/
    Running 20s test @ http://localhost:4560/
      12 threads and 400 connections
      Thread Stats   Avg      Stdev     Max   +/- Stdev
        Latency    94.40ms   32.65ms 656.72ms   86.44%
        Req/Sec   351.61     84.32   666.00     79.32%
      84181 requests in 20.09s, 207.05MB read
      Socket errors: connect 0, read 165, write 0, timeout 0
    Requests/sec:   4189.75
    Transfer/sec:     10.30MB
