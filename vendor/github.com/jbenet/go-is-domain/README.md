# go-is-domain

This package is dedicated to [@whyrusleeping](https://github.com/whyrusleeping).

Docs: https://godoc.org/github.com/jbenet/go-is-domain


Check whether something is a domain.


```Go

import (
  isd "github.com/jbenet/go-is-domain"
)

isd.IsDomain("foo.com") // true
isd.IsDomain("foo.bar.com.") // true
isd.IsDomain("foo.bar.baz") // false

```

MIT Licensed

## Updating TLDs

To update non-extended TLDs, IANA publishes, you can retrieve them from [data.iana.org](https://data.iana.org/TLD/tlds-alpha-by-domain.txt).

After retrieving the updated list, enter them into the file `tlds-alpha-by-domain.txt`. In order to update the `TLDs` map in `tlds.go`, you can run the `gen.sh` script which will generate the contents of a `string -> bool` map. After that, you'll want to replace the contents of the existing `TLDs` map, with the one that was generated and stored in `formatted_tlds.txt`
