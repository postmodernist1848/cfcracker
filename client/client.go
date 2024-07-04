// Package client provides networking capabilities and the generic Crack method
// that uses Cracker interface.
package client

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/postmodernist1848/cfcracker/compilation"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ftaa = compilation.RandString(18)
var bfaa = "f1b3f18c715565b589b7823cda7448ce"

type Client struct {
	http.Client `json:"-"`
	HostUrl     string                `json:"host_url,omitempty"` // as in https://codeforces.com, derived from contest_url
	MyUrl       string                `json:"my_url"`             // as in https://codeforces.com/contest/4/my
	ContestUrl  string                `json:"contest_url"`        // as in https://codeforces.com/contest/4
	LangId      string                `json:"lang_id"`
	ContestId   string                `json:"contest_id"`
	ProblemId   string                `json:"problem_id"`
	Cases       compilation.TestCases `json:"test_cases"`
}

func (client *Client) SubmitUrl() string {
	return client.ContestUrl + "/submit"
}

func (client *Client) LoginUrl() string {
	return client.HostUrl + "/enter"
}

func (client *Client) FindCSRF(URL string) (string, error) {
	resp, err := client.Get(URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// Find the input tag with name="csrf_token"
	csrfToken, exists := doc.Find("input[name='csrf_token']").Attr("value")
	if !exists {
		return csrfToken, fmt.Errorf("%v: failed to find csrf token", URL)
	}
	return csrfToken, nil
}

func (client *Client) Login(handleOrEmail string, password string) error {

	csrf, err := client.FindCSRF(client.LoginUrl())
	if err != nil {
		return err
	}

	resp, err := client.PostForm(client.LoginUrl(), url.Values{
		"csrf_token":    {csrf},
		"action":        {"enter"},
		"ftaa":          {ftaa},
		"bfaa":          {bfaa},
		"handleOrEmail": {handleOrEmail},
		"password":      {password},
		"_tta":          {"176"},
		"remember":      {"off"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	reg := regexp.MustCompile(`handle = "([\s\S]+?)"`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil || len(tmp) < 2 {
		return errors.New("failed to log in")
	}
	handle := string(tmp[1])
	fmt.Printf("Logged in as %s\n", handle)
	//TODO: return handle maybe
	return nil
}

func findErrorMessage(body []byte) (string, error) {
	reg := regexp.MustCompile(`error[a-zA-Z_\-\ ]*">(.*?)</span>`)
	tmp := reg.FindSubmatch(body)
	if tmp == nil {
		return "", errors.New("cannot find error")
	}
	return string(tmp[1]), nil
}

func findMessage(body []byte) (string, error) {
	reg := regexp.MustCompile(`Codeforces.showMessage\("([^"]*)"\);\s*?Codeforces\.reformatTimes\(\);`)
	tmp := reg.FindSubmatch(body)
	if tmp != nil {
		return string(tmp[1]), nil
	}
	return "", errors.New("could not find any message")
}

func (client *Client) PostSubmission(csrf string, source string) error {
	resp, err := client.PostForm(fmt.Sprintf("%v?csrf_token=%v", client.SubmitUrl(), csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {ftaa},
		"bfaa":                  {bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {client.ProblemId},
		"programTypeId":         {client.LangId},
		"contestId":             {client.ContestId},
		"source":                {source},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("submission response status code %d", resp.StatusCode)
	}

	// Check if sent successfully
	body, err := io.ReadAll(resp.Body)
	errMsg, err := findErrorMessage(body)
	if err == nil {
		return errors.New(errMsg)
	}

	msg, err := findMessage(body)
	if err != nil {
		return errors.New("submit failed")
	}
	if !strings.Contains(msg, "submitted successfully") {
		return errors.New(msg)
	}

	return nil
}

type TestEndError struct{}

func (TestEndError) Error() string {
	return "last test reached"
}

type ValueError struct{}

func (ValueError) Error() string {
	return "last value is incorrect"
}

type Verdict int

const (
	WrongAnswer           = iota
	RuntimeError          = iota
	IdlenessLimitExceeded = iota
	MemoryLimitExceeded
)

func printCases(cases compilation.TestCases) {
	fmt.Print("[")
	for i, c := range cases {
		fmt.Print("[")
		if len(c) > 0 {
			fmt.Print(c[0])
			for _, v := range c[1:] {
				fmt.Print(", ", v)
			}
		}
		fmt.Print("]")
		if i != len(cases)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println("]")
}

type Cracker interface {
	GetNextValue(client *Client, parts compilation.Parts) (int, error)
}

func removeLast(slice *[]int) {
	*slice = (*slice)[:len(*slice)-1]
}

func (client *Client) Crack(source []byte, cracker Cracker) error {

	parts, err := compilation.NewParts(source)
	if err != nil {
		return err
	}

	if len(client.Cases) == 0 {
		client.Cases = append(client.Cases, []int{})
	}

	for {
		printCases(client.Cases)
		next, err := cracker.GetNextValue(client, parts)
		if err != nil {
			if _, ok := err.(ValueError); ok {
				log.Printf("Error detected in last value. Retrying...")
				removeLast(&client.Cases[len(client.Cases)-1])
				continue
			}
			if _, ok := err.(TestEndError); ok {
				client.Cases = append(client.Cases, []int{})
				continue
			}
			return err
		}
		client.Cases[len(client.Cases)-1] = append(client.Cases[len(client.Cases)-1], next)
	}
}

func (client *Client) getVerdict() (idText string, verdictText string, testNo string, timeText string, err error) {
	resp, err := client.Get(client.MyUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	verdictText, exists := doc.Find("span.submissionVerdictWrapper").Attr("submissionverdict")
	if !exists {
		return "", "", "", "", errors.New("submission verdict not found in response")
	}

	idText = doc.Find("td.id-cell").First().Text()
	testNo = doc.Find("span.verdict-format-judged").First().Text()
	timeText = doc.Find("td.time-consumed-cell").First().Text()
	return
}

func (client *Client) getVerdictChecked(prevId string) (idText string, verdictText string, testNo string, resTime time.Duration, err error) {
	var timeText string
	for {
		time.Sleep(time.Second)
		idText, verdictText, testNo, timeText, err = client.getVerdict()
		if err != nil {
			return
		}

		if idText != prevId && !strings.Contains(verdictText, "TESTING") {

			var id2 string
			var verdict2 string
			var timeText2 string
			id2, verdict2, _, timeText2, err = client.getVerdict()
			if err != nil {
				return
			}

			var unit string
			var subTime int64
			_, err = fmt.Sscan(timeText, &subTime, &unit)
			if err != nil {
				return
			}
			if subTime == 0 {
				log.Printf("Encountered 0ms time. Retrying...")
				continue
			}
			if unit != "ms" {
				log.Printf("Warning: parsed %v time unit. Expected ms", unit)
			}
			resTime = time.Millisecond * time.Duration(subTime)

			if idText == id2 && verdictText == verdict2 && timeText == timeText2 {
				break
			}
		}
	}
	return
}

type SubmissionResult struct {
	Verdict
	Time time.Duration
}

func (client *Client) SendSubmission(csrf string, source string) (res SubmissionResult, err error) {
	prevId, _, _, _, err := client.getVerdict()
	if err != nil {
		return
	}
	err = client.PostSubmission(csrf, source)
	if err != nil {
		return
	}

	id, verdict, testNoStr, subTime, err := client.getVerdictChecked(prevId)

	log.Printf("%v: Test %v: %v (%v)\n", strings.TrimSpace(id), testNoStr, verdict, subTime)

	res.Time = subTime

	if strings.Contains(verdict, "IDLENESS_LIMIT_EXCEEDED") {
		res.Verdict = IdlenessLimitExceeded
		return
	}

	if strings.Contains(verdict, "RUNTIME_ERROR") {
		res.Verdict = RuntimeError
		return
	}
	if verdict == "MEMORY_LIMIT_EXCEEDED" {
		res.Verdict = MemoryLimitExceeded
		return
	}
	if strings.Contains(verdict, "WRONG_ANSWER") {

		var testNo int
		testNo, err = strconv.Atoi(testNoStr)
		if err != nil {
			return
		}
		if testNo < len(client.Cases) {
			err = fmt.Errorf("wrong answer on test %v", testNo)
			return
		}

		res.Verdict = WrongAnswer
		return
	}

	err = fmt.Errorf("unknown verdict %v", verdict)
	return
}
