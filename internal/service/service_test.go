package service_test

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"monitor-endpoint/internal/metrics"
	"monitor-endpoint/internal/service"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestService_CheckStatus(t *testing.T) {

	t.Run("should aggregate metrics based on status code", func(t *testing.T) {

		// sending a request to ts.URL?statusCode=xyz will return a response with status code = xyz
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			statusCodeStr := r.URL.Query()["statusCode"][0]
			statusCode, err := strconv.Atoi(statusCodeStr)
			assert.NoError(t, err)
			w.WriteHeader(statusCode)
			fmt.Fprintln(w, "Hello, client")
		}))
		defer ts.Close()

		ctx := context.Background()
		m := metrics.NewMetrics()
		testServerUrl, err := url.Parse(ts.URL)
		assert.NoError(t, err)
		urlByStatusCode := map[string]string{}

		// mock successful request
		{
			q := testServerUrl.Query()
			q.Set("statusCode", "200")
			testServerUrl.RawQuery = q.Encode()
			urlByStatusCode["200"] = testServerUrl.String()
		}

		config := service.Config{
			Timeout: 1 * time.Minute,
		}
		srv := service.NewService(m, &http.Client{}, config, time.Now)

		numSuccessfulReq := 4
		for i := 0; i < numSuccessfulReq; i++ {
			srv.CheckStatus(ctx, urlByStatusCode["200"])
		}

		// mock error request
		{
			q := testServerUrl.Query()
			q.Set("statusCode", "500")
			testServerUrl.RawQuery = q.Encode()
			urlByStatusCode["500"] = testServerUrl.String()
		}

		numFailedReq := 2
		for i := 0; i < numFailedReq; i++ {
			srv.CheckStatus(ctx, urlByStatusCode["500"])
		}

		// expected metrics
		successfulLabels := prometheus.Labels{
			"code":     "200",
			"endpoint": urlByStatusCode["200"],
		}
		assert.Equal(t, float64(numSuccessfulReq), testutil.ToFloat64(m.NumRequests.With(successfulLabels)))

		failedLabels := prometheus.Labels{
			"code":     "500",
			"endpoint": urlByStatusCode["500"],
		}
		assert.Equal(t, float64(numFailedReq), testutil.ToFloat64(m.NumRequests.With(failedLabels)))
	})

	t.Run("should return 499 if client cancels request due to timeout", func(t *testing.T) {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// waiting, to produce timeout in the client side
			<-time.After(50 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Hello, I'm slow")
		}))
		defer ts.Close()

		ctx := context.Background()
		m := metrics.NewMetrics()
		config := service.Config{
			Timeout: 10 * time.Millisecond,
		}
		srv := service.NewService(m, &http.Client{}, config, time.Now)
		// make request, notice timeout 10 ms < 50 ms
		srv.CheckStatus(ctx, ts.URL)

		// expected metrics
		labels := prometheus.Labels{
			"code":     "499",
			"endpoint": ts.URL,
		}
		assert.Equal(t, float64(1), testutil.ToFloat64(m.NumRequests.With(labels)))
	})

}
