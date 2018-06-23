package route53

import r "github.com/mitchellh/goamz/route53"

// Zone.
type Zone struct {
	Id string
	c  *Client
}

// Issue change against `name`.
func (z *Zone) change(action, t, name, value string) (*r.ChangeResourceRecordSetsResponse, error) {
	req := &r.ChangeResourceRecordSetsRequest{
		Changes: []r.Change{
			r.Change{
				Action: action,
				Record: r.ResourceRecordSet{
					Name:    name,
					Type:    t,
					TTL:     300,
					Records: []string{value},
				},
			},
		},
	}

	return z.c.ChangeResourceRecordSets(z.Id, req)
}

// Add record type `t` with the given name and value,
// for example .Add("A", "foo.example.com", "n.n.n.n")
func (z *Zone) Add(t, name, value string) (*r.ChangeResourceRecordSetsResponse, error) {
	return z.change("CREATE", t, name, value)
}

// Remove record type `t` with the given name and value.
func (z *Zone) Remove(t, name, value string) (*r.ChangeResourceRecordSetsResponse, error) {
	return z.change("DELETE", t, name, value)
}

// Records returns the records present in the zone.
func (z *Zone) Records() ([]r.ResourceRecordSet, error) {
	if res, err := z.c.ListResourceRecordSets(z.Id, nil); err == nil {
		return res.Records, nil
	} else {
		return nil, err
	}
}

// RecordsByName returns records via name.
func (z *Zone) RecordsByName(name string) ([]r.ResourceRecordSet, error) {
	records := []r.ResourceRecordSet{}

	res, err := z.c.ListResourceRecordSets(z.Id, &r.ListOpts{Name: name})
	if err != nil {
		return records, nil
	}

	if nil == res.Records {
		return records, nil
	}

	for _, record := range res.Records {
		if record.Name == name+"." {
			records = append(records, record)
		}
	}

	return records, nil
}
