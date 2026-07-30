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

import "testing"

func TestStatesAreUnique(t *testing.T) {
	// Assert the invariant that all registered state processors have a unique ID.
	// Prevent bugs due to re-use of state IDs.
	hasMap := make(map[stateID]struct{})
	for _, s := range stateList {
		curState := s.state()
		if _, ok := hasMap[curState]; ok {
			t.Errorf("multiple state processors with the same id: %q", curState)
		}
		hasMap[curState] = struct{}{}
	}
}

func TestGetStateProcessorBadInput(t *testing.T) {
	// Ensure garbage sent to getStateProcessor returns the unknown state.
	got := getStateProcessor(stateID("blahblahblah"))
	if _, ok := got.(*unknownState); !ok {
		t.Fatalf("getStateProcessor returned unexpected type %T", got)
	}
}
