package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/postmodernist1848/cfcracker/client"
	"github.com/postmodernist1848/cfcracker/crackers"
	"io"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

// Possible outcomes that can be forced manually:
// Wrong answer            (exit(0) with garbage in stdout)
// Runtime error           (non-zero exit code)
// Idleness limit exceeded (sleep)
// Memory limit exceeded   (malloc in a loop)
// Time limit exceeded probably shouldn't be used for anything

// Timer cracker
// Idea: create a global start time and wait until some fixed time has passed since program startup
// before waiting out the input value to negate differences in input reading time (before cfc_crack call)
// Objection: read time is really not that different from test to test and can be accounted for in startup time
// So it's better not to waste precious milliseconds

// Idea: query length of input first to free up ILE verdict?

// Idea: start waiting at 400ms since startup or something to ignore longer runs before current test
// allocate readings from 400ms to 900 to numbers from 0 to 5
// (requires using WRONG_ANSWER and RUNTIME_ERROR for different halves in Timer cracker)

// Idea: use memory to communicate the value
// Memory might be for passed tests only, though
// Can't figure out how the memory reporting works right now
// (really weird values show up in reports)
// Currently only using MEMORY_LIMIT_EXCEEDED to signal error

func fatalln(a ...interface{}) {
	fmt.Fprintf(os.Stderr, "%v: ", os.Args[0])
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func checkNonEmpty(c *client.Client) error {
	if c.ContestUrl == "" {
		return errors.New("no contest URL")
	}
	if c.LangId == "" {
		return errors.New("no language ID")
	}
	if c.ContestId == "" {
		return errors.New("no contest ID")
	}
	if c.ProblemId == "" {
		return errors.New("no problem ID")
	}
	return nil
}

func createConfig(path string) error {
	// TODO: createConfig
	emptyClientBytes, _ := json.MarshalIndent(client.Client{Cases: [][]int{{}}}, "", "    ")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists", path)
		}
		return err
	}
	defer file.Close()

	_, err = file.Write(emptyClientBytes)
	if err != nil {
		return fmt.Errorf("could not write to file %s: %w", path, err)
	}

	fmt.Println("Created sample config at", path)
	return nil
}

func clientFromJSON(path string) (*client.Client, error) {

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("config file not found. Create one using -create-config option")
		}
		return nil, err
	}
	contents, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return nil, err
	}

	var c client.Client
	err = json.Unmarshal(contents, &c)
	if err != nil {
		return nil, err
	}

	err = checkNonEmpty(&c)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(c.ContestUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse contest_url: %w", err)
	}
	c.HostUrl = parsedURL.Scheme + "://" + parsedURL.Host

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c.Jar = jar
	return &c, nil
}

func main() {

	// Options:
	// -source
	// -config
	// -create-config
	// -strategy ?

	// Required:
	sourcePath := flag.String("source", "", "`path` to the problem solution")

	// With default:
	configPath := flag.String("config", "cfcracker.json", "`path` to the config json")

	// Optional:
	createConfigPath := flag.String("create-config", "", "create a new config file at `PATH`")

	flag.Parse()

	shouldCreateConfig := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "create-config" {
			shouldCreateConfig = true
		}
	})

	var err error
	if shouldCreateConfig {
		err = createConfig(*createConfigPath)
		if err != nil {
			fatalln("could not create config file:", err)
		}
		return
	}

	if *configPath == "" {
		fatalln("no config path provided")
	}

	if *sourcePath == "" {
		fatalln("source path is required")
	}

	if len(flag.Args()) < 2 {
		fatalln("expected login and password\n")
	}

	handleOrEmail := flag.Args()[0]
	password := flag.Args()[1]

	c, err := clientFromJSON(*configPath)
	if err != nil {
		fatalln(err)
	}

	debugCLI(handleOrEmail, password, c, sourcePath, createConfigPath)

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatalln(err)
	}
	err = c.Login(handleOrEmail, password)
	if err != nil {
		fatalln(err)
	}

	//cracker := &crackers.BinSearchCracker{
	//	Low:  1,
	//	High: 100 + 1,
	//}

	cracker := &crackers.TimerCracker{
		Increment: 100 * time.Millisecond,
	}

	err = c.Crack(source, cracker)
	if err != nil {
		fatalln(err)
	}
}
