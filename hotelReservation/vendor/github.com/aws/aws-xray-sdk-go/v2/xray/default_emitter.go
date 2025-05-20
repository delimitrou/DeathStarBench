// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"encoding/json"
	"net"
	"runtime/debug"
	"sync"

	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"
)

// Header is added before sending segments to daemon.
const Header = `{"format": "json", "version": 1}` + "\n"

// DefaultEmitter provides the naive implementation of emitting trace entities.
type DefaultEmitter struct {
	sync.Mutex
	conn *net.UDPConn
	addr *net.UDPAddr
}

// NewDefaultEmitter initializes and returns a
// pointer to an instance of DefaultEmitter.
func NewDefaultEmitter(raddr *net.UDPAddr) (*DefaultEmitter, error) {
	initLambda()
	d := &DefaultEmitter{addr: raddr}
	return d, nil
}

// RefreshEmitterWithAddress dials UDP based on the input UDP address.
func (de *DefaultEmitter) RefreshEmitterWithAddress(raddr *net.UDPAddr) {
	de.Lock()
	de.refresh(raddr)
	de.Unlock()
}

func (de *DefaultEmitter) refresh(raddr *net.UDPAddr) (err error) {
	de.conn, err = net.DialUDP("udp", nil, raddr)
	de.addr = raddr

	if err != nil {
		logger.Errorf("Error dialing emitter address %v: %s", raddr, err)
		return err
	}

	logger.Infof("Emitter using address: %v", raddr)
	return nil
}

// Emit segment or subsegment if root segment is sampled.
// seg has a write lock acquired by the caller.
func (de *DefaultEmitter) Emit(seg *Segment) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic emitting segment: %s\n%s", r, string(debug.Stack()))
		}
	}()
	HeaderBytes := []byte(Header)

	if seg == nil || !seg.ParentSegment.Sampled {
		return
	}

	for _, p := range packSegments(seg, nil) {
		logger.Debug(string(p))

		de.Lock()

		if de.conn == nil {
			if err := de.refresh(de.addr); err != nil {
				de.Unlock()
				return
			}
		}

		_, err := de.conn.Write(append(HeaderBytes, p...))
		if err != nil {
			logger.Error(err)
		}
		de.Unlock()
	}
}

// seg has a write lock acquired by the caller.
func packSegments(seg *Segment, outSegments [][]byte) [][]byte {
	trimSubsegment := func(s *Segment) []byte {
		ss := globalCfg.StreamingStrategy()
		if seg.ParentSegment.Configuration != nil && seg.ParentSegment.Configuration.StreamingStrategy != nil {
			ss = seg.ParentSegment.Configuration.StreamingStrategy
		}
		for ss.RequiresStreaming(s) {
			if len(s.rawSubsegments) == 0 {
				break
			}
			cb := ss.StreamCompletedSubsegments(s)
			outSegments = append(outSegments, cb...)
		}
		b, err := json.Marshal(s)
		if err != nil {
			logger.Errorf("JSON error while marshalling (Sub)Segment: %v", err)
		}
		return b
	}

	for _, s := range seg.rawSubsegments {
		s.Lock()
		outSegments = packSegments(s, outSegments)
		if b := trimSubsegment(s); b != nil {
			seg.Subsegments = append(seg.Subsegments, b)
		}
		s.Unlock()
	}
	if seg.isOrphan() {
		if b := trimSubsegment(seg); b != nil {
			outSegments = append(outSegments, b)
		}
	}
	return outSegments
}
