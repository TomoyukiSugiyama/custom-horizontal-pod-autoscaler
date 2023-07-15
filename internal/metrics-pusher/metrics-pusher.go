package pusher

type MetricsPusher interface {
}

type metricsPusher struct {
}

func NewPusher() (MetricsPusher, error) {
	p := metricsPusher{}

	return p, nil
}
