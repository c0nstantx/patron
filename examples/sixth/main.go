package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/beatlabs/patron"
	"github.com/beatlabs/patron/log"
	"github.com/beatlabs/patron/sync"
	patronhttp "github.com/beatlabs/patron/sync/http"
	tracehttp "github.com/beatlabs/patron/trace/http"
	"github.com/pkg/errors"
)

const cacheTTL = 60

var cl tracehttp.Client

func init() {
	err := os.Setenv("PATRON_LOG_LEVEL", "debug")
	if err != nil {
		fmt.Printf("failed to set log level env var: %v", err)
		os.Exit(1)
	}
}

func main() {
	name := "sixth"
	version := "1.0.0"

	err := patron.Setup(name, version)
	if err != nil {
		fmt.Printf("failed to set up logging: %v", err)
		os.Exit(1)
	}

	// Set up routes
	routes := []patronhttp.Route{
		patronhttp.NewGetRoute("/", sixth, true),
	}

	srv, err := patron.New(
		name,
		version,
		patron.Routes(routes),
	)
	if err != nil {
		log.Fatalf("failed to create service %v", err)
	}

	ctx := context.Background()
	cl, err = tracehttp.New(tracehttp.Cache(cacheTTL))
	if err != nil {
		log.Fatalf("failed to create HTTP client %v", err)
	}
	err = srv.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run service %v", err)
	}
}

func sixth(ctx context.Context, req *sync.Request) (*sync.Response, error) {

	var p struct {
		UserID    int    `json:"userId"`
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	apiReq, err := http.NewRequest("GET", "https://jsonplaceholder.typicode.com/todos/1", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed create request")
	}
	apiReq.Header.Set("Content-Type", "application/json")
	rsp, err := cl.Do(ctx, apiReq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request to service")
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading API response")
	}
	err = json.Unmarshal(body, &p)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding API response")
	}

	return sync.NewResponse(fmt.Sprintf("got %v from service API response", p)), nil
}
