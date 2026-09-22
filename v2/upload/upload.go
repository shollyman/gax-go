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

// Package upload provides common functionality related to resumable media uploads
// over HTTP.
//
// It is EXPERIMENTAL and subject to change or removal without notice.
package upload

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/proto"
)

const (
	// Request and response headers used to communicate information related to a resumable upload.
	hdrProtocol     = "X-Goog-Upload-Protocol"
	hdrUploadURL    = "X-Goog-Upload-URL"
	hdrStatus       = "X-Goog-Upload-Status"
	hdrCommand      = "X-Goog-Upload-Command"
	hdrOffset       = "X-Goog-Upload-Offset"
	hdrSizeReceived = "X-Goog-Upload-Size-Received"
	hdrContentType  = "Content-Type"
	jsonContentType = "application/json"
)

// Typed string corresponding to X-Goog-Upload-Command directives.
type uploadCommand string

var (
	// Specific commands supported by the upload protocol.
	cmdStart    uploadCommand = "start"
	cmdUpload   uploadCommand = "upload"
	cmdFinalize uploadCommand = "finalize"
	cmdQuery    uploadCommand = "query"
)

type uploadStatus string

var (
	statusActive uploadStatus = "active"
	statusFinal  uploadStatus = "final"
)

type uploadProtocol string

var (
	resumableProtocol uploadProtocol = "resumable"
)

// Uploader is responsible for handling a specific upload session.
type Uploader struct {

	// current state handler.
	curHandler stateHandler

	// mutex guards changes to config and state.
	mu sync.Mutex

	// TODO: config fields
	hc        *http.Client
	startReq  proto.Message
	startURL  string
	uploadURL string

	clientChunkSize int64

	// TODO: runtime fields

}

// UploaderOption is an option pattern used to configure an Uploader.
// It's main usage is for controlling the instantiation of NewUploader.
type UploaderOption func(up *Uploader)

// NewUploader is used to instantiate a new Uploader, which supports uploading media
// to service endpoints that support the operation.
func NewUploader(ctx context.Context, opts ...UploaderOption) (*Uploader, error) {
	up := new(Uploader)
	// ensure we start with an initial state, though option processing and validation
	// can cause transition.
	up.transitionState(startStateID)
	for _, opt := range opts {
		opt(up)
	}
	err := up.validate(ctx)
	if err != nil {
		return nil, err
	}
	return up, nil
}

// handler validation and any setup of the uploader once options have
// been applied.
func (up *Uploader) validate(ctx context.Context) error {
	// TODO
	if up.curHandler == nil {
		return errors.New("no current handler")
	}
	if up.clientChunkSize < 0 {
		return errors.New("negative chunk size specified")
	}
	if up.hc == nil {
		return errors.New("no http client supplied")
	}
	return nil
}

// stateID is the typed identifier used to uniquely identify and lookup a state processor.
type stateID string

// stateHandler is the basic state processing interface all processors must satisfy.
// We use the empty interface for the time being, but can become more restrictive if needed.
type stateHandler interface {
}

// well known state IDs
var (
	// This sentinel value signals "no state change" for methods that return next state.
	noStateChange = stateID("")

	startStateID    = stateID("START")
	transmitStateID = stateID("TRANSMIT")
	terminalStateID = stateID("TERMINAL")
)

// stateRegistry is used to resolve states by ID.
// All states in the registry must have a unique state ID.
var stateRegistry = map[stateID]stateHandler{
	startStateID:    new(startState),
	transmitStateID: new(transmitState),
	terminalStateID: new(terminalState),
}

// transitionState updates the state handler of the uploader.
// If the noStateChange sentinel is passed, it does nothing.
// If an unknown state is passed, it panics.
func (up *Uploader) transitionState(nextState stateID) {
	if nextState == noStateChange {
		return
	}
	if s, ok := stateRegistry[nextState]; ok {
		up.curHandler = s
		return
	}
	panic(fmt.Sprintf("no such requested state %q", nextState))
}

// starter satisfies dispatched Start() requests.
type starter interface {
	// start sets up a new upload session,
	// and returns the next state or an error.
	start(ctx context.Context, up *Uploader, opts ...gax.CallOption) (err error, nextState stateID)
}

// Start is responsible for initializing a new upload session.
// Generally, this should only be called once to establish the initial upload session.
func (up *Uploader) Start(ctx context.Context, callOpts ...gax.CallOption) error {
	up.mu.Lock()
	defer up.mu.Unlock()
	if hdl, ok := up.curHandler.(starter); ok {
		err, nextState := hdl.start(ctx, up, callOpts...)
		up.transitionState(nextState)
		return err
	}
	return up.stateError()
}

// transmitter satisfies dispatched Write() requests.
type transmitter interface {
	// wraps the io.Writer contract
	write(ctx context.Context, up *Uploader, b []byte, callOpts ...gax.CallOption) (n int, err error, nextState stateID)
}

// Write transmits bytess as part of an upload.
func (up *Uploader) Write(ctx context.Context, b []byte, callOpts ...gax.CallOption) (n int, err error) {
	up.mu.Lock()
	defer up.mu.Unlock()
	if hdl, ok := up.curHandler.(transmitter); ok {
		n, err, nextState := hdl.write(ctx, up, b, callOpts...)
		up.transitionState(nextState)
		return n, err
	}
	return 0, up.stateError()
}

// remote information about an upload.
type remoteStatus struct {
	status       string
	sizeReceived int64
}

// queryer is used to get remote information about an upload.
type queryer interface {
	// fetch remote state of an upload.
	query(ctx context.Context, up *Uploader, callOpts ...gax.CallOption) (status *remoteStatus, err error, nextState stateID)
}

// finalizer satisfies dispatched Finalize() requests.
type finalizer interface {
	finalize(ctx context.Context, up *Uploader) (err error, nextState stateID)
}

// Finalize marks an upload as completed.
func (up *Uploader) Finalize(ctx context.Context, callOpts ...gax.CallOption) (err error) {
	up.mu.Lock()
	defer up.mu.Unlock()
	if hdl, ok := up.curHandler.(finalizer); ok {
		err, nextState := hdl.finalize(ctx, up)
		up.transitionState(nextState)
		return err
	}
	return up.stateError()
}

// canceller is used to explicit cancel an in progress upload.
type canceller interface {
	cancel(ctx context.Context, up *Uploader) (err error, nextState stateID)
}

// Cancel abandons an in progress upload.
func (up *Uploader) Cancel(ctx context.Context, callOpts ...gax.CallOption) (err error) {
	up.mu.Lock()
	defer up.mu.Unlock()
	if hdl, ok := up.curHandler.(canceller); ok {
		err, nextState := hdl.cancel(ctx, up)
		up.transitionState(nextState)
		return err
	}
	return up.stateError()
}

// Status provides information about the current upload.
// It provides information about the local upload state as well as remote state of the upload, if available.
type Status struct{}

// Status returns information about the current upload.
func (up *Uploader) Status(ctx context.Context) (*Status, error) {
	return nil, errors.New("unimplmented")
}

// Close closes the underlying client being used to manage the upload.
func (up *Uploader) Close() error {
	return errors.New("unimplemented")
}

// stateError provides a common error when a state machine doesn't support the dispatched method.
// TODO: report current state.
func (up *Uploader) stateError() error {
	return fmt.Errorf("operation not supported in current state")
}
