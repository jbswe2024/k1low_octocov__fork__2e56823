package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/k1LoW/octocov/pkg/coverage"
	"github.com/k1LoW/octocov/pkg/ratio"
	"github.com/k1LoW/octocov/report"
)

func TestMain(m *testing.M) {
	envCache := os.Environ()

	m.Run()

	if err := revertEnv(envCache); err != nil {
		panic(err)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		wd      string
		path    string
		wantErr bool
	}{
		{testdataDir(t), "", false},
		{filepath.Join(testdataDir(t), "config"), "", false},
		{filepath.Join(testdataDir(t), "config"), ".octocov.yml", false},
		{filepath.Join(testdataDir(t), "config"), "no.yml", true},
	}
	for _, tt := range tests {
		c := New()
		c.wd = tt.wd
		if err := c.Load(tt.path); err != nil {
			if !tt.wantErr {
				t.Errorf("got %v\nwantErr %v", err, tt.wantErr)
			}
		} else {
			if tt.wantErr {
				t.Errorf("got %v\nwantErr %v", nil, tt.wantErr)
			}
		}
	}
}

func TestDatasourceGithubPath(t *testing.T) {
	if err := clearEnv(); err != nil {
		t.Fatal(err)
	}
	os.Setenv("GITHUB_REPOSITORY", "foo/bar")

	c := New()
	c.Datastore = &ConfigDatastore{
		Github: &ConfigDatastoreGithub{
			Repository: "report/dest",
		},
	}

	c.Build()
	if got := c.DatastoreConfigReady(); got != true {
		t.Errorf("got %v\nwant %v", got, true)
	}
	if err := c.BuildDatastoreConfig(); err != nil {
		t.Fatal(err)
	}
	want := "reports/foo/bar/report.json"
	if got := c.Datastore.Github.Path; got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestCoverageAcceptable(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"60%", true},
		{"50%", false},
		{"49.9%", false},
		{"49.9", false},
	}
	for _, tt := range tests {
		c := New()
		c.Coverage.Acceptable = tt.in
		c.Build()

		r := report.New()
		r.Coverage = &coverage.Coverage{
			Covered: 50,
			Total:   100,
		}
		if err := c.Acceptable(r); err != nil {
			if !tt.wantErr {
				t.Errorf("got %v\nwantErr %v", err, tt.wantErr)
			}
		} else {
			if tt.wantErr {
				t.Errorf("got %v\nwantErr %v", nil, tt.wantErr)
			}
		}
	}
}

func TestCodeToTestRatioAcceptable(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"1:1", false},
		{"1:1.1", true},
		{"1", false},
		{"1.1", true},
	}
	for _, tt := range tests {
		c := New()
		c.CodeToTestRatio = &ConfigCodeToTestRatio{
			Acceptable: tt.in,
			Test:       []string{"*_test.go"},
		}
		c.Build()
		r := report.New()
		r.CodeToTestRatio = &ratio.Ratio{
			Code: 100,
			Test: 100,
		}
		if err := c.Acceptable(r); err != nil {
			if !tt.wantErr {
				t.Errorf("got %v\nwantErr %v", err, tt.wantErr)
			}
		} else {
			if tt.wantErr {
				t.Errorf("got %v\nwantErr %v", nil, tt.wantErr)
			}
		}
	}
}

func TestTestExecutionTimeAcceptable(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"1min", false},
		{"59s", true},
		{"61sec", false},
	}
	for _, tt := range tests {
		c := New()
		c.TestExecutionTime = &ConfigTestExecutionTime{
			Acceptable: tt.in,
		}
		c.Build()
		r := report.New()
		e := float64(time.Minute)
		r.TestExecutionTime = &e
		if err := c.Acceptable(r); err != nil {
			if !tt.wantErr {
				t.Errorf("got %v\nwantErr %v", err, tt.wantErr)
			}
		} else {
			if tt.wantErr {
				t.Errorf("got %v\nwantErr %v", nil, tt.wantErr)
			}
		}
	}
}

func TestTraverseGitPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "a", "b", "c", "d"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "a", "b", ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, "a", "b", ".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	tests := []struct {
		base    string
		wantErr bool
	}{
		{filepath.Join(dir, "a", "b", "c"), false},
		{filepath.Join(dir, "a", "b", "c", "d"), false},
		{filepath.Join(dir, "a", "b"), false},
		{filepath.Join(dir, "a"), true},
	}
	for _, tt := range tests {
		got, err := traverseGitPath(tt.base)
		if err != nil {
			if !tt.wantErr {
				t.Errorf("got %v\nwantErr %v", err, tt.wantErr)
			}
		} else {
			if tt.wantErr {
				t.Errorf("got %v\nwantErr %v", nil, tt.wantErr)
			}
			if want := filepath.Join(dir, "a", "b"); got != want {
				t.Errorf("got %v\nwant %v", got, want)
			}
		}
	}
}

func revertEnv(envCache []string) error {
	if err := clearEnv(); err != nil {
		return err
	}
	for _, e := range envCache {
		splitted := strings.Split(e, "=")
		if err := os.Setenv(splitted[0], splitted[1]); err != nil {
			return err
		}
	}
	return nil
}

func clearEnv() error {
	for _, e := range os.Environ() {
		splitted := strings.Split(e, "=")
		if err := os.Unsetenv(splitted[0]); err != nil {
			return err
		}
	}
	return nil
}

func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := filepath.Abs(filepath.Join(filepath.Dir(wd), "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
