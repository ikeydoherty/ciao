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
	"sync"

	"github.com/01org/ciao/ssntp/uuid"
)

// Component is a tracing identifier for each ciao component.
type Component string

const (
	// Anonymous is a special component for anonymous tracing.
	// Anonymous tracing can only carry span messages but not
	// component specific payloads.
	Anonymous Component = "anonymous"

	// SSNTP is for tracing SSNTP traffic.
	SSNTP Component = "ssntp"

	// Libsnnet is for tracing ciao's networking.
	Libsnnet Component = "libsnnet"
)

const nullUUID = "00000000-0000-0000-0000-000000000000"
const spanChannelDepth = 256

type status uint8

const (
	running status = iota
	stopped
)

type tracerStatus struct {
	sync.Mutex
	status status
}

// Tracer is a handle to a ciao tracing agent that will collect
// local spans and send them back to ciao trace collectors.
type Tracer struct {
	ssntpUUID uuid.UUID
	component Component
	spanner   Spanner

	spanChannel chan Span
	stopChannel chan struct{}

	caCert string
	cert   string

	status tracerStatus
}

// TracerConfig represents a tracer configuration.
// This structure is parsed when creating a new tracer
// with NewTracer().
type TracerConfig struct {
	// UUID is the caller SSNTP UUID.
	UUID string

	// Component is the tracer creator component, e.g. "SSNTP"
	// or "Libsnnet". If this string is empty, the tracer will
	// be anonymous.
	Component Component

	// Spanner is a component specific span constructor.
	Spanner Spanner

	// CACert is the Certification Authority certificate path
	// to use when verifiying the peer identity.
	CAcert string

	// Cert is the tracer x509 signed certificate path.
	Cert string
}

// Context is an opaque structure that gets passed to Trace()
// calls in order to link spans together.
// If you want to link a span A to span B, you should pass the
// trace context returned when calling Trace() to create span A to
// the Trace() call for creating span B.
type Context struct {
	parentUUID uuid.UUID
}

// NewTracer creates a new tracer.
//
func NewTracer(config *TracerConfig) (*Tracer, *Context, error) {
	if config.UUID == "" {
		return nil, nil, fmt.Errorf("Empty SSNTP UUID")
	}

	if config.CAcert == "" {
		return nil, nil, fmt.Errorf("Missing CA")
	}

	if config.Cert == "" {
		return nil, nil, fmt.Errorf("Missing private key")
	}

	if config.Component == "" {
		config.Component = Anonymous
	}

	if config.Spanner == nil {
		config.Spanner = AnonymousSpanner{}
	}

	rootUUID, err := uuid.Parse(nullUUID)
	if err != nil {
		return nil, nil, err
	}

	ssntpUUID, err := uuid.Parse(config.UUID)
	if err != nil {
		return nil, nil, err
	}

	tracer := Tracer{
		ssntpUUID:   ssntpUUID,
		component:   config.Component,
		spanner:     config.Spanner,
		spanChannel: make(chan Span, spanChannelDepth),
		stopChannel: make(chan struct{}),
		caCert:      config.CAcert,
		cert:        config.Cert,
	}
	tracer.status.status = running

	traceContext := Context{
		parentUUID: rootUUID,
	}

	go tracer.spanListener()

	return &tracer, &traceContext, nil
}

func (t *Tracer) spanListener() {
	for {
		select {
		case span := <-t.spanChannel:
			// TODO Send spans to collectors
			fmt.Printf("SPAN: %s\n", span)
		case <-t.stopChannel:
			return
		}
	}
}

// Stop will stop a tracer.
// Spans will no longer be listened for and thus won't make
// it up to a trace collector.
func (t *Tracer) Stop() {
	defer t.status.Unlock()

	t.status.Lock()

	if t.status.status != running {
		return
	}

	t.status.status = stopped
	close(t.stopChannel)
}
