package rollbar

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
)

// Recovery middleware for rollbar error monitoring
func Recovery(onlyCrashes bool) gin.HandlerFunc {

	return func(c *gin.Context) {
		defer func() {

			if rval := recover(); rval != nil {
				debug.PrintStack()

				roll.CriticalStack(errors.New(fmt.Sprint(rval)), getCallers(3), map[string]string{
					"endpoint": c.Request.RequestURI})

				c.AbortWithStatus(http.StatusInternalServerError)
			}

			if !onlyCrashes {
				for _, item := range c.Errors {
					roll.Error(item.Err, map[string]string{
						"meta":     fmt.Sprint(item.Meta),
						"endpoint": c.Request.RequestURI,
					})
				}
			}
		}()

		c.Next()
	}
}

func getCallers(skip int) (pc []uintptr) {
	pc = make([]uintptr, 1000)
	i := runtime.Callers(skip+1, pc)
	return pc[0:i]
}
