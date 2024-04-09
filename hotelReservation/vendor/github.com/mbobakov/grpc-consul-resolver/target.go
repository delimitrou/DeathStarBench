package consul

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/form"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
)

type target struct {
	Addr              string        `form:"-"`
	User              string        `form:"-"`
	Password          string        `form:"-"`
	Service           string        `form:"-"`
	Wait              time.Duration `form:"wait"`
	Timeout           time.Duration `form:"timeout"`
	MaxBackoff        time.Duration `form:"max-backoff"`
	Tag               string        `form:"tag"`
	Near              string        `form:"near"`
	Limit             int           `form:"limit"`
	Healthy           bool          `form:"healthy"`
	TLSInsecure       bool          `form:"insecure"`
	Token             string        `form:"token"`
	Dc                string        `form:"dc"`
	AllowStale        bool          `form:"allow-stale"`
	RequireConsistent bool          `form:"require-consistent"`
	// TODO(mbobakov): custom parameters for the http-transport
	// TODO(mbobakov): custom parameters for the TLS subsystem
}

func (t *target) String() string {
	return fmt.Sprintf("service='%s' healthy='%t' tag='%s'", t.Service, t.Healthy, t.Tag)
}

//  parseURL with parameters
// see README.md for the actual format
// URL schema will stay stable in the future for backward compatibility
func parseURL(u string) (target, error) {
	rawURL, err := url.Parse(u)
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL")
	}

	if rawURL.Scheme != schemeName ||
		len(rawURL.Host) == 0 || len(strings.TrimLeft(rawURL.Path, "/")) == 0 {
		return target{},
			errors.Errorf("Malformed URL('%s'). Must be in the next format: 'consul://[user:passwd]@host/service?param=value'", u)
	}

	var tgt target
	tgt.User = rawURL.User.Username()
	tgt.Password, _ = rawURL.User.Password()
	tgt.Addr = rawURL.Host
	tgt.Service = strings.TrimLeft(rawURL.Path, "/")
	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.ParseDuration(vals[0])
	}, time.Duration(0))

	err = decoder.Decode(&tgt, rawURL.Query())
	if err != nil {
		return target{}, errors.Wrap(err, "Malformed URL parameters")
	}
	if len(tgt.Near) == 0 {
		tgt.Near = "_agent"
	}
	if tgt.MaxBackoff == 0 {
		tgt.MaxBackoff = time.Second
	}
	return tgt, nil
}

// consulConfig returns config based on the parsed target.
// It uses custom http-client.
func (t *target) consulConfig() *api.Config {
	var creds *api.HttpBasicAuth
	if len(t.User) > 0 && len(t.Password) > 0 {
		creds = new(api.HttpBasicAuth)
		creds.Password = t.Password
		creds.Username = t.User
	}
	// custom http.Client
	c := &http.Client{
		Timeout: t.Timeout,
	}
	return &api.Config{
		Address:    t.Addr,
		HttpAuth:   creds,
		WaitTime:   t.Wait,
		HttpClient: c,
		TLSConfig: api.TLSConfig{
			InsecureSkipVerify: t.TLSInsecure,
		},
		Token: t.Token,
	}
}
