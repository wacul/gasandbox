package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

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

	interval  = Cmd.Flag("interval", "interval second").Default("1.0").Float()
	reqCount  = Cmd.Flag("count", "request count").Default("5").Int()
	startDate = "2019-09-25"

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
	intervalSec := *interval * float64(time.Second)
	start := time.Now()

	service, err := newService(ctx)
	if err != nil {
		return err
	}

	date, err := time.Parse(dateFormat, startDate)
	if err != nil {
		return err
	}

	for i := 0; i < *reqCount; i++ {
		reqStart := time.Now()
		if err := doRequest(ctx, service, date); err != nil {
			return err
		}
		reqSec := time.Since(reqStart).Seconds()
		fmt.Printf("%03d: %vs\n", i, reqSec)

		date = date.AddDate(-1, 0, 0)

		time.Sleep(time.Duration(intervalSec))
	}

	fmt.Println("")
	fmt.Printf("%vs\n", time.Since(start).Seconds())

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
