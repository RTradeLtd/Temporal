package metricsprometheus

import (
	"strings"

	logging "github.com/ipfs/go-log"
	metrics "github.com/ipfs/go-metrics-interface"
	pro "github.com/prometheus/client_golang/prometheus"
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
	err := pro.Register(res)
	if err != nil {
		if registered, ok := err.(pro.AlreadyRegisteredError); ok {
			if existing, ok := registered.ExistingCollector.(pro.Counter); ok {
				log.Warningf("using existing prometheus collector: %s\n", c.name)
				return existing
			}
		}
		log.Errorf("Registering prometheus collector, name: %s, error: %s\n", c.name, err.Error())
	}
	return res
}
func (c *creator) Gauge() metrics.Gauge {
	res := pro.NewGauge(pro.GaugeOpts{
		Name: c.name,
		Help: c.helptext,
	})
	err := pro.Register(res)
	if err != nil {
		if registered, ok := err.(pro.AlreadyRegisteredError); ok {
			if existing, ok := registered.ExistingCollector.(pro.Gauge); ok {
				log.Warningf("using existing prometheus collector: %s\n", c.name)
				return existing
			}
		}
		log.Errorf("Registering prometheus collector, name: %s, error: %s\n", c.name, err.Error())
	}
	return res
}
func (c *creator) Histogram(buckets []float64) metrics.Histogram {
	res := pro.NewHistogram(pro.HistogramOpts{
		Name:    c.name,
		Help:    c.helptext,
		Buckets: buckets,
	})
	err := pro.Register(res)
	if err != nil {
		if registered, ok := err.(pro.AlreadyRegisteredError); ok {
			if existing, ok := registered.ExistingCollector.(pro.Histogram); ok {
				log.Warningf("using existing prometheus collector: %s\n", c.name)
				return existing
			}
		}
		log.Errorf("Registering prometheus collector, name: %s, error: %s\n", c.name, err.Error())
	}
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
	err := pro.Register(res)
	if err != nil {
		if registered, ok := err.(pro.AlreadyRegisteredError); ok {
			if existing, ok := registered.ExistingCollector.(pro.Summary); ok {
				log.Warningf("using existing prometheus collector: %s\n", c.name)
				return existing
			}
		}
		log.Errorf("Registering prometheus collector, name: %s, error: %s\n", c.name, err.Error())
	}
	return res
}
