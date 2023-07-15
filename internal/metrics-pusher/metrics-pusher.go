package pusher

type MetricsPusher interface {
}

type metricsPusher struct {
}

type PusherOption func(*metricsPusher)

func NewPusher(opts ...PusherOption) (MetricsPusher, error) {
	p := &metricsPusher{}

	for _, opt := range opts {
		opt(p)
	}

	return p, nil
}
