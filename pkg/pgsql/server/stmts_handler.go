/*
Copyright 2022 Codenotary Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"regexp"

	pserr "github.com/codenotary/immudb/pkg/pgsql/errors"
)

var set = regexp.MustCompile(`(?i)set\s+.+`)
var selectVersion = regexp.MustCompile(`(?i)select\s+version\(\s*\)`)

func (s *session) isInBlackList(statement string) bool {
	if set.MatchString(statement) {
		return true
	}
	if statement == ";" {
		return true
	}
	return false
}

func (s *session) isEmulableInternally(statement string) interface{} {
	if selectVersion.MatchString(statement) {
		return &version{}
	}
	return nil
}
func (s *session) tryToHandleInternally(command interface{}) error {
	switch command.(type) {
	case *version:
		if err := s.writeVersionInfo(); err != nil {
			return err
		}
	default:
		return pserr.ErrMessageCannotBeHandledInternally
	}
	return nil
}

type version struct{}
