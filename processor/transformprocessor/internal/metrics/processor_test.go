// Copyright  The OpenTelemetry Authors
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

package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

var (
	StartTime      = time.Date(2020, 2, 11, 20, 26, 12, 321, time.UTC)
	StartTimestamp = pcommon.NewTimestampFromTime(StartTime)
)

func TestProcess(t *testing.T) {
	tests := []struct {
		query string
		want  func(pmetric.Metrics)
	}{
		{
			query: `set(attributes["test"], "pass") where metric.name == "operationA"`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(1).Attributes().InsertString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where resource.attributes["host.name"] == "myhost"`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(1).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(1).Histogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(1).Histogram().DataPoints().At(1).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(1).Attributes().InsertString("test", "pass")
			},
		},
		{
			query: `keep_keys(attributes, "attr2") where metric.name == "operationA"`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).Attributes().Clear()
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(0).Attributes().InsertString("attr2", "test2")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(1).Attributes().Clear()
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().DataPoints().At(1).Attributes().InsertString("attr2", "test2")
			},
		},
		{
			query: `set(metric.description, "test") where attributes["attr1"] == "test1"`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetDescription("test")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(1).SetDescription("test")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).SetDescription("test")
			},
		},
		{
			query: `set(metric.unit, "new unit")`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetUnit("new unit")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(1).SetUnit("new unit")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).SetUnit("new unit")
			},
		},
		{
			query: `set(metric.description, "Sum") where metric.type == "Sum"`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).SetDescription("Sum")
			},
		},
		{
			query: `set(metric.aggregation_temporality, 1) where metric.aggregation_temporality == 0`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(1).Histogram().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().SetAggregationTemporality(pmetric.MetricAggregationTemporalityDelta)
			},
		},
		{
			query: `set(metric.is_monotonic, true) where metric.is_monotonic == false`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Sum().SetIsMonotonic(true)
			},
		},
		{
			query: `set(attributes["test"], "pass") where count == 1`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(1).Histogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where scale == 1`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where zero_count == 1`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where positive.offset == 1`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
			},
		},
		{
			query: `set(attributes["test"], "pass") where negative.offset == 1`,
			want: func(td pmetric.Metrics) {
				td.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(2).ExponentialHistogram().DataPoints().At(0).Attributes().InsertString("test", "pass")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			td := constructMetrics()
			processor, err := NewProcessor([]string{tt.query}, DefaultFunctions(), component.ProcessorCreateSettings{})
			assert.NoError(t, err)

			_, err = processor.ProcessMetrics(context.Background(), td)
			assert.NoError(t, err)

			exTd := constructMetrics()
			tt.want(exTd)

			assert.Equal(t, exTd, td)
		})
	}
}

func constructMetrics() pmetric.Metrics {
	td := pmetric.NewMetrics()
	rm0 := td.ResourceMetrics().AppendEmpty()
	rm0.Resource().Attributes().InsertString("host.name", "myhost")
	rm0ils0 := rm0.ScopeMetrics().AppendEmpty()
	fillMetricOne(rm0ils0.Metrics().AppendEmpty())
	fillMetricTwo(rm0ils0.Metrics().AppendEmpty())
	fillMetricThree(rm0ils0.Metrics().AppendEmpty())
	return td
}

func fillMetricOne(m pmetric.Metric) {
	m.SetName("operationA")
	m.SetDescription("operationA description")
	m.SetUnit("operationA unit")
	m.SetDataType(pmetric.MetricDataTypeSum)

	dataPoint0 := m.Sum().DataPoints().AppendEmpty()
	dataPoint0.SetStartTimestamp(StartTimestamp)
	dataPoint0.Attributes().InsertString("attr1", "test1")
	dataPoint0.Attributes().InsertString("attr2", "test2")
	dataPoint0.Attributes().InsertString("attr3", "test3")

	dataPoint1 := m.Sum().DataPoints().AppendEmpty()
	dataPoint1.SetStartTimestamp(StartTimestamp)
	dataPoint1.Attributes().InsertString("attr1", "test1")
	dataPoint1.Attributes().InsertString("attr2", "test2")
	dataPoint1.Attributes().InsertString("attr3", "test3")
}

func fillMetricTwo(m pmetric.Metric) {
	m.SetName("operationB")
	m.SetDescription("operationB description")
	m.SetUnit("operationB unit")
	m.SetDataType(pmetric.MetricDataTypeHistogram)

	dataPoint0 := m.Histogram().DataPoints().AppendEmpty()
	dataPoint0.SetStartTimestamp(StartTimestamp)
	dataPoint0.Attributes().InsertString("attr1", "test1")
	dataPoint0.Attributes().InsertString("attr2", "test2")
	dataPoint0.Attributes().InsertString("attr3", "test3")
	dataPoint0.SetCount(1)

	dataPoint1 := m.Histogram().DataPoints().AppendEmpty()
	dataPoint1.SetStartTimestamp(StartTimestamp)
	dataPoint1.Attributes().InsertString("attr1", "test1")
	dataPoint1.Attributes().InsertString("attr2", "test2")
	dataPoint1.Attributes().InsertString("attr3", "test3")
}

func fillMetricThree(m pmetric.Metric) {
	m.SetName("operationC")
	m.SetDescription("operationC description")
	m.SetUnit("operationC unit")
	m.SetDataType(pmetric.MetricDataTypeExponentialHistogram)

	dataPoint0 := m.ExponentialHistogram().DataPoints().AppendEmpty()
	dataPoint0.SetStartTimestamp(StartTimestamp)
	dataPoint0.Attributes().InsertString("attr1", "test1")
	dataPoint0.Attributes().InsertString("attr2", "test2")
	dataPoint0.Attributes().InsertString("attr3", "test3")
	dataPoint0.SetCount(1)
	dataPoint0.SetScale(1)
	dataPoint0.SetZeroCount(1)
	dataPoint0.Positive().SetOffset(1)
	dataPoint0.Negative().SetOffset(1)

	dataPoint1 := m.ExponentialHistogram().DataPoints().AppendEmpty()
	dataPoint1.SetStartTimestamp(StartTimestamp)
	dataPoint1.Attributes().InsertString("attr1", "test1")
	dataPoint1.Attributes().InsertString("attr2", "test2")
	dataPoint1.Attributes().InsertString("attr3", "test3")
}
