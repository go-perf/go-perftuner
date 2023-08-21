package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"golang.org/x/perf/benchstat"
)

type benchstatRunner struct{}

func (r *benchstatRunner) ExecCommand(_ context.Context, args []string) error {
	fset := flag.NewFlagSet("benchstat", flag.ContinueOnError)
	flagDeltaTest := fset.String("delta-test", "utest", "significance `test` to apply to delta: utest, ttest, or none")
	flagAlpha := fset.Float64("alpha", 0.05, "consider change significant if p < `α`")
	flagGeomean := fset.Bool("geomean", false, "print the geometric mean of each file")
	flagSplit := fset.String("split", "pkg,goos,goarch", "split benchmarks by `labels`")
	flagSort := fset.String("sort", "none", "sort by `order`: [-]delta, [-]name, none")
	noColor := fset.Bool("no-color", false, "disable the colored output")
	if err := fset.Parse(args); err != nil {
		return err
	}
	args = fset.Args()

	colorsEnabled := !*noColor

	var deltaTestNames = map[string]benchstat.DeltaTest{
		"none":   benchstat.NoDeltaTest,
		"u":      benchstat.UTest,
		"u-test": benchstat.UTest,
		"utest":  benchstat.UTest,
		"t":      benchstat.TTest,
		"t-test": benchstat.TTest,
		"ttest":  benchstat.TTest,
	}

	var sortNames = map[string]benchstat.Order{
		"none":  nil,
		"name":  benchstat.ByName,
		"delta": benchstat.ByDelta,
	}

	deltaTest := deltaTestNames[strings.ToLower(*flagDeltaTest)]
	if deltaTest == nil {
		return fmt.Errorf("invalid delta-test argument: %q", *flagDeltaTest)
	}

	reverse := false
	sortName := *flagSort
	if strings.HasPrefix(sortName, "-") {
		reverse = true
		sortName = sortName[1:]
	}

	order, ok := sortNames[sortName]
	if !ok {
		return fmt.Errorf("invalid sort argument: %q", sortName)
	}

	if len(fset.Args()) == 0 {
		// TODO(oleg): print command help here?
		return errors.New("expected at least 1 positional argument, the benchmarking target")
	}

	c := &benchstat.Collection{
		Alpha:      *flagAlpha,
		AddGeoMean: *flagGeomean,
		DeltaTest:  deltaTest,
	}

	if *flagSplit != "" {
		c.SplitBy = strings.Split(*flagSplit, ",")
	}

	if order != nil {
		if reverse {
			order = benchstat.Reverse(order)
		}
		c.Order = order
	}

	fmt.Printf("args: %+v", args)
	for _, file := range args {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		if err := c.AddFile(file, f); err != nil {
			return err
		}
		f.Close()
	}

	tables := c.Tables()
	r.fixBenchstatTables(tables)
	if colorsEnabled {
		r.colorizeBenchstatTables(tables)
	}

	var buf bytes.Buffer
	benchstat.FormatText(&buf, tables)
	os.Stdout.Write(buf.Bytes())

	return nil
}

func (r *benchstatRunner) fixBenchstatTables(tables []*benchstat.Table) {
	disabledGeomean := map[string]struct{}{}
	for _, table := range tables {
		selectedRows := table.Rows[:0]
		for _, row := range table.Rows {
			if row.PctDelta == 0 && strings.Contains(row.Delta, "0.00%") {
				// For whatever reason, sometimes we get +0.00% results
				// in delta which will be painted red. This is misleading.
				// Let's replace +0.00% with tilde.
				row.Delta = "~"
			}
			for _, m := range row.Metrics {
				for _, v := range m.RValues {
					if v < 0.01 {
						disabledGeomean[m.Unit] = struct{}{}
					}
				}
			}
			if row.Benchmark == "[Geo mean]" {
				if len(row.Metrics) != 0 {
					_, disabled := disabledGeomean[row.Metrics[0].Unit]
					if disabled {
						continue
					}
				}
			}
			selectedRows = append(selectedRows, row)
			if len(row.Metrics) == 0 {
				continue
			}
			if len(row.Metrics[0].RValues) < 5 && row.Benchmark != "[Geo mean]" {
				log.Printf("WARNING: %s needs more samples, re-run with -count=5 or higher?", row.Benchmark)
			}
		}
		table.Rows = selectedRows
	}
}

func (r *benchstatRunner) colorizeBenchstatTables(tables []*benchstat.Table) {
	for _, table := range tables {
		for _, row := range table.Rows {
			if r.isEpsilonDelta(row.Metrics) {
				row.Delta = r.yellowColorize("~")
				continue
			}

			d := r.calculateCombinedMeanDiff(row.Metrics)
			if r.isTinyValue(row.Metrics) {
				d *= 2 // For tiny values, require x2 precision.
			}

			d++
			if math.Abs(row.PctDelta) < d {
				row.Delta = r.yellowColorize("~")
				continue
			}

			switch {
			case strings.HasPrefix(row.Delta, "+"):
				row.Delta = r.redColorize(row.Delta)
			case strings.HasPrefix(row.Delta, "-"):
				row.Delta = r.greenColorize(row.Delta)
			default:
				row.Delta = r.yellowColorize(row.Delta)
			}
		}
	}
}

func (r *benchstatRunner) isEpsilonDelta(metrics []*benchstat.Metrics) bool {
	if len(metrics) != 2 {
		return false
	}

	eps := r.getValueEpsilon(r.avgValue(metrics))
	m0 := metrics[0].Mean
	m1 := metrics[1].Mean
	return math.Abs(m0-m1) <= eps
}

func (r *benchstatRunner) avgValue(metrics []*benchstat.Metrics) float64 {
	var sum float64
	for _, m := range metrics {
		sum += m.Mean
	}
	return sum / float64(len(metrics))
}

func (r *benchstatRunner) getValueEpsilon(avg float64) float64 {
	switch {
	case avg < 10:
		return 1
	case avg < 32:
		return 2
	case avg < 80:
		return 3
	default:
		return 4
	}
}

func (r *benchstatRunner) calculateCombinedMeanDiff(metrics []*benchstat.Metrics) float64 {
	var sum float64
	for _, m := range metrics {
		if m.Max != m.Min {
			sum += 100.0 * r.calculateMeanDiff(m)
		}
	}
	return sum
}

func (r *benchstatRunner) calculateMeanDiff(m *benchstat.Metrics) float64 {
	if m.Mean == 0 || m.Max == 0 {
		return 0
	}

	diff := 1 - m.Min/m.Mean
	if d := m.Max/m.Mean - 1; d > diff {
		diff = d
	}
	return diff
}

func (r *benchstatRunner) isTinyValue(metrics []*benchstat.Metrics) bool {
	const tinyValueThreshold = 32.0 // in nanosecs

	for _, m := range metrics {
		if m.Mean >= tinyValueThreshold {
			return false
		}
	}
	return true
}

func (r *benchstatRunner) redColorize(s string) string    { return "\033[31m" + s + "\033[0m" }
func (r *benchstatRunner) greenColorize(s string) string  { return "\033[32m" + s + "\033[0m" }
func (r *benchstatRunner) yellowColorize(s string) string { return "\033[33m" + s + "\033[0m" }
