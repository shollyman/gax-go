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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestOptions(t *testing.T) {

	for _, tc := range []struct {
		desc string
		opts []UploaderOption
		want *Uploader
	}{
		{
			desc: "no opts",
			want: new(Uploader),
		},
		{
			desc: "request url",
			opts: []UploaderOption{
				WithRequestURL("http://foo"),
			},
			want: &Uploader{
				startURL: "http://foo",
			},
		},
		{
			desc: "chunk size",
			opts: []UploaderOption{
				WithChunkSize(99),
			},
			want: &Uploader{
				clientChunkSize: 99,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			got := new(Uploader)
			for _, o := range tc.opts {
				o(got)
			}

			if diff := cmp.Diff(tc.want, got, cmp.AllowUnexported(Uploader{}), cmpopts.IgnoreFields(Uploader{}, "mu", "hc")); diff != "" {
				t.Errorf("Uploader mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
