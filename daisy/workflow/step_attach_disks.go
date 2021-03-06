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

import "fmt"

// AttachDisks is a Daisy AttachDisks workflow step.
type AttachDisks []AttachDisk

// AttachDisk attaches a GCE disk to an instance.
type AttachDisk struct {
	Disk, Instance string
}

func (a *AttachDisks) validate(s *Step) error {
	for _, ad := range *a {
		if !diskValid(s.w, ad.Disk) {
			return fmt.Errorf("cannot attach disk. Disk not found: %s", ad.Disk)
		}
		if !instanceValid(s.w, ad.Instance) {
			return fmt.Errorf("cannot attach disk. Instance not found: %s", ad.Instance)
		}
	}

	return nil
}

func (a *AttachDisks) run(s *Step) error {
	return nil
}
