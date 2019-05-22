package main

import (
	"flag"
	"fmt"
	"github.com/chrismytton/procfile"
	"github.com/direnv/go-dotenv"
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

	parseProcfile()
	parseEnv()

	if err := setEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set environment: %v\n", err)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "start":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).start(); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to stop process %s: %s\n", name, err)
				os.Exit(1)
			}
		} else {
			for _, p := range processes {
				if err := p.start(); err != nil {
					fmt.Fprintf(os.Stderr, "Unable to start process %s: %s\n", p.Name, err)
					os.Exit(1)
				}
			}
		}

	case "stop":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).stop(); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to stop process %s: %s\n", name, err)
				os.Exit(1)
			}
		} else {
			for _, p := range processes {
				if err := p.stop(); err != nil {
					fmt.Fprintf(os.Stderr, "Unable to stop process %s: %s\n", p.Name, err)
					os.Exit(1)
				}
			}
		}

	case "restart":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).restart(); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to restart process %s: %s\n", name, err)
				os.Exit(1)
			}
		} else {
			for _, p := range processes {
				if err := p.restart(); err != nil {
					fmt.Fprintf(os.Stderr, "Unable to restart process %s: %s\n", p.Name, err)
					os.Exit(1)
				}
			}
		}

	case "log":
		name := flag.Arg(1)
		if name == "" {
			quit(usage, 1)
		}

		if err := lookupProcess(name).viewLog(); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to view log of %s: %s\n", name, err)
			os.Exit(1)
		}

	case "tail":
		name := flag.Arg(1)
		if name == "" {
			quit(usage, 1)
		}

		if err := lookupProcess(name).tailLog(); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to tail log of %s: %s\n", name, err)
			os.Exit(1)
		}

	case "run":
		if flag.Arg(1) == "" {
			quit(usage, 1)
		}

		cmd := exec.Command(flag.Arg(1), flag.Args()[2:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to run command: %s\n", err)
			os.Exit(1)
		}

	case "status":
		for name, p := range processes {
			fmt.Printf("%s:\t%s\n", name, p.status())
		}

	default:
		quit(usage, 1)
	}
}

func quit(message string, code int) {
	fmt.Println(message)
	os.Exit(code)
}

func setEnv() error {
	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return nil
}

func parseProcfile() {
	if _, err := os.Stat("Procfile"); os.IsNotExist(err) {
		quit("Unable to find Procfile", 1)
	}

	data, err := ioutil.ReadFile("Procfile")
	if err != nil {
		panic(err)
	}

	for name, process := range procfile.Parse(string(data)) {
		processes[name] = NewProcess(name, process.Command, process.Arguments)
	}
}

func parseEnv() {
	var err error

	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return
	}

	data, err := ioutil.ReadFile(".env")
	if err != nil {
		panic(err)
	}

	env, err = dotenv.Parse(string(data))
	if err != nil {
		panic(err)
	}
}

func lookupProcess(name string) *Process {
	if p, ok := processes[name]; ok {
		return p
	}

	quit("Unable to find process "+name, 1)

	// Never reached. Appease the compiler.
	return nil
}
