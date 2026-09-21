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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestStartState_start(t *testing.T) {

	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("handler failure: %v", err)
		}
		t.Logf("req: %q", string(b))
		for k, v := range req.Header {
			t.Logf("header %q: %v", k, v)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Add(hdrStatus, string(statusActive))
	}))
	defer serv.Close()

	up, err := NewUploader(t.Context(),
		WithHTTPClient(http.DefaultClient),
		WithRequestURL(serv.URL),
		WithRequest(new(emptypb.Empty{})),
	)
	if err != nil {
		t.Fatalf("NewUploader: %v", err)
	}

	ss := &startState{}
	err, next := ss.start(t.Context(), up)
	t.Errorf("got err: %v", err)
	t.Errorf("got next: %q", next)
}
