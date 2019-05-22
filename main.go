package main

import (
	"flag"
	"fmt"
	"github.com/chrismytton/procfile"
	"github.com/direnv/go-dotenv"
	_ "golang.org/x/xerrors"
	"io/ioutil"
	"os"
	"os/exec"
)

const usage = `
Usage: spirit COMMAND [opts]

Commands:

	start    [process name]
	stop     [process name]
	restart  [process name]
	log      [process name]
	tail     [process name]
	run      [command]
	status
`

var (
	processes = make(map[string]*Process)
	env       map[string]string
)

func init() {
	flag.Usage = func() {
		fmt.Print(usage)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if err := parseProcfile(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := parseEnv(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := setEnv(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "start":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).start(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			for _, p := range processes {
				if err := p.start(); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

	case "stop":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).stop(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			for _, p := range processes {
				if err := p.stop(); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

	case "restart":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).restart(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			for _, p := range processes {
				if err := p.restart(); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

	case "log":
		name := flag.Arg(1)
		if name == "" {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(1)
		}

		if err := lookupProcess(name).viewLog(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "tail":
		name := flag.Arg(1)
		if name == "" {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(1)
		}

		if err := lookupProcess(name).tailLog(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "run":
		if flag.Arg(1) == "" {
			fmt.Fprintln(os.Stderr, usage)
			os.Exit(1)
		}

		cmd := exec.Command(flag.Arg(1), flag.Args()[2:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "status":
		for name, p := range processes {
			fmt.Printf("%s:\t%s\n", name, p.status())
		}

	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}
}

func setEnv() error {
	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("Error setting environment: %w", err)
		}
	}

	return nil
}

func parseProcfile() error {
	if _, err := os.Stat("Procfile"); os.IsNotExist(err) {
		return fmt.Errorf("Procfile doesn't exist")
	}

	data, err := ioutil.ReadFile("Procfile")
	if err != nil {
		return fmt.Errorf("Unable to read Procfile: %w", err)
	}

	for name, process := range procfile.Parse(string(data)) {
		processes[name] = NewProcess(name, process.Command, process.Arguments)
	}

	return nil
}

func parseEnv() error {
	var err error

	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("Unable to read .env: %w", err)
	}

	env, err = dotenv.Parse(string(data))
	if err != nil {
		return fmt.Errorf("Error parsing .env: %w", err)
	}

	return nil
}

func lookupProcess(name string) *Process {
	if p, ok := processes[name]; ok {
		return p
	}

	fmt.Fprintf(os.Stderr, "Unable to process %s\n", name)
	os.Exit(1)

	// Never reached. Appease the compiler.
	return nil
}
