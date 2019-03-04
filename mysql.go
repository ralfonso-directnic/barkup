package barkup

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	// TarCmd is the path to the `tar` executable
	TarCmd = "tar"
	// MysqlDumpCmd is the path to the `mysqldump` executable
	MysqlDumpCmd = "mysqldump"
)

// MySQL is an `Exporter` interface that backs up a MySQL database via the `mysqldump` command
type MySQL struct {
	// DB Host (e.g. 127.0.0.1)
	Host string
	// DB Port (e.g. 3306)
	Port string
	// DB Name
	DB string
	// DB User
	User string
	// DB Password
	Password string
	// Extra mysqldump options
	// e.g []string{"--extended-insert"}
	Options []string
}

// Export produces a `mysqldump` of the specified database, and creates a gzip compressed tarball archive.
func (x MySQL) Export() *ExportResult {
	result := &ExportResult{MIME: "application/x-7z-compressed"}

	fileSuffix := time.Now().Format("2006_01_02_15_04_05")
	dumpPath := fmt.Sprintf(`%v_%s.sql`, x.DB, fileSuffix)

	options := append(x.dumpOptions(), fmt.Sprintf(`-r%v`, dumpPath))
	out, err := exec.Command(MysqlDumpCmd, options...).CombinedOutput()
	if err != nil {
		result.Error = makeErr(err, string(out))
		return result
	}

	result.Path = dumpPath + ".7z"

	_, err = exec.Command("7za", "a", result.Path, dumpPath).CombinedOutput()
	if err != nil {
		os.Remove(dumpPath)

		result.Error = makeErr(err, string(out))
		return result
	}
	os.Remove(dumpPath)

	return result
}

func (x MySQL) dumpOptions() []string {
	options := x.Options
	options = append(options, fmt.Sprintf(`-h%v`, x.Host))
	options = append(options, fmt.Sprintf(`-P%v`, x.Port))
	options = append(options, fmt.Sprintf(`-u%v`, x.User))
	if x.Password != "" {
		options = append(options, fmt.Sprintf(`-p%v`, x.Password))
	}
	options = append(options, x.DB)
	return options
}
