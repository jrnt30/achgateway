// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package notify

import (
	"context"
	"errors"
	"fmt"

	"github.com/moov-io/achgateway/internal/service"

	"github.com/PagerDuty/go-pagerduty"
)

type PagerDuty struct {
	client     *pagerduty.Client
	from       string
	serviceKey string
}

func NewPagerDuty(cfg *service.PagerDuty) (*PagerDuty, error) {
	client := &PagerDuty{
		client:     pagerduty.NewClient(cfg.ApiKey),
		from:       cfg.From,
		serviceKey: cfg.ServiceKey,
	}
	if err := client.Ping(); err != nil {
		return nil, err
	}
	return client, nil
}

func (pd *PagerDuty) Ping() error {
	if pd == nil || pd.client == nil {
		return errors.New("pagerduty: nil")
	}

	// make a call and verify we don't error
	ctx := context.Background()
	resp, err := pd.client.ListAbilitiesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("pagerduty list abilities: %v", err)
	}
	if len(resp.Abilities) <= 0 {
		return fmt.Errorf("pagerduty: missing abilities")
	}

	return nil
}

func (pd *PagerDuty) Info(msg *Message) error {
	// Skip sending Info notifications, PagerDuty is setup for critical alerts
	return nil
}

func (pd *PagerDuty) Critical(msg *Message) error {
	opts := &pagerduty.CreateIncidentOptions{
		Type:  "incident",
		Title: fmt.Sprintf("ERROR during file %s", msg.Direction),
		Body: &pagerduty.APIDetails{
			Type:    "incident_body",
			Details: fmt.Sprintf("FAILURE on %s of %s", msg.Direction, msg.Filename),
		},
		Service: &pagerduty.APIReference{
			Type: "service_reference",
			ID:   pd.serviceKey,
		},
	}
	if msg.Direction == Download {
		// Downloads don't have to such a high priority
		opts.Urgency = "low"
	}
	return pd.createIncident(opts)
}

func (pd *PagerDuty) createIncident(opts *pagerduty.CreateIncidentOptions) error {
	ctx := context.Background()
	_, err := pd.client.CreateIncidentWithContext(ctx, pd.from, opts)
	return err
}
