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
package upload

import (
	"context"
	"sync"
)

const (
	// Request and response headers used to manage a resumable upload.
	hdrProtocol     = "X-Goog-Upload-Protocol"
	hdrUploadURL    = "X-Goog-Upload-URL"
	hdrStatus       = "X-Goog-Upload-Status"
	hdrCommand      = "X-Goog-Upload-Command"
	hdrOffset       = "X-Goog-Upload-Offset"
	hdrSizeReceived = "X-Goog-Upload-Size-Received"
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

// Typed string corresponding to X-Goog-Upload-Protocol values
type uploadProtocol string

var (
	ResumableProtocol uploadProtocol = "resumable"
)

// Uploader is responsible for handling a specific upload.
type Uploader struct {
	// Uploader derives its own context.
	ctx      context.Context
	cancelFn context.CancelFunc

	// mutex guards changes to config or state.
	mu sync.RWMutex

	// TODO: config fields

	// TODO: runtime fields
}

// UploaderOption is an option type used to configure an uploader
// as part of the NewUploader function.
type UploaderOption func(up *Uploader)

func NewUploader(ctx context.Context, opts ...UploaderOption) (*Uploader, error) {
	up := new(Uploader)
	// Create a cancellable context.
	up.ctx, up.cancelFn = context.WithCancel(ctx)
	for _, opt := range opts {
		opt(up)
	}
	// TODO: handle basic uploader validation before returning it.
	return up, nil
}
