package controller

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type JobClient interface {
	Start()
	Stop()
}

type jobClient struct {
	api v1.API
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

func (c *jobClient) Start() {
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

func (c *jobClient) Stop() {

}
