package main

import (
	"fmt"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/colorgful"
)

var (
	defaultSocketPath = fmt.Sprintf("/var/run/user/%d/monk.sock", os.Getuid())
)

var (
	version = "[manual build]"
	usage   = "monk " + version + os.ExpandEnv(`

blah

Usage:
  monk [options] [--interface <prefix>]... daemon
  monk [options] peers [-f]
  monk [options] trust <machine>
  monk [options] stream [<machine>]
  monk [options] fingerprint
  monk -h | --help
  monk --version

Options:
  -s --socket <path>            Listen and serve specified unix socket. Used for
                                 internal purposes only.
                                 [default: `+defaultSocketPath+`]
  -f --fingerprint              Show peers' fingerprints.
  -p --port <port>              Specify port [default: 12345].
  -i --interface <prefix>       Specify network interface to use.
  --data-dir <path>             Directory with sensitive data.
                                 [default: $HOME/.config/monk/]
  --stream-buffer-size <bytes>  Max buffer size for streaming.
                                 [default: 536870912]
  -h --help                     Show this screen.
  --version                     Show version.
`)
)

var (
	logger = lorg.NewLog()
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	logger.SetFormat(
		colorgful.MustApplyDefaultTheme(
			"${time} ${level:%s:left} ${prefix}%s",
			colorgful.Default,
		),
	)

	logger.SetIndentLines(true)

	logger.SetLevel(lorg.LevelDebug)

	if args["daemon"].(bool) {
		handleDaemon(args)
	} else {
		handleClient(args)
	}
}
