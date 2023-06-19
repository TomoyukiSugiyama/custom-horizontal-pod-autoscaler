package job

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type JobClient interface {
	Start(ctx context.Context)
	Stop()
}

type jobClient struct {
	api      v1.API
	interval time.Duration
	stopCh   chan struct{}
}

func NewClient() (JobClient, error) {

	client, err := api.NewClient(api.Config{Address: "http://localhost:9090"})

	if err != nil {
		return nil, err
	}
	api := v1.NewAPI(client)

	c := &jobClient{api: api}

	return c, nil
}

func (c *jobClient) getTemporaryScaleMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	query := "temporary_scale"
	now := time.Now()
	rangeParam := v1.Range{Start: now.Add(-time.Hour), End: now, Step: 60 * time.Second}
	resQueryRange, warning, err := c.api.QueryRange(ctx, query, rangeParam)
	if err != nil {
		fmt.Printf("err : %s", err)
		os.Exit(1)
	}
	if len(warning) > 0 {
		fmt.Printf("err get query range %v", warning)
	}

	fmt.Printf("Res : \n %v\n", resQueryRange)
}

func (c *jobClient) Start(ctx context.Context) {
	logger := log.FromContext(ctx)
	logger.Info("start job")

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.getTemporaryScaleMetrics()
		case <-c.stopCh:
			logger.Info("received stop signal")
			return
		}
	}
}

func (c *jobClient) Stop() {
	close(c.stopCh)
}
