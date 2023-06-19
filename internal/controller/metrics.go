package controller

import (
	"context"
)

type Metrics interface {
	Start()
	Stop()
}

type metrics struct {
}

func New(ctx context.Context) (metrics, error) {
	m := metrics{}
	return m, nil
}

func (m *metrics) Start() {

}

func (m *metrics) Stop() {

}

// func TestMetrics() {
// 	client, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
// 	if err != nil {
// 		fmt.Println("connection error")
// 		os.Exit(1)
// 	}
// 	api := v1.NewAPI(client)
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	_, err = api.TargetsMetadata(ctx, "{job=\"prometheus-pushgateway\"}", "", "")
// 	if err != nil {
// 		fmt.Printf("err : %s", err)
// 		os.Exit(1)
// 	}

// 	query := "temporary_scale"
// 	now := time.Now()
// 	rangeParam := v1.Range{Start: now.Add(-time.Hour), End: now, Step: 60 * time.Second}
// 	resQueryRange, warning, err := api.QueryRange(ctx, query, rangeParam)
// 	if err != nil {
// 		fmt.Printf("err : %s", err)
// 		os.Exit(1)
// 	}

// 	if len(warning) > 0 {
// 		fmt.Printf("err get query range %v", warning)
// 	}

// 	fmt.Printf("Res : \n %v\n", resQueryRange)
// }
