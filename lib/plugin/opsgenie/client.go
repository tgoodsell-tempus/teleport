/*
Copyright 2015-2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package opsgenie

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gravitational/teleport-plugins/lib"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/trace"
)

const (
	ogMaxConns    = 100
	ogHTTPTimeout = 10 * time.Second

	ogAlertKeyPrefix = "teleport-access-request"
)

var alertBodyTemplate = template.Must(template.New("alert body").Parse(
	`{{.User}} requested permissions for roles {{range $index, $element := .Roles}}{{if $index}}, {{end}}{{ . }}{{end}} on Teleport at {{.Created.Format .TimeFormat}}.
{{if .RequestReason}}Reason: {{.RequestReason}}{{end}}
{{if .RequestLink}}To approve or deny the request, proceed to {{.RequestLink}}{{end}}
`,
))
var reviewNoteTemplate = template.Must(template.New("review note").Parse(
	`{{.Author}} reviewed the request at {{.Created.Format .TimeFormat}}.
Resolution: {{.ProposedState}}.
{{if .Reason}}Reason: {{.Reason}}.{{end}}`,
))
var resolutionNoteTemplate = template.Must(template.New("resolution note").Parse(
	`Access request has been {{.Resolution}}
{{if .ResolveReason}}Reason: {{.ResolveReason}}{{end}}`,
))

// OpsgenieClient is a wrapper around resty.Client.
type OpsgenieClient struct {
	client *resty.Client
	// webProxyURL is the address used when building the bodies of the alerts
	// allowing links to the access requests to be built
	webProxyURL      *url.URL
	clusterName      string
	defaultSchedules []string
	Priority         String
}

type OpsgenieClientConfig struct {
	APIKey           string
	APIEndpoint      string
	defaultSchedules []string
	Priority         String
}

func NewOpsgenieClient(conf OpsgenieClientConfig, clusterName, webProxyAddr string) (OpsgenieClient, error) {
	var (
		webProxyURL *url.URL
		err         error
	)
	if webProxyAddr != "" {
		if webProxyURL, err = lib.AddrToURL(webProxyAddr); err != nil {
			return OpsgenieClient{}, trace.Wrap(err)
		}
	}

	client := resty.NewWithClient(&http.Client{
		Timeout: ogHTTPTimeout,
		Transport: &http.Transport{
			MaxConnsPerHost:     ogMaxConns,
			MaxIdleConnsPerHost: ogMaxConns,
		},
	})
	if conf.APIEndpoint != "" {
		client.SetHostURL(conf.APIEndpoint)
	} else {
		client.SetHostURL("") // TODO: Find reasonable default
	}
	client.SetHeader("Authorization", "GenieKey "+conf.APIKey)
	return OpsgenieClient{
		client:           client,
		clusterName:      clusterName,
		webProxyURL:      webProxyURL,
		defaultSchedules: conf.defaultSchedules,
		Priority:         conf.Priority,
	}, nil
}

// CreateAlert creates an opsgenie alert.
func (og OpsgenieClient) CreateAlert(ctx context.Context, reqID string, reqData RequestData) (OpsgenieData, error) {
	bodyDetails, err := og.buildAlertBody(reqID, reqData)
	if err != nil {
		return OpsgenieData{}, trace.Wrap(err)
	}

	body := AlertBody{
		Message:     fmt.Sprintf("Access request from %s", reqData.User),
		Alias:       fmt.Sprintf("%s/%s", ogAlertKeyPrefix, reqID),
		Description: bodyDetails,
		Responders:  og.getResponders(reqData),
		Priority:    og.Priority,
	}

	var result AlertResult
	if _, err = og.client.NewRequest().
		SetContext(ctx).
		SetBody(body).
		SetResult(&result).
		Post("alerts"); err != nil {
		return OpsgenieData{}, trace.Wrap(err)
	}

	return OpsgenieData{
		AlertID: result.Alert.ID,
	}, nil
}

func (og OpsgenieClient) getResponders(reqData RequestData) []Responder {
	schedules := og.defaultSchedules
	if reqSchedules, ok := reqData.RequestAnnotations[ReqAnnotationRespondersKey]; ok {
		schedules = reqSchedules
	}
	responders := []Responder{}
	for _, s := range schedules {
		responders = append(responders, Responder{
			Type: "schedule",
			ID:   s,
		})
	}
	return responders
}

// PostReviewNote posts a note once a new request review appears.
func (og OpsgenieClient) PostReviewNote(ctx context.Context, alertID string, review types.AccessReview) error {
	note, err := og.buildReviewNoteBody(review)
	if err != nil {
		return trace.Wrap(err)
	}
	body := AlertNote{
		Note: note,
	}
	if _, err := og.client.NewRequest().
		SetContext(ctx).
		SetBody(body).
		SetPathParams(map[string]string{"alertID": alertID}).
		SetQueryParams(map[string]string{"identifierType": "id"}).
		Post("alerts/{alertID}/notes"); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// ResolveAlert resolves an alert and posts a note with resolution details.
func (og OpsgenieClient) ResolveAlert(ctx context.Context, alertID string, resolution Resolution) error {
	note, err := og.buildResolutionNoteBody(resolution)
	if err != nil {
		return trace.Wrap(err)
	}
	body := AlertNote{
		Note: note,
	}
	if _, err := og.client.NewRequest().
		SetContext(ctx).
		SetBody(body).
		SetPathParams(map[string]string{"alertID": alertID}).
		SetQueryParams(map[string]string{"identifierType": "id"}).
		Post("alerts/{alertID}/close"); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

func (og OpsgenieClient) buildAlertBody(reqID string, reqData RequestData) (string, error) {
	var requestLink string
	if og.webProxyURL != nil {
		reqURL := *og.webProxyURL
		reqURL.Path = lib.BuildURLPath("web", "requests", reqID)
		requestLink = reqURL.String()
	}

	var builder strings.Builder
	err := alertBodyTemplate.Execute(&builder, struct {
		ID          string
		TimeFormat  string
		RequestLink string
		RequestData
	}{
		reqID,
		time.RFC822,
		requestLink,
		reqData,
	})
	if err != nil {
		return "", trace.Wrap(err)
	}
	return builder.String(), nil
}

func (og OpsgenieClient) buildReviewNoteBody(review types.AccessReview) (string, error) {
	var builder strings.Builder
	err := reviewNoteTemplate.Execute(&builder, struct {
		types.AccessReview
		ProposedState string
		TimeFormat    string
	}{
		review,
		review.ProposedState.String(),
		time.RFC822,
	})
	if err != nil {
		return "", trace.Wrap(err)
	}
	return builder.String(), nil
}

func (og OpsgenieClient) buildResolutionNoteBody(resolution Resolution) (string, error) {
	var builder strings.Builder
	err := resolutionNoteTemplate.Execute(&builder, struct {
		Resolution    string
		ResolveReason string
	}{
		string(resolution.Tag),
		resolution.Reason,
	})
	if err != nil {
		return "", trace.Wrap(err)
	}
	return builder.String(), nil
}

// GetOnCall returns the list of responders on-call for a schedule.
func (og OpsgenieClient) GetOnCall(ctx context.Context, scheduleName string) (RespondersResult, error) {
	// v2/schedules/ScheduleName/on-calls?scheduleIdentifierType=name&flat=true'

	var result RespondersResult
	if _, err := og.client.NewRequest().
		SetContext(ctx).
		SetPathParams(map[string]string{"scheduleName": scheduleName}).
		SetQueryParams(map[string]string{
			"scheduleIdentifierType": "name",
			"flat":                   "true",
		}).
		SetResult(&result).
		Post("schedules/{scheduleName}/on-calls"); err != nil {
		return RespondersResult{}, trace.Wrap(err)
	}
	return result, nil
}
