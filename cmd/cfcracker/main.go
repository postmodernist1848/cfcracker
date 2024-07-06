package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/postmodernist1848/cfcracker/client"
	"github.com/postmodernist1848/cfcracker/crackers"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

func createSampleConfig(path string) error {
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

func createConfig(path string, URL string, langId string) error {

	parts := strings.Split(URL, "/problem/")
	if len(parts) != 2 {
		return errors.New("expected '/problem' in the url")
	}
	contestUrl := parts[0]
	parts = strings.Split(parts[1], "/")

	hostUrl, err := hostUrlFromContestUrl(contestUrl)
	if err != nil {
		return fmt.Errorf("could not get host from contest url: %w", err)
	}
	var contestId string
	var problemId string
	var myUrl string

	if len(parts) == 2 {
		// /problemset
		contestId = parts[0]
		problemId = parts[1]
		myUrl = contestUrl + "/status?my=on"
	} else {
		myUrl = contestUrl + "/my"
		// contest
		problemId = parts[0]
		idx := strings.LastIndex(contestUrl, "/")
		if idx == -1 {
			return errors.New("expected / in contest URL")
		}
		contestId = contestUrl[idx+1:]
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	c := &client.Client{
		Client:     http.Client{Jar: jar},
		HostUrl:    hostUrl,
		MyUrl:      myUrl,
		ContestUrl: contestUrl,
		ContestId:  contestId,
		ProblemId:  problemId,
		Cases: [][]int{
			{},
		},
	}

	if len(langId) > 0 {
		c.LangId = langId
	} else {

		fmt.Print("Enter handle or Email: ")
		var handleOrEmail string
		_, err = fmt.Scanf("%s", &handleOrEmail)
		if err != nil {
			return err
		}

		fmt.Print("Enter password: ")
		var password string
		_, err = fmt.Scanf("%s", &password)
		if err != nil {
			return err
		}

		err = c.Login(handleOrEmail, password)
		if err != nil {
			return fmt.Errorf("failed to log in: %w", err)
		}
		resp, err := c.Get(c.SubmitUrl())
		if err != nil {
			return err
		}
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return err
		}
		doc.Find("select[name='programTypeId'] option").Each(func(index int, item *goquery.Selection) {
			fmt.Printf("%s: %s\n", item.AttrOr("value", "null"), item.Text())
		})
		fmt.Print("Enter language id (one of the above): ")
		var id int
		_, err = fmt.Scanf("%d", &id)
		if err != nil {
			return err
		}
		c.LangId = strconv.Itoa(id)
	}

	clientBytes, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists", path)
		}
		return err
	}
	defer file.Close()
	_, err = file.Write(clientBytes)
	if err != nil {
		return fmt.Errorf("could not write to file %s: %w", path, err)
	}

	fmt.Println("Created config at", path)
	//fmt.Println("host_url", c.HostUrl)
	fmt.Println("my_url:     ", c.MyUrl)
	fmt.Println("contest_url:", c.ContestUrl)
	fmt.Println("contest_id: ", c.ContestId)
	fmt.Println("problem_id: ", c.ProblemId)
	fmt.Println("lang_id:    ", c.LangId)
	fmt.Print("test_cases:   ")
	c.PrintCases()
	return nil
}

func hostUrlFromContestUrl(contestUrl string) (string, error) {
	parsedURL, err := url.Parse(contestUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse contest_url: %w", err)
	}
	return parsedURL.Scheme + "://" + parsedURL.Host, nil
}

func clientFromConfig(path string) (*client.Client, error) {

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

	c.HostUrl, err = hostUrlFromContestUrl(c.ContestUrl)
	if err != nil {
		return nil, fmt.Errorf("could not get host from contest url: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c.Jar = jar
	return &c, nil
}

func clientToConfig(c *client.Client, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	JSON, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	_, err = file.Write(JSON)
	return err
}

func createCreateConfigFlags() (flags *flag.FlagSet, URL *string, langId *string, empty *bool) {
	flags = flag.NewFlagSet("create-config", flag.ExitOnError)
	URL = flags.String("url", "", "Codeforces problem `URL`. If not provided, uses stdin")
	langId = flags.String("langid", "", "Language ID. If not provided, uses stdin")
	empty = flags.Bool("empty", false, "Create empty config")
	return
}

func createSubmitFlags(name string) (flags *flag.FlagSet, sourcePath *string, configPath *string) {
	flags = flag.NewFlagSet(name, flag.ExitOnError)
	sourcePath = flags.String("source", "", "`path` to the problem solution")
	configPath = flags.String("config", "cfcracker.json", "`path` to the config json")
	return
}

func help() {
	const usage = `Usage: cfcracker subcommand [OPTIONS]
    subcommand may be one of the following:
        crack [OPTIONS] - start cracking
        submit [OPTIONS] - submit code without modifications
        create-config <path> - create config file`
	fmt.Println(usage)

	flags, _, _ := createSubmitFlags("crack")
	flags.SetOutput(os.Stdout)
	flags.Usage()

	flags, _, _ = createSubmitFlags("submit")
	flags.SetOutput(os.Stdout)
	flags.Usage()

	flags, _, _, _ = createCreateConfigFlags()
	flags.SetOutput(os.Stderr)
	flags.Usage()
}

func firstArg(args []string, errorMsg string) string {
	if len(args) == 0 {
		fatalln(errorMsg)
	}
	return args[0]
}

func main() {

	// Options:
	// -source
	// -config
	// -create-config
	// -strategy ?

	subcommand := firstArg(os.Args[1:], "expected subcommand")

	var err error

	if subcommand == "help" {
		help()
		return
	}

	if subcommand == "create-config" {
		flags, URL, langId, empty := createCreateConfigFlags()
		flags.Parse(os.Args[2:])
		if len(flags.Args()) > 1 {
			fatalln("too many arguments for create-config")
		}
		path := firstArg(flags.Args(), "expected path for create-config")

		if *empty {
			err = createSampleConfig(path)
			if err != nil {
				fatalln(err)
			}
			return
		}

		if *URL == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter Codeforces problem URL: ")
			*URL, _ = reader.ReadString('\n')
			*URL = strings.TrimSpace(*URL)
		}

		if *URL == "" {
			fatalln("no URL provided")
		}

		err = createConfig(path, *URL, *langId)
		if err != nil {
			fatalln("could not create config file:", err)
		}
		return
	}

	if subcommand != "submit" && subcommand != "crack" {
		fatalln("unknown subcommand", subcommand)
	}

	flags, sourcePath, configPath := createSubmitFlags(subcommand)

	flags.Parse(os.Args[2:])

	if *configPath == "" {
		fatalln("no config path")
	}

	if *sourcePath == "" {
		fatalln("no source path")
	}

	if len(flags.Args()) < 2 {
		fatalln("expected login and password")
	}

	handleOrEmail := flags.Args()[0]
	password := flags.Args()[1]

	c, err := clientFromConfig(*configPath)
	if err != nil {
		fatalln("could not parse config:", err)
	}

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatalln("could not read source file:", err)
	}

	debugCLI(subcommand, c, *sourcePath, *configPath, handleOrEmail, password)

	err = c.Login(handleOrEmail, password)
	if err != nil {
		fatalln("could not log in:", err)
	}

	if subcommand == "submit" {
		csrf, err := c.FindCSRF(c.SubmitUrl())
		if err != nil {
			fatalln("could not find csrf token:", err)
		}
		_, err = c.Submit(csrf, string(source))
		if err != nil {
			fatalln("submission failed:", err)
		}
		return
	}

	// TODO: strategy flag
	//cracker := &crackers.BinSearchCracker{
	//	Low:  1,
	//	High: 100 + 1,
	//}

	cracker := &crackers.TimerCracker{
		Increment: 100 * time.Millisecond,
	}

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	go func() {
		<-sigChannel
		err := clientToConfig(c, *configPath)
		if err != nil {
			fatalln("ERROR: could not save config:", err)
		}
		os.Exit(1)
	}()

	err = c.Crack(source, cracker)
	if err != nil {
		saveErr := clientToConfig(c, *configPath)
		if saveErr != nil {
			fmt.Fprintln(os.Stderr, "ERROR: failed to save config:", saveErr)
		}
		fatalln(err)
	}
}
