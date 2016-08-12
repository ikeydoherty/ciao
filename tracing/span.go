//
// Copyright (c) 2016 Intel Corporation
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
//

package trace

import (
	"fmt"
	"time"

	"github.com/docker/distribution/uuid"
)

type Span struct {
	uuid        uuid.UUID
	parentUUID  uuid.UUID
	creatorUUID uuid.UUID

	timestamp        time.Time
	component        Component
	componentPayload []byte
	message          string
}

type Spanner interface {
	Span(componentContext interface{}) []byte
}

func (s Span) String() string {
	return fmt.Sprintf("\n\tSpan UUID [%s]\n\tParent UUID [%s]\n\tTimestamp [%v]\n\tComponent [%s]\n\tMessage [%s]\n",
		s.uuid, s.parentUUID, s.timestamp, s.component, s.message)
}

type AnonymousSpanner struct{}

func (s AnonymousSpanner) Span(context interface{}) []byte {
	return nil
}
