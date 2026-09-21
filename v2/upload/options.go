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
	"net/http"
)

// WithHTTPClient sets the HTTP client used for upload requests.
func WithHTTPClient(c *http.Client) UploaderOption {
	return func(up *Uploader) {
		up.hc = c
	}
}

// WithRequestURL sets the request URL for the initial start request.
func WithRequestURL(url string) UploaderOption {
	return func(up *Uploader) {
		up.startURL = url
	}
}

// WithChunkSize specifies the client's maximum chunk size in bytes.
// Effective chunk size is also a function of what the service specifies,
// so this is only an upper bound.
func WithChunkSize(size int64) UploaderOption {
	return func(up *Uploader) {
		up.clientChunkSize = size
	}
}
