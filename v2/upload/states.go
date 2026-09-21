// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

// startState is the default initial state, corresponding to an uploader being configured but no
// setup RPCs for an upload being sent.
type startState struct{}

func (ss *startState) start(ctx context.Context, up *Uploader, opts ...gax.CallOption) (err error, nextState stateID) {
	if up.startReq == nil {
		return errors.New("no request supplied"), stateID("TERMINAL")
	}
	if up.startURL == "" {
		return errors.New("no start url supplied"), stateID("TERMINAL")
	}
	b, err := protojson.Marshal(up.startReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err), stateID("TERMINAL")
	}
	var resp *http.Response

	err = gax.Invoke(ctx, func(ctx context.Context, settings gax.CallSettings) error {
		// TODO: propagate other headers like ua etc via gax.BuildHeaders
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, up.startURL, bytes.NewReader(b))
		req.Header.Add(hdrProtocol, string(resumableProtocol))
		req.Header.Add(hdrCommand, string(cmdStart))
		req.Header.Add(hdrContentType, jsonContentType)

		// TODO: switch to executeHTTPRequest like a GAPIC
		resp, err = up.hc.Do(req)
		return err
	}, opts...)

	if err != nil {
		// We might still have a transient error here, so don't transition to terminal state.
		// TODO: consider how firm our "only call start once" might really be.
		return err, noStateChange
	}

	// may be too strict and/or unnecessary
	status := uploadStatus(resp.Header.Get(hdrStatus))
	if status != statusActive {
		return fmt.Errorf("upload status is not active, got %q", status), terminalStateID
	}

	upload := resp.Header.Get(hdrUploadURL)
	if upload == "" {
		return errors.New("did not get a valid upload URL during start"), terminalStateID
	}

	// Record the upload URL, indicate we're ready to transition to transmitting.
	up.uploadURL = up.uploadURL
	return nil, transmitStateID
}

// transmitState corresponds to an upload that has been configured and started, and able to send data.
type transmitState struct{}

// terminalState is the final terminal state.  It doesn't implement
// any dispatched methods, as the upload has concluded.
type terminalState struct{}
