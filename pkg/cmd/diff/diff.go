package diff

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/heptio/ark/pkg/cmd/util/flag"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/heptio/theseus/pkg/diff"
	"github.com/heptio/theseus/pkg/source"
)

func NewCommand() *cobra.Command {
	var (
		outputDir      string
		includedScopes flag.StringArray
		labelSelector  flag.LabelSelector
	)

	c := &cobra.Command{
		Use: "diff [left-source] [right-source]",
		Run: func(cmd *cobra.Command, args []string) {
			options, err := getDiffOptions(args, includedScopes, labelSelector.LabelSelector)
			if err != nil {
				fmt.Println(err)
				return
			}

			report, err := diff.Generate(options)
			if err != nil {
				fmt.Println(err)
				return
			}

			// print line-by-line diffs to stdout with highlighting
			diff.PrintDeltas(report.Both, os.Stdout, true)

			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Println(err)
				return
			}

			// print summary (existence diffs) to txt file
			func() {
				reportFile, err := os.Create(fmt.Sprintf("%s/summary.txt", outputDir))
				if err != nil {
					fmt.Println(err)
					return
				}
				defer reportFile.Close()

				if err := diff.PrintReportSummary(report, reportFile); err != nil {
					fmt.Println(err)
					return
				}
			}()

			// print line-by-line diffs to txt file
			func() {
				diffFile, err := os.Create(fmt.Sprintf("%s/item-diffs.txt", outputDir))
				if err != nil {
					fmt.Println(err)
					return
				}
				defer diffFile.Close()
				diff.PrintDeltas(report.Both, diffFile, false)
			}()
		},
	}

	c.Flags().StringVar(&outputDir, "output-dir", ".", "directory where the diff reports should be created")
	c.Flags().Var(&includedScopes, "included-scopes", "scopes to include in the diff (use 'cluster' for cluster-scoped resources and <ns-name> for namespaces)")
	c.Flags().Var(&labelSelector, "selector", "only diff resources matching this label selector")

	return c
}

func getDiffOptions(args []string, includedScopes []string, labelSelector *metav1.LabelSelector) (*diff.Options, error) {
	if len(args) < 2 {
		return nil, errors.New("you must supply a left and right source config")
	}

	left, err := getSource(args[0])
	if err != nil {
		return nil, err
	}

	right, err := getSource(args[1])
	if err != nil {
		return nil, err
	}

	return &diff.Options{
		Left:          left,
		Right:         right,
		Scopes:        source.NewIncludes(includedScopes...),
		LabelSelector: labelSelector,
	}, nil
}

func getSource(arg string) (source.ResourceLister, error) {
	parts := strings.Split(arg, "=")
	if len(parts) != 2 {
		return nil, errors.New("format for source must be <type>=<location>")
	}

	return source.Get(parts[0], parts[1])
}
