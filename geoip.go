package geoip

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

var (
	_ caddy.Module             = (*GeoIP)(nil)
	_ caddyhttp.RequestMatcher = (*GeoIP)(nil)
	_ caddy.Provisioner        = (*GeoIP)(nil)
	_ caddy.CleanerUpper       = (*GeoIP)(nil)
	_ caddyfile.Unmarshaler    = (*GeoIP)(nil)
)

func init() {
	caddy.RegisterModule(GeoIP{})
}

type GeoIP struct {
	// refresh Interval
	Interval caddy.Duration `json:"interval,omitempty"`
	// request Timeout
	Timeout caddy.Duration `json:"timeout,omitempty"`

	ctx    caddy.Context
	lock   *sync.RWMutex
	logger *zap.Logger
}

func (GeoIP) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.geoip",
		New: func() caddy.Module { return new(GeoIP) },
	}
}

// getContext returns a cancelable context, with a timeout if configured.
func (m *GeoIP) getContext() (context.Context, context.CancelFunc) {
	if m.Timeout > 0 {
		return context.WithTimeout(m.ctx, time.Duration(m.Timeout))
	}
	return context.WithCancel(m.ctx)
}

func (m *GeoIP) Provision(ctx caddy.Context) error {
	m.ctx = ctx
	m.lock = new(sync.RWMutex)
	m.logger = ctx.Logger(m)
	// update in background
	go m.refreshLoop()
	return nil
}

func (m *GeoIP) refreshLoop() {
	if m.Interval == 0 {
		m.Interval = caddy.Duration(time.Hour * 12)
	}
	ticker := time.NewTicker(time.Duration(m.Interval))
	// first time update
	m.lock.Lock()
	// it's nil anyway if there is an error
	// TODO: handle
	m.lock.Unlock()
	for {
		select {
		case <-ticker.C:
			m.lock.Lock()
			m.lock.Unlock()
		case <-m.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (m *GeoIP) Cleanup() error {
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *GeoIP) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

func (m *GeoIP) Match(r *http.Request) bool {
	return true
}
