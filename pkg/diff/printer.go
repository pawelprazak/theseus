package diff

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yudai/gojsondiff/formatter"

	"github.com/heptio/theseus/pkg/source"
)

func PrintDeltas(diff source.ResourceSet, w io.Writer, color bool) {
	for _, v := range diff {
		if !v.Diff.Modified() {
			continue
		}

		formatter := formatter.NewAsciiFormatter(v.Object, formatter.AsciiFormatterConfig{Coloring: color})

		text, err := formatter.Format(v.Diff)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Fprintln(w, text)
	}
}

func PrintReportSummary(report *Report, w io.Writer) error {
	printSection := func(title string, resourceSet source.ResourceSet) error {
		header := fmt.Sprintf("%s (%v items)", title, len(resourceSet))

		// print header
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, strings.Repeat("-", len(header))); err != nil {
			return err
		}

		// sort item keys alphabetically
		var keys []source.ResourceKey
		for k := range resourceSet {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		// loop through them printing
		var currentScope, currentGVK string

		for _, key := range keys {
			scope, gvk, name, err := key.Parts()
			if err != nil {
				return err
			}

			// we're starting a new scope, so print the scope header
			if scope != currentScope {
				if _, err := fmt.Fprintln(w, scope); err != nil {
					return err
				}
			}

			// we're either starting a new scope or starting a new GVK, so
			// print the GVK header
			if scope != currentScope || gvk != currentGVK {
				if _, err := fmt.Fprintf(w, "\t%s\n", gvk); err != nil {
					return err
				}
			}

			// we always print the item
			if _, err := fmt.Fprintf(w, "\t\t%s\n", name); err != nil {
				return err
			}

			currentScope = scope
			currentGVK = gvk
		}

		return nil
	}

	if err := printSection("Left Only", report.LeftOnly); err != nil {
		return err
	}

	if err := printSection("Right Only", report.RightOnly); err != nil {
		return err
	}

	if err := printSection("Both", report.Both); err != nil {
		return err
	}

	return nil
}
