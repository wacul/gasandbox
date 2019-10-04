package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/analytics/v3"
	analyticsreporting "google.golang.org/api/analyticsreporting/v4"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// Secret :
type Secret struct {
	ViewID       string `json:"viewId"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RefreshToken string `json:"refreshToken"`
}

var (
	// Cmd :
	Cmd = kingpin.CommandLine

	secretFile = Cmd.Flag("secret", "secret config file").Required().ExistingFile()
	secret     = Secret{}

	concurrent = Cmd.Flag("concurrent", "concurrent count").Default("5").Int()
	reqCount   = Cmd.Flag("count", "request count").Default("5").Int()
	startDate  = "2019-09-25"

	dateFormat = "2006-01-02"
)

func main() {
	Cmd.Name = "ga"

	if _, err := Cmd.Parse(os.Args[1:]); err != nil {
		Cmd.FatalUsage(fmt.Sprintf("\x1b[33m%+v\x1b[0m", err))
	}
	if err := run(); err != nil {
		Cmd.Fatalf("+%v", err)
	}
}

func run() error {
	ctx := context.Background()
	start := time.Now()

	service, err := newService(ctx)
	if err != nil {
		return err
	}

	date, err := time.Parse(dateFormat, startDate)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	ch := make(chan error)
	for i := 0; i < *reqCount; i += *concurrent {
		fmt.Printf("all count: %v\n", i)

		var errs *multierror.Error
		go func() {
			for i := 0; i < *concurrent; i++ {
				err := <-ch
				if err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}()

		wg.Add(*concurrent)
		reqStart := time.Now()
		for count := 0; count < *concurrent; count++ {
			fmt.Printf("concurrent: %v\n", count)
			go func() {
				ch <- doRequest(ctx, service, date)
				wg.Done()
			}()
		}
		wg.Wait()

		if err := errs.ErrorOrNil(); err != nil {
			return err
		}

		took := time.Since(reqStart)
		fmt.Printf("took: %vs\n", took.Seconds())
		sleep := 1.0*time.Second - took
		if sleep > 0 {
			time.Sleep(sleep)
			fmt.Printf("sleep: %vs\n", sleep.Seconds())
		}
	}

	fmt.Printf("\n%vs\n", time.Since(start).Seconds())

	return nil
}

func doRequest(ctx context.Context, service *analyticsreporting.Service, date time.Time) error {
	r := &analyticsreporting.ReportRequest{
		ViewId: secret.ViewID,
		DateRanges: []*analyticsreporting.DateRange{
			{
				StartDate: date.Format(dateFormat),
				EndDate:   date.Format(dateFormat),
			},
		},
		Metrics: []*analyticsreporting.Metric{
			{
				Expression: "ga:sessions",
			},
		},
	}
	req := analyticsreporting.GetReportsRequest{
		ReportRequests: []*analyticsreporting.ReportRequest{r},
	}

	quotaUser := googleapi.QuotaUser("fixed")
	if _, err := service.Reports.BatchGet(&req).Context(ctx).Do(quotaUser); err != nil {
		return err
	}

	return nil
}

func newService(ctx context.Context) (*analyticsreporting.Service, error) {
	file, err := ioutil.ReadFile(*secretFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(file, &secret); err != nil {
		return nil, err
	}

	config := &oauth2.Config{
		ClientID:     secret.ClientID,
		ClientSecret: secret.ClientSecret,
		Scopes: []string{
			analytics.AnalyticsScope,
			analytics.AnalyticsReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}
	token := oauth2.Token{
		RefreshToken: secret.RefreshToken,
	}
	tokenSource := config.TokenSource(ctx, &token)

	service, err := analyticsreporting.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}

	return service, nil
}
