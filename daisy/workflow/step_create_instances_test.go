//  Copyright 2017 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package workflow

import (
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestCreateInstancesRun(t *testing.T) {
	w := testWorkflow()
	s := &Step{w: w}
	w.Sources = map[string]string{"file": "gs://some/file"}
	disks[w].m = map[string]*resource{
		"d1": {"d1", w.genName("d1"), "link", false, false},
		"d2": {"d2", w.genName("d2"), "link", false, false},
		"d3": {"d3", w.genName("d3"), "link", false, false},
	}
	ci := &CreateInstances{
		{Name: "i1", MachineType: "foo-type", AttachedDisks: []string{"d1"}, StartupScript: "file"},
		{Name: "i2", AttachedDisks: []string{"d2"}, Zone: "zone", Project: "project"},
		{Name: "i3", MachineType: "foo-type", AttachedDisks: []string{"d3"}, NoCleanup: true},
		{Name: "i4", MachineType: "foo-type", AttachedDisks: []string{"d3"}, ExactName: true},
		{Name: "i5", MachineType: "foo-type", AttachedDisks: []string{"zones/zone/disks/disk"}},
		{Name: "i6", AttachedDisks: []string{"d1"}, AttachedDisksRO: []string{"d2"}},
	}
	if err := ci.run(s); err != nil {
		t.Errorf("error running CreateInstances.run(): %v", err)
	}

	// Bad cases.
	badTests := []struct {
		name string
		ci   CreateInstances
		err  string
	}{
		{
			"disk DNE",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"dne"}}},
			"invalid or missing reference to AttachedDisk \"dne\"",
		},
		{
			"RO disk DNE",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"d1"}, AttachedDisksRO: []string{"dne"}}},
			"invalid or missing reference to AttachedDisk \"dne\"",
		},
	}

	for _, tt := range badTests {
		if err := tt.ci.run(s); err == nil {
			t.Errorf("%q: expected error, got nil", tt.name)
		} else if err.Error() != tt.err {
			t.Errorf("%q: did not get expected error from validate():\ngot: %q\nwant: %q", tt.name, err.Error(), tt.err)
		}
	}

	want := map[string]*resource{
		"i1": {"i1", w.genName("i1"), "link", false, false},
		"i2": {"i2", w.genName("i2"), "link", false, false},
		"i3": {"i3", w.genName("i3"), "link", true, false},
		"i4": {"i4", "i4", "link", false, false},
		"i5": {"i5", w.genName("i5"), "link", false, false},
		"i6": {"i6", w.genName("i6"), "link", false, false},
	}

	if diff := pretty.Compare(instances[w].m, want); diff != "" {
		t.Errorf("instanceRefs do not match expectation: (-got +want)\n%s", diff)
	}
}

func TestCreateInstancesValidate(t *testing.T) {
	// Set up.
	w := &Workflow{}
	s := &Step{w: w}
	validatedDisks = nameSet{w: {"d-foo", "d-bar"}}
	validatedInstances = nameSet{w: {"i-foo"}}
	w.Sources = map[string]string{"file": "gs://some/file"}

	// Good cases.
	goodTests := []struct {
		name string
		ci   CreateInstances
		want []string
	}{
		{
			"using multiple disks",
			CreateInstances{{Name: "i-boo", AttachedDisks: []string{"d-foo", "d-bar"}}},
			[]string{"i-foo", "i-boo"},
		},
		{
			"using read only disks",
			CreateInstances{{Name: "i-bar", AttachedDisks: []string{"d-foo"}, AttachedDisksRO: []string{"d-bar"}}},
			[]string{"i-foo", "i-boo", "i-bar"},
		},
		{
			"using StartupScript",
			CreateInstances{{Name: "i-bas", AttachedDisks: []string{"d-foo", "d-bar"}, StartupScript: "file"}},
			[]string{"i-foo", "i-boo", "i-bar", "i-bas"},
		},
		{
			"partial disk url no project",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"zones/zone/disks/disk"}, Zone: "zone", Project: "project"}},
			[]string{"i-foo", "i-boo", "i-bar", "i-bas", "i-baz"},
		},
		{
			"partial disk url",
			CreateInstances{{Name: "i-bax", AttachedDisks: []string{"projects/project/zones/zone/disks/disk"}, Zone: "zone", Project: "project"}},
			[]string{"i-foo", "i-boo", "i-bar", "i-bas", "i-baz", "i-bax"},
		},
	}

	for _, tt := range goodTests {
		if err := tt.ci.validate(s); err != nil {
			t.Errorf("%q: unexpected error: %v", tt.name, err)
		}
		if !reflect.DeepEqual(validatedInstances[w], tt.want) {
			t.Errorf("%q: got:(%v) != want(%v)", tt.name, validatedInstances[w], tt.want)
		}
	}

	// Bad cases.
	badTests := []struct {
		name string
		ci   CreateInstances
		err  string
	}{
		{
			"dupe name",
			CreateInstances{{Name: "i-bar", AttachedDisks: []string{"d-foo", "d-bar"}}},
			"error adding instance: workflow \"\" has duplicate references for \"i-bar\"",
		},
		{
			"StartupScript not in Sources",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"d-foo", "d-bar"}, StartupScript: "dne-file"}},
			"cannot create instance: file not found: dne-file",
		},
		{
			"no disks",
			CreateInstances{{Name: "i-baz"}},
			"cannot create instance: no disks provided",
		},
		{
			"disk DNE",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"d-foo", "d-bar", "d-dne"}}},
			"cannot create instance: disk not found: d-dne",
		},
		{
			"RO disk DNE",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"d-foo", "d-bar"}, AttachedDisksRO: []string{"d-dne"}}},
			"cannot create instance: disk not found: d-dne",
		},
		{
			"partial disk url wrong project",
			CreateInstances{{Name: "i-baz", AttachedDisks: []string{"projects/project1/zones/zone/disks/disk"}, Zone: "zone", Project: "project2"}},
			"cannot create instance in project \"project2\" with disk in project \"project1\": \"projects/project1/zones/zone/disks/disk\"",
		},
		{
			"partial disk url wrong zone",
			CreateInstances{{Name: "i-bax", AttachedDisks: []string{"projects/project/zones/zone1/disks/disk"}, Zone: "zone2", Project: "project"}},
			"cannot create instance in project \"zone2\" with disk in project \"zone1\": \"projects/project/zones/zone1/disks/disk\"",
		},
	}

	for _, tt := range badTests {
		if err := tt.ci.validate(s); err == nil {
			t.Errorf("%q: expected error, got nil", tt.name)
		} else if err.Error() != tt.err {
			t.Errorf("%q: did not get expected error from validate():\ngot: %q\nwant: %q", tt.name, err.Error(), tt.err)
		}
	}

	want := []string{"i-foo", "i-boo", "i-bar", "i-bas", "i-baz", "i-bax"}
	if !reflect.DeepEqual(validatedInstances[w], want) {
		t.Fatalf("got:(%v) != want(%v)", validatedInstances[w], want)
	}
}
