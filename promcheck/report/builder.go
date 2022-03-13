package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

const (
	// DefaultFormat dumps CheckReports
	DefaultFormat = "graph"
	// YAMLFormat dumps CheckReports as YAML.
	YAMLFormat = "yaml"
	// JSONFormat dumps CheckReports as JSON.
	JSONFormat = "json"
	// PrometheusFormat pushes CheckReports as Prometheus metrics. todo(cbrgm): implement me
	PrometheusFormat = "prometheus"
)

// Builder represents the report
type Builder struct {

	// Report represents the report data
	Report Report `json:"promcheck" yaml:"promcheck"`

	// outputFormat represents the output format
	outputFormat string

	// outputTarget represents the output target
	// Default: stdout  todo(cbrgm): make me configurable
	outputTarget io.ReadWriteCloser
}

// NewBuilder returns a new Builder
func NewBuilder(outputFormat string, noColor bool) *Builder {
	color.NoColor = noColor
	if outputFormat == "" {
		outputFormat = DefaultFormat
	}
	return &Builder{
		Report:       Report{},
		outputFormat: outputFormat,
		outputTarget: os.Stdout,
	}
}

// Report represents report data
type Report struct {

	// Sections represents a list of result data
	Sections Sections `json:"results,omitempty" yaml:"results,omitempty"`

	// SectionsCount represents the length of Sections
	SectionsCount int `json:"rules_warnings,omitempty" yaml:"rules_warnings,omitempty"`

	// TotalRules represents the total amount of checked rules
	TotalRules int `json:"rules_total,omitempty" yaml:"rules_total,omitempty"`

	// TotalSelectorsFailed represents the total amount of probed selectors not containing a result value
	TotalSelectorsFailed int `json:"selectors_success_total,omitempty" yaml:"selectors_success_total,omitempty"`

	// TotalSelectorsSuccess represents the total amount of probed selectors containing a result value
	TotalSelectorsSuccess int `json:"selectors_failed_total,omitempty" yaml:"selectors_failed_total,omitempty"`

	// RatioFailedTotal represents the ratio of selectors without a result value / total amount of selectors
	RatioFailedTotal float32 `json:"ratio_failed_total,omitempty" yaml:"ratio_failed_total,omitempty"`
}

// Sections represents a collection of sections.
type Sections []Section

// Section represents a report section
type Section struct {

	// File represents the file name of the checked rule
	File string `json:"file" yaml:"file"`

	// Group represents the group name of the checked rule
	Group string `json:"group" yaml:"group"`

	// Name represents the recording rule or alert name
	Name string `json:"name" yaml:"name"`

	// Expression represents the rule's PromQL expression string
	Expression string `json:"expression" yaml:"expression"`

	// NoResults represents a list of the rule's PromQL selectors which did not successfully returned a result value
	NoResults []string `json:"no_results" yaml:"no_results"`

	// Results represents a list of the rule's PromQL selectors which successfully returned a result value
	Results []string `json:"results" yaml:"results"`
}

// Len returns the list size.
func (s Report) Len() int {
	return len(s.Sections)
}

// HasContent checks if we actually have anything to report.
func (b *Builder) HasContent() bool {
	return b.Report.SectionsCount != 0
}

func (b *Builder) finalize() {
	totalSelectors := b.Report.TotalSelectorsFailed + b.Report.TotalSelectorsSuccess
	b.Report.RatioFailedTotal = (float32(b.Report.TotalSelectorsFailed) / float32(totalSelectors)) * 100
}

func (b *Builder) AddSection(file, group, name, expression string, failed []string, success []string) {
	b.Report.Sections = append(b.Report.Sections, Section{
		File:       file,
		Group:      group,
		Name:       name,
		Expression: expression,
		NoResults:  failed,
		Results:    success,
	})
	b.Report.SectionsCount += 1
	b.Report.TotalSelectorsFailed += len(failed)
	b.Report.TotalSelectorsSuccess += len(success)
}

func (b *Builder) AddTotalCheckedRules(count int) {
	b.Report.TotalRules += count
}

func (b *Builder) ToYAML() (string, error) {
	b.finalize()
	raw, err := yaml.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (b *Builder) ToJSON() (string, error) {
	b.finalize()
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (b *Builder) Dump() error {
	if !b.HasContent() {
		return errors.New("nothing to report")
	}
	var err error
	switch b.outputFormat {
	case YAMLFormat:
		err = b.DumpYAML()
	case JSONFormat:
		err = b.DumpJSON()
	case DefaultFormat:
		err = b.DumpTree()
	default:
		err = b.DumpTree()
	}
	return err
}

func (b *Builder) DumpYAML() error {
	res, err := b.ToYAML()
	if err != nil {
		return err
	}
	fmt.Fprintf(b.outputTarget, "%v\n", res)
	return nil
}

func (b *Builder) DumpJSON() error {
	res, err := b.ToJSON()
	if err != nil {
		return err
	}
	fmt.Fprintf(b.outputTarget, "%v\n", res)
	return nil
}

func (b *Builder) DumpTree() error {
	res, err := b.ToTree()
	if err != nil {
		return err
	}
	fmt.Fprintf(b.outputTarget, "%v\n", res)
	return nil
}
