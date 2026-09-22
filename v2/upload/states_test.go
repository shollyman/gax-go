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
	"net/http/httptest"
	"testing"

	"google.golang.org/protobuf/types/known/emptypb"
)

func TestStartState_start(t *testing.T) {

	// Create a test server and handler that satisfies validates basic request validation
	// and provides a proper start response.
	serv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if meth := req.Method; meth != http.MethodPost {
			t.Errorf("expected POST, got %q", meth)
		}
		if cmd := req.Header.Get(hdrCommand); cmd != string(cmdStart) {
			t.Errorf("expected header %q to be start, got %q", hdrCommand, cmd)
		}
		if prot := req.Header.Get(hdrProtocol); prot != string(resumableProtocol) {
			t.Errorf("expected header %q to be resumable, got %q", hdrProtocol, prot)
		}
		w.Header().Add(hdrStatus, string(statusActive))
		w.Header().Add(hdrUploadURL, "http://foo/1234")
		w.WriteHeader(http.StatusOK)

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
	if err != nil {
		t.Errorf("start() error: %v", err)
	}
	if next != transmitStateID {
		t.Errorf("start() got next state %q, want %q", next, transmitStateID)
	}
}
