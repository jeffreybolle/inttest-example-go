package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jeffreybolle/inttest-example-go/pkg/api"
	"github.com/jeffreybolle/inttest-example-go/pkg/creditscore"
	"github.com/jeffreybolle/inttest-example-go/pkg/service"
	"github.com/jeffreybolle/inttest-example-go/pkg/store"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func initProcess(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		log.Println("process stopping")
		cancel()
		time.Sleep(time.Second)
		os.Exit(1)
	}()

	svc, err := initService(c, ctx)
	if err != nil {
		return fmt.Errorf("error while creating service: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Int("ServicePort")))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	api.RegisterAPIServer(s, svc)

	log.Println("process started")
	go registerHealthCheck(c.Int("HealthCheckPort"))

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func initService(c *cli.Context, ctx context.Context) (*service.Service, error) {
	s, err := store.NewStore(ctx, c.String("GCPProjectID"))
	if err != nil {
		return nil, err
	}
	cs := creditscore.NewCreditScore(c.String("CreditScoreURL"))
	return service.NewService(s, cs), nil
}

func registerHealthCheck(port int) {
	http.HandleFunc("/live", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("failed to server health check: %v", err)
	}
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "ServicePort",
			Usage:   "Port that the service will listen on",
			EnvVars: []string{"SERVICE_PORT"},
			Value:   9000,
		},
		&cli.IntFlag{
			Name:    "HealthCheckPort",
			Usage:   "Port that the service's health check will listen on",
			EnvVars: []string{"HEALTH_CHECK_PORT"},
			Value:   9001,
		},
		&cli.StringFlag{
			Name:    "GCPProjectID",
			Usage:   "GCP Project ID",
			EnvVars: []string{"GCP_PROJECT_ID"},
			Value:   "example",
		},
		&cli.StringFlag{
			Name:    "CreditScoreURL",
			Usage:   "URL of the Credit Score API",
			EnvVars: []string{"CREDIT_SCORE_URL"},
			Value:   "https://creditscore/api/score",
		},
	}
	app.Action = initProcess
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
