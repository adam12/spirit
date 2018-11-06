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

const envFilename = "./.env"

var (
	processes = make(map[string]*Process)
	env       map[string]string
)

func main() {
	flag.Parse()

	parseProcfile()
	parseEnv()
	setEnv()

	switch flag.Arg(0) {
	case "start":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).start(); err != nil {
				panic(err)
			}
		} else {
			for _, p := range processes {
				if err := p.start(); err != nil {
					panic(err)
				}
			}
		}

	case "stop":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).stop(); err != nil {
				panic(err)
			}
		} else {
			for _, p := range processes {
				if err := p.stop(); err != nil {
					panic(err)
				}
			}
		}

	case "restart":
		name := flag.Arg(1)

		if name != "" {
			if err := lookupProcess(name).restart(); err != nil {
				panic(err)
			}
		} else {
			for _, p := range processes {
				if err := p.restart(); err != nil {
					panic(err)
				}
			}
		}

	case "log":
		name := flag.Arg(1)
		if name == "" {
			quit(usage, 1)
		}

		if err := lookupProcess(name).viewLog(); err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

	case "tail":
		name := flag.Arg(1)
		if name == "" {
			quit(usage, 1)
		}

		if err := lookupProcess(name).tailLog(); err != nil {
			fmt.Print(err)
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
			fmt.Print(err)
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
		os.Setenv(key, value)
	}

	return nil
}

func parseProcfile() {
	if _, err := os.Stat("./Procfile"); os.IsNotExist(err) {
		quit("Unable to find Procfile", 1)
	}

	data, err := ioutil.ReadFile("./Procfile")
	if err != nil {
		panic(err)
	}

	for name, process := range procfile.Parse(string(data)) {
		processes[name] = NewProcess(name, process.Command, process.Arguments)
	}
}

func parseEnv() {
	var err error

	if _, err := os.Stat(envFilename); os.IsNotExist(err) {
		return
	}

	data, err := ioutil.ReadFile(envFilename)
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
