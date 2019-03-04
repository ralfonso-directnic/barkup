package barkup

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	// PGDumpCmd is the path to the `pg_dump` executable
	PGDumpCmd = "pg_dump"
)

// Postgres is an `Exporter` interface that backs up a Postgres database via the `pg_dump` command
type Postgres struct {
	// DB Host (e.g. 127.0.0.1)
	Host string
	// DB Port (e.g. 5432)
	Port string
	// DB Name
	DB string
	// Connection Username
	Username string
	// Connecion Password
	Password string
	// Extra pg_dump options
	// e.g []string{"--inserts"}
	Options []string
}

// Export produces a `pg_dump` of the specified database, and creates a gzip compressed tarball archive.
func (x Postgres) Export() *ExportResult {
	dumpPath := fmt.Sprintf(`%v_%s.sql`, x.DB, time.Now().Format("2006_01_02_15_04_05"))

	//fmt.Println(result.Path)

	options := append(x.dumpOptions(), fmt.Sprintf(`-f%v`, dumpPath))

	//fmt.Println(options)

	cmd := exec.Command(PGDumpCmd, options...)

	// Adds a password varible to exec environment.
	// Can be used instead of ~/.pgpass
	if x.Password != "" {
		cmd.Env = append([]string{"PGPASSWORD=" + x.Password}, cmd.Args...)
	}

	//fmt.Println(cmd)

	result := &ExportResult{MIME: "application/x-7z-compressed"}
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		result.Error = makeErr(err, string(out))
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

func (x Postgres) dumpOptions() []string {
	options := x.Options

	if x.DB != "" {
		options = append(options, fmt.Sprintf(`-d%v`, x.DB))
	}

	if x.Host != "" {
		options = append(options, fmt.Sprintf(`-h%v`, x.Host))
	}

	if x.Port != "" {
		options = append(options, fmt.Sprintf(`-p%v`, x.Port))
	}

	if x.Username != "" {
		options = append(options, fmt.Sprintf(`-U%v`, x.Username))
	}

	return options
}
