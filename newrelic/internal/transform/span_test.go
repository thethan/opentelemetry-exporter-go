// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package transform

import (
	"reflect"
	"testing"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"go.opentelemetry.io/otel/api/core"
	"go.opentelemetry.io/otel/sdk/export/trace"
	"google.golang.org/grpc/codes"
)

const (
	service              = "myService"
	sampleTraceIDString  = "4bf92f3577b34da6a3ce929d0e0e4736"
	sampleSpanIDString   = "00f067aa0ba902b7"
	sampleParentIDString = "83887e5d7da921ba"
)

var (
	sampleTraceID, _  = core.TraceIDFromHex(sampleTraceIDString)
	sampleSpanID, _   = core.SpanIDFromHex(sampleSpanIDString)
	sampleParentID, _ = core.SpanIDFromHex(sampleParentIDString)
)

func TestTransformSpans(t *testing.T) {
	now := time.Now()
	testcases := []struct {
		testname string
		input    *trace.SpanData
		expect   telemetry.Span
	}{
		{
			testname: "basic span",
			input: &trace.SpanData{
				SpanContext: core.SpanContext{
					TraceID: sampleTraceID,
					SpanID:  sampleSpanID,
				},
				StartTime: now,
				EndTime:   now.Add(2 * time.Second),
				Name:      "mySpan",
			},
			expect: telemetry.Span{
				Name:        "mySpan",
				ID:          sampleSpanIDString,
				TraceID:     sampleTraceIDString,
				Timestamp:   now,
				Duration:    2 * time.Second,
				ServiceName: service,
				Attributes: map[string]interface{}{
					instrumentationProviderAttrKey: instrumentationProviderAttrValue,
					collectorNameAttrKey:           collectorNameAttrValue,
				},
			},
		},
		{
			testname: "span with parent",
			input: &trace.SpanData{
				SpanContext: core.SpanContext{
					TraceID: sampleTraceID,
					SpanID:  sampleSpanID,
				},
				ParentSpanID: sampleParentID,
				StartTime:    now,
				EndTime:      now.Add(2 * time.Second),
				Name:         "mySpan",
			},
			expect: telemetry.Span{
				Name:        "mySpan",
				ID:          sampleSpanIDString,
				TraceID:     sampleTraceIDString,
				ParentID:    sampleParentIDString,
				Timestamp:   now,
				Duration:    2 * time.Second,
				ServiceName: service,
				Attributes: map[string]interface{}{
					instrumentationProviderAttrKey: instrumentationProviderAttrValue,
					collectorNameAttrKey:           collectorNameAttrValue,
				},
			},
		},
		{
			testname: "span with error",
			input: &trace.SpanData{
				SpanContext: core.SpanContext{
					TraceID: sampleTraceID,
					SpanID:  sampleSpanID,
				},
				StatusCode:    codes.ResourceExhausted,
				StatusMessage: "ResourceExhausted",
				StartTime:     now,
				EndTime:       now.Add(2 * time.Second),
				Name:          "mySpan",
			},
			expect: telemetry.Span{
				Name:        "mySpan",
				ID:          sampleSpanIDString,
				TraceID:     sampleTraceIDString,
				Timestamp:   now,
				Duration:    2 * time.Second,
				ServiceName: service,
				Attributes: map[string]interface{}{
					instrumentationProviderAttrKey: instrumentationProviderAttrValue,
					collectorNameAttrKey:           collectorNameAttrValue,
					errorCodeAttrKey:               uint32(codes.ResourceExhausted),
					errorMessageAttrKey:            "ResourceExhausted",
				},
			},
		},
		{
			testname: "span with attributes",
			input: &trace.SpanData{
				SpanContext: core.SpanContext{
					TraceID: sampleTraceID,
					SpanID:  sampleSpanID,
				},
				StartTime: now,
				EndTime:   now.Add(2 * time.Second),
				Name:      "mySpan",
				Attributes: []core.KeyValue{
					core.Key("x0").Bool(true),
					core.Key("x1").Float32(1.0),
					core.Key("x2").Float64(2.0),
					core.Key("x3").Int(3),
					core.Key("x4").Int32(4),
					core.Key("x5").Int64(5),
					core.Key("x6").String("6"),
					core.Key("x7").Uint(7),
					core.Key("x8").Uint32(8),
					core.Key("x9").Uint64(9),
				},
			},
			expect: telemetry.Span{
				Name:        "mySpan",
				ID:          sampleSpanIDString,
				TraceID:     sampleTraceIDString,
				Timestamp:   now,
				Duration:    2 * time.Second,
				ServiceName: service,
				Attributes: map[string]interface{}{
					"x0":                           true,
					"x1":                           float32(1.0),
					"x2":                           float64(2.0),
					"x3":                           int64(3),
					"x4":                           int32(4),
					"x5":                           int64(5),
					"x6":                           "6",
					"x7":                           uint64(7),
					"x8":                           uint32(8),
					"x9":                           uint64(9),
					instrumentationProviderAttrKey: instrumentationProviderAttrValue,
					collectorNameAttrKey:           collectorNameAttrValue,
				},
			},
		},
		{
			testname: "span with attributes and error",
			input: &trace.SpanData{
				SpanContext: core.SpanContext{
					TraceID: sampleTraceID,
					SpanID:  sampleSpanID,
				},
				StatusCode:    codes.ResourceExhausted,
				StatusMessage: "ResourceExhausted",
				StartTime:     now,
				EndTime:       now.Add(2 * time.Second),
				Name:          "mySpan",
				Attributes: []core.KeyValue{
					core.Key("x0").Bool(true),
				},
			},
			expect: telemetry.Span{
				Name:        "mySpan",
				ID:          sampleSpanIDString,
				TraceID:     sampleTraceIDString,
				Timestamp:   now,
				Duration:    2 * time.Second,
				ServiceName: service,
				Attributes: map[string]interface{}{
					"x0":                           true,
					instrumentationProviderAttrKey: instrumentationProviderAttrValue,
					collectorNameAttrKey:           collectorNameAttrValue,
					errorCodeAttrKey:               uint32(codes.ResourceExhausted),
					errorMessageAttrKey:            "ResourceExhausted",
				},
			},
		},
	}
	for _, tc := range testcases {
		if got := Span(service, tc.input); !reflect.DeepEqual(got, tc.expect) {
			t.Errorf("%s: %#v != %#v", tc.testname, got, tc.expect)
		}
	}
}
