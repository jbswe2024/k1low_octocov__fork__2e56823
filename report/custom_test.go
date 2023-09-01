package report

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tenntenn/golden"
)

func TestCustomMetricSetTable(t *testing.T) {
	tests := []struct {
		s *CustomMetricSet
	}{
		{&CustomMetricSet{}},
		{&CustomMetricSet{
			Key:  "benchmark_0",
			Name: "Benchmark-0",
			Metrics: []*CustomMetric{
				{Key: "count", Name: "Count", Value: 1000.0, Unit: ""},
				{Key: "ns_per_op", Name: "ns/op", Value: 676.0, Unit: "ns/op"},
			},
		}},
		{&CustomMetricSet{
			Key:  "benchmark_1",
			Name: "Benchmark-1",
			Metrics: []*CustomMetric{
				{Key: "count", Name: "Count", Value: 1500.0, Unit: ""},
				{Key: "ns_per_op", Name: "ns/op", Value: 1340.0, Unit: "ns/op"},
			},
		}},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.s.Table()
			f := filepath.Join("custom_metrics", fmt.Sprintf("custom_metric_set_table.%d", i))
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(t), f, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(t), f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCustomMetricSetOut(t *testing.T) {
	tests := []struct {
		s *CustomMetricSet
	}{
		{&CustomMetricSet{}},
		{&CustomMetricSet{
			Key:  "benchmark_0",
			Name: "Benchmark-0",
			Metrics: []*CustomMetric{
				{Key: "count", Name: "Count", Value: 1000.0, Unit: ""},
				{Key: "ns_per_op", Name: "ns/op", Value: 676.0, Unit: "ns/op"},
			},
			report: &Report{
				Ref:      "main",
				Commit:   "1234567890",
				covPaths: []string{"testdata/cover.out"},
			},
		}},
		{&CustomMetricSet{
			Key:  "benchmark_1",
			Name: "Benchmark-1",
			Metrics: []*CustomMetric{
				{Key: "count", Name: "Count", Value: 1500.0, Unit: ""},
				{Key: "ns_per_op", Name: "ns/op", Value: 1340.0, Unit: "ns/op"},
			},
			report: &Report{
				Ref:      "main",
				Commit:   "1234567890",
				covPaths: []string{"testdata/cover.out"},
			},
		}},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := new(bytes.Buffer)
			if err := tt.s.Out(got); err != nil {
				t.Fatal(err)
			}
			f := filepath.Join("custom_metrics", fmt.Sprintf("custom_metric_set_out.%d", i))
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testdataDir(t), f, got)
				return
			}
			if diff := golden.Diff(t, testdataDir(t), f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}