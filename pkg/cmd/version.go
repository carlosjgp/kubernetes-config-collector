package cmd

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// nolint: gochecknoglobals
// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Tag       string
	BuildUser string
	BuildDate string
	GoVersion = runtime.Version()
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of config-collector",
	Long:  `All software has versions. This is config-collector's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(BuildVersion())
	},
}

// Print returns version information.
func BuildVersion() string {
	return print("kubernetes-config-collector")
}

// versionInfoTmpl contains the template used by Info.
var versionInfoTmpl = `
{{.program}}, version {{.version}} (branch: {{.branch}}, revision: {{.revision}})
  build user:       {{.buildUser}}
  build date:       {{.buildDate}}
  go version:       {{.goVersion}}
`

func print(program string) string {
	m := map[string]string{
		"program":   program,
		"version":   Version,
		"revision":  Revision,
		"tag":       Tag,
		"buildUser": BuildUser,
		"buildDate": BuildDate,
		"goVersion": GoVersion,
	}
	t := template.Must(template.New("version").Parse(versionInfoTmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "version", m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

// Info returns version, branch and revision information.
func Info() string {
	return fmt.Sprintf("(version=%s, tag=%s, revision=%s)", Version, Tag, Revision)
}

// BuildContext returns goVersion, buildUser and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("(go=%s, user=%s, date=%s)", GoVersion, BuildUser, BuildDate)
}
