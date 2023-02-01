/*
Copyright 2023 Gravitational, Inc.

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

package utils

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gravitational/trace"

	"github.com/gravitational/teleport"
)

// GetAndReplaceRequestBody returns the request body and replaces the drained
// body reader with io.NopCloser allowing for further body processing by http
// transport.
func GetAndReplaceRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.Body == http.NoBody {
		return []byte{}, nil
	}
	// req.Body is closed during tryDrainBody call.
	payload, err := tryDrainBody(req.Body)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	// Replace the drained body with io.NopCloser reader allowing for further request processing by HTTP transport.
	req.Body = io.NopCloser(bytes.NewReader(payload))
	return payload, nil
}

// GetAndReplaceResponseBody returns the response body and replaces the drained
// body reader with io.NopCloser allowing for further body processing.
func GetAndReplaceResponseBody(response *http.Response) ([]byte, error) {
	if response.Body == nil {
		return []byte{}, nil
	}

	payload, err := tryDrainBody(response.Body)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	response.Body = io.NopCloser(bytes.NewReader(payload))
	return payload, nil
}

// tryDrainBody tries to drain and close the body, returning the read bytes.
// It may fail to completely drain the body if the size of the body exceeds MaxHTTPRequestSize.
func tryDrainBody(b io.ReadCloser) (payload []byte, err error) {
	defer func() {
		if closeErr := b.Close(); closeErr != nil {
			err = trace.NewAggregate(err, closeErr)
		}
	}()
	payload, err = ReadAtMost(b, teleport.MaxHTTPRequestSize)
	if err != nil {
		err = trace.Wrap(err)
		return
	}
	return
}

// RenameHeader moves all values from the old header key to the new header key.
func RenameHeader(header http.Header, oldKey, newKey string) {
	if oldKey == newKey {
		return
	}
	for _, value := range header.Values(oldKey) {
		header.Add(newKey, value)
	}
	header.Del(oldKey)
}

// GetAnyHeader returns the first non-empty value by the provided keys.
func GetAnyHeader(header http.Header, keys ...string) string {
	for _, key := range keys {
		if value := header.Get(key); value != "" {
			return value
		}
	}
	return ""
}
