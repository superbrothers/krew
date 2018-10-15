// Copyright © 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/GoogleContainerTools/krew/pkg/installation"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	// listCmd represents the list command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed plugin names",
		Long: `List all installed plugin names.
Plugins will be shown as "PLUGIN,VERSION"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := installation.ListInstalledPlugins(paths.InstallPath(), paths.BinPath())
			if err != nil {
				return errors.Wrap(err, "failed to find all installed versions")
			}

			// return sorted list of plugin names when piped to other commands or file
			if !isTerminal(os.Stdout) {
				var names []string
				for name := range plugins {
					names = append(names, name)
				}
				sort.Strings(names)
				fmt.Fprintln(os.Stdout, strings.Join(names, "\n"))
				return nil
			}

			// print table
			var rows [][]string
			for p, version := range plugins {
				rows = append(rows, []string{p, version})
			}
			rows = sortByFirstColumn(rows)
			return printTable(os.Stdout, []string{"PLUGIN", "VERSION"}, rows)
		},
		PreRunE: checkIndex,
	}

	krewCmd.AddCommand(listCmd)
}

func printTable(out io.Writer, columns []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, strings.Join(columns, "\t"))
	fmt.Fprintln(w)
	for _, values := range rows {
		fmt.Fprintf(w, strings.Join(values, "\t"))
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func sortByFirstColumn(rows [][]string) [][]string {
	sort.Slice(rows, func(a, b int) bool {
		return rows[a][0] < rows[b][0]
	})
	return rows
}
