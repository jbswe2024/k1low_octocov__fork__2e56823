package coverage

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type Type string

const (
	TypeLOC  Type = "loc"
	TypeStmt Type = "statement"
)

type Coverage struct {
	Type    Type          `json:"type"`
	Format  string        `json:"format"`
	Total   int           `json:"total"`
	Covered int           `json:"covered"`
	Files   FileCoverages `json:"files"`
}

type FileCoverage struct {
	File    string         `json:"file"`
	Total   int            `json:"total"`
	Covered int            `json:"covered"`
	Blocks  BlockCoverages `json:"blocks,omitempty"`
	cache   map[int]BlockCoverages
}

type FileCoverages []*FileCoverage

type BlockCoverage struct {
	Type      Type `json:"type"`
	StartLine *int `json:"start_line,omitempty"`
	StartCol  *int `json:"start_col,omitempty"`
	EndLine   *int `json:"end_line,omitempty"`
	EndCol    *int `json:"end_col,omitempty"`
	NumStmt   *int `json:"num_stmt,omitempty"`
	Count     *int `json:"count,omitempty"`
}

type BlockCoverages []*BlockCoverage

type Processor interface {
	Name() string
	ParseReport(path string) (*Coverage, string, error)
}

func New() *Coverage {
	return &Coverage{
		Files: FileCoverages{},
	}
}

func NewFileCoverage(file string) *FileCoverage {
	return &FileCoverage{
		File:    file,
		Total:   0,
		Covered: 0,
		Blocks:  BlockCoverages{},
		cache:   map[int]BlockCoverages{},
	}
}

func (c *Coverage) FlushBlockCoverages() {
	for _, f := range c.Files {
		f.Blocks = BlockCoverages{}
	}
}

func (fcs FileCoverages) FindByFile(file string) (*FileCoverage, error) {
	for _, fc := range fcs {
		if fc.File == file {
			return fc, nil
		}
	}
	return nil, fmt.Errorf("file name not found: %s", file)
}

func (fcs FileCoverages) FuzzyFindByFile(file string) (*FileCoverage, error) {
	for _, fc := range fcs {
		if strings.Contains(strings.TrimLeft(fc.File, "./"), strings.TrimLeft(file, "./")) {
			return fc, nil
		}
	}
	return nil, fmt.Errorf("file name not found: %s", file)
}

func (fcs FileCoverages) PathPrefix() (string, error) {
	if len(fcs) == 0 {
		return "", errors.New("no file coverages")
	}
	p := strings.Split(filepath.Dir(filepath.ToSlash(fcs[0].File)), "/")
	for _, fc := range fcs {
		d := strings.Split(filepath.Dir(filepath.ToSlash(fc.File)), "/")
		i := 0
		for {
			if len(p) <= i {
				break
			}
			if len(d) <= i {
				break
			}
			if p[i] != d[i] {
				break
			}
			i += 1
		}
		p = p[:i]
	}
	s := strings.Join(p, "/")
	if s == "" && strings.HasPrefix(fcs[0].File, "/") {
		s = "/"
	}
	if s == "." {
		s = ""
	}
	return s, nil
}

func (fc *FileCoverage) FindBlocksByLine(n int) BlockCoverages {
	if fc == nil {
		return BlockCoverages{}
	}
	if len(fc.cache) == 0 {
		fc.cache = map[int]BlockCoverages{}
		for _, b := range fc.Blocks {
			for i := *b.StartLine; i <= *b.EndLine; i++ {
				_, ok := fc.cache[i]
				if !ok {
					fc.cache[i] = BlockCoverages{}
				}
				fc.cache[i] = append(fc.cache[i], b)
			}
		}
	}
	blocks, ok := fc.cache[n]
	if ok {
		return blocks
	} else {
		return BlockCoverages{}
	}
}

func (dfcs DiffFileCoverages) FuzzyFindByFile(file string) (*DiffFileCoverage, error) {
	for _, dfc := range dfcs {
		if strings.Contains(strings.TrimLeft(dfc.File, "./"), strings.TrimLeft(file, "./")) {
			return dfc, nil
		}
	}
	return nil, fmt.Errorf("file name not found: %s", file)
}

func (bcs BlockCoverages) MaxCount() int {
	c := map[int]int{}
	for _, bc := range bcs {
		sl := *bc.StartLine
		el := *bc.EndLine
		for i := sl; i <= el; i++ {
			_, ok := c[i]
			if !ok {
				c[i] = 0
			}
			c[i] += *bc.Count
		}
	}
	max := 0
	for _, v := range c {
		if v > max {
			max = v
		}
	}
	return max
}
