package trust

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const controllerAPIWaitSeconds = 60

// WaitForControllerAPI polls the controller /status endpoint with trust-aware TLS.
func WaitForControllerAPI(ctx context.Context, namespace, endpoint string) error {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return err
	}
	statusURL := baseURL.String() + "/status"

	var lastErr error
	for seconds := 0; seconds < controllerAPIWaitSeconds; seconds++ {
		cfg := ResolveTransport(ctx, namespace, endpoint)
		client := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: cfg.TLSConfig,
			},
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			util.DrainAndCloseHTTPBody(resp.Body)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("controller status returned %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return lastErr
}
