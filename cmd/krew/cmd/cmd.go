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
	"flag"
	"os"

	"github.com/GoogleContainerTools/krew/pkg/environment"
	"github.com/GoogleContainerTools/krew/pkg/gitutil"
	isatty "github.com/mattn/go-isatty"
	"github.com/pkg/errors"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	paths               environment.Paths // krew paths used by the process
	krewExecutedVersion string            // resolved version of krew
)

// krewCmd represents the base command when called without any subcommands
var krewCmd = &cobra.Command{
	Use:   "krew",
	Short: "krew is the kubectl plugin manager",
	Long: `krew is the kubectl plugin manager.
You can invoke krew through kubectl with: "kubectl plugin [krew] option..."`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the krewCmd.
func Execute() {
	if err := krewCmd.Execute(); err != nil {
		if glog.V(1) {
			glog.Fatalf("%+v", err) // with stack trace
		} else {
			glog.Fatal(err) // just error message
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Set glog default to stderr
	flag.Set("logtostderr", "true")
	// Required by glog
	flag.Parse()
	paths = environment.MustGetKrewPaths()
	if err := ensureDirs(paths.BasePath(),
		paths.DownloadPath(),
		paths.InstallPath(),
		paths.BinPath()); err != nil {
		glog.Fatal(err)
	}

	selfPath, err := os.Executable()
	if err != nil {
		glog.Fatalf("failed to get the own executable path")
	}
	if krewVersion, ok, err := environment.GetExecutedVersion(paths.InstallPath(), selfPath, environment.Realpath); err != nil {
		glog.Fatalf("failed to find current krew version, err: %v", err)
	} else if ok {
		krewExecutedVersion = krewVersion
	}

	setGlogFlags(krewExecutedVersion != "")
}

// setGlogFlags will add glog flags to the CLI. This command can be executed multiple times.
func setGlogFlags(hidden bool) {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// Add glog flags if not run as a plugin.
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		pflag.Lookup(f.Name).Hidden = hidden
	})
}

func checkIndex(_ *cobra.Command, _ []string) error {
	if ok, err := gitutil.IsGitCloned(paths.IndexPath()); err != nil {
		return errors.Wrap(err, "failed to check local index git repository")
	} else if !ok {
		return errors.New(`krew local plugin index is not initialized (run "krew update")`)
	}
	return nil
}

func ensureDirs(paths ...string) error {
	for _, p := range paths {
		glog.V(4).Infof("Ensure creating dir: %q", p)
		if err := os.MkdirAll(p, 0755); err != nil {
			return errors.Wrapf(err, "failed to ensure create directory %q", p)
		}
	}
	return nil
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv()
}
