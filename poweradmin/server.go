// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import (
	"context"
	"net/url"
	"strings"

	"github.com/contentways/poweradmin-go/v4/poweradmin/schema"
)

// ServerStatus is the status of the PowerDNS server behind Poweradmin.
type ServerStatus struct {
	Running    bool
	ServerID   string
	DaemonType string
	Version    string
	// UptimeSeconds is nil when PowerDNS does not report an uptime.
	UptimeSeconds *int
	// Metrics maps PowerDNS statistic names to their values, as reported by PowerDNS.
	Metrics map[string]string
	// Slaves is only filled when [ServerStatusOpts.IncludeSlaves] is set.
	Slaves []SlaveStatus
}

// SlaveStatus is the reachability of a configured autoprimary (supermaster) server.
type SlaveStatus struct {
	IP string
	// Status is "ok", "unreachable" or "skipped".
	Status      string
	LastChecked *string
	Error       *string
}

// ServerStatusOpts controls what [ServerClient.Status] returns.
type ServerStatusOpts struct {
	// Metrics limits the returned metrics to these names. Empty returns all.
	Metrics []string
	// IncludeSlaves additionally probes the autoprimary servers. This is slower.
	IncludeSlaves bool
}

func (o ServerStatusOpts) values() url.Values {
	v := url.Values{}
	if len(o.Metrics) > 0 {
		v.Set("metrics", strings.Join(o.Metrics, ","))
	}
	if o.IncludeSlaves {
		v.Set("include", "slaves")
	}
	return v
}

// ServerClient reads the status of the PowerDNS server.
type ServerClient struct {
	client *Client
}

// Status returns the status of the PowerDNS server.
//
// Requires Poweradmin 4.5 or later and the server_status_view permission
// (administrators have it implicitly). When PowerDNS is unreachable the server
// answers 503 and Status returns an error for which [IsServiceUnavailable]
// reports true; a server without the PowerDNS API configured answers 501.
func (s *ServerClient) Status(ctx context.Context, opts ServerStatusOpts) (*ServerStatus, *Response, error) {
	var result schema.ServerStatusResponse
	resp, err := s.client.get(ctx, appendQuery("server/status", opts.values()), &result)
	if err != nil {
		return nil, resp, err
	}
	status := ServerStatusFromSchema(result)
	return &status, resp, nil
}
