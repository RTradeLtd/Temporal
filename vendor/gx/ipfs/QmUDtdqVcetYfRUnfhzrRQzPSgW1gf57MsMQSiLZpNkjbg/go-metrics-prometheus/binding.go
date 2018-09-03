package metricsprometheus

import (
	"strings"

	pro "gx/ipfs/QmYYv3QFnfQbiwmi1tpkgKF8o4xFnZoBrvpupTiGJwL9nH/client_golang/prometheus"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
	metrics "gx/ipfs/QmekzFM3hPZjTjUFGTABdQkEnQ3PTiMstY198PwSFr5w1Q/go-metrics-interface"
)

var log logging.EventLogger = logging.Logger("metrics-prometheus")

func Inject() error {
	return metrics.InjectImpl(newCreator)
}

func newCreator(name, helptext string) metrics.Creator {
	return &creator{
		name:     strings.Replace(name, ".", "_", -1),
		helptext: helptext,
	}
}

var _ metrics.Creator = &creator{}

type creator struct {
	name     string
	helptext string
}

func (c *creator) Counter() metrics.Counter {
	res := pro.NewCounter(pro.CounterOpts{
		Name: c.name,
		Help: c.helptext,
	})
	c.register(res)
	return res
}
func (c *creator) Gauge() metrics.Gauge {
	res := pro.NewGauge(pro.GaugeOpts{
		Name: c.name,
		Help: c.helptext,
	})
	c.register(res)
	return res
}
func (c *creator) Histogram(buckets []float64) metrics.Histogram {
	res := pro.NewHistogram(pro.HistogramOpts{
		Name:    c.name,
		Help:    c.helptext,
		Buckets: buckets,
	})
	c.register(res)
	return res
}

func (c *creator) Summary(opts metrics.SummaryOpts) metrics.Summary {
	res := pro.NewSummary(pro.SummaryOpts{
		Name: c.name,
		Help: c.helptext,

		Objectives: opts.Objectives,
		MaxAge:     opts.MaxAge,
		AgeBuckets: opts.AgeBuckets,
		BufCap:     opts.BufCap,
	})
	c.register(res)
	return res
}

func (c *creator) register(col pro.Collector) {
	err := pro.Register(col)
	if err != nil {
		log.Errorf("Registering prometheus collector, name: %s, error: %s\n", c.name, err.Error())
	}
}
