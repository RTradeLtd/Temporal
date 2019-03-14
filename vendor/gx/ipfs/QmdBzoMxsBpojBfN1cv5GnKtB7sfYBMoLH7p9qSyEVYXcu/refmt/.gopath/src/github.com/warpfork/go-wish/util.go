package wish

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/warpfork/go-wish/difflib"
)

func getCheckerShortName(fn Checker) string {
	fqn := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	cut := strings.LastIndex(fqn, ".")
	if cut < 0 {
		return fqn
	}
	return fqn[cut+1:]
}

func strdiff(a, b string) string {
	result, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:       escapishSlice(strings.SplitAfter(a, "\n")),
		B:       escapishSlice(strings.SplitAfter(b, "\n")),
		Context: 3,
	})
	if err != nil {
		panic(fmt.Errorf("diffing failed: %s", err))
	}
	return result
}

func escapish(s string) string {
	q := strconv.Quote(s)
	return q[1:len(q)-1] + "\n"
}

func escapishSlice(ss []string) []string {
	for i, s := range ss {
		ss[i] = escapish(s)
	}
	return ss
}
