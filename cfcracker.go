package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func randString(n int) string {
	const CHA = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, n)
	for i := range b {
		b[i] = CHA[rand.Intn(len(CHA))]
	}
	return string(b)
}

var ftaa string = randString(18)
var bfaa string = "f1b3f18c715565b589b7823cda7448ce"

func login(client *http.Client, handleOrEmail string, password string) error {
	const host = "https://codeforces.com"
	resp, err := client.Get(host + "/enter")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	csrf, err := findCSRF(client, host+"/enter")
	if err != nil {
		return err
	}

	resp, err = client.PostForm(host+"/enter", url.Values{
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
	fmt.Printf("Logged in as %v", handle)
	return nil
}

func findCSRF(client *http.Client, submitUrl string) (string, error) {
	resp, err := client.Get(submitUrl)
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
		return csrfToken, errors.New("failed to find csrf token")
	}
	return csrfToken, nil
}

type TestCases struct {
	cases    [][]int
	complete bool
}

type Parts struct {
	part1 []byte
	part2 []byte
}

func trimUntilNewline(s []byte) []byte {

	newlineIdx := bytes.IndexByte(s, '\n')
	if newlineIdx == -1 {
		return s
	}
	return s[newlineIdx+1:]
}

const crackerSignature = " void crack(const std::vector<int> &test_case);"

func intoParts(source []byte) (Parts, error) {
	magic := []byte("/* CFCRACKER */")
	signature := []byte(crackerSignature)

	index := bytes.Index(source, magic)
	if index == -1 {
		return Parts{}, errors.New("CFCRACKER magic not found")
	}

	after := source[index+len(magic):]
	if !bytes.HasPrefix(after, signature) {
		return Parts{}, errors.New("crack signature incorrect")
	}

	const headers = `#include <vector>
#include <chrono>
#include <thread>
#include <cassert>
`
	part1 := append([]byte(headers), source[:index]...)
	part2 := source[index+len(magic)+len(signature):]

	return Parts{
			part1,
			part2,
		},
		nil
}

// Possible outcomes that can be forced and their meanings:
// Incorrect answer = last test reached
// Runtime error = stop timer to read inputs
// Program frozen (sleep) = input can't be processed by cfcracker
// TL exceeded

// TODO: create a global start time and wait until ~100ms have passed
// before waiting out the input value to achieve more precise readings

// Idea: set the start time to +400ms or something to ignore longer runs before this test
// allocate readings from 400ms to 950 to numbers from 0 to 10 (i. e. 50ms increments)

func constructFromParts(parts Parts, testCases TestCases) string {
	var builder strings.Builder
	builder.Write(parts.part1)

	// test_cases vector
	builder.Write([]byte("std::vector<std::vector<int>> cfcracker_test_cases {"))
	for _, testCase := range testCases.cases {
		builder.WriteByte('{')
		for _, x := range testCase {
			builder.Write(strconv.AppendInt([]byte{}, int64(x), 10))
			builder.WriteByte(',')
		}
		builder.WriteByte('}')
		builder.WriteByte(',')
	}
	builder.Write([]byte("};\n"))

	// crack function
	builder.Write([]byte(
		`void crack(const std::vector<int> &test_case) {
	for (auto &cfcracker_test_case : cfcracker_test_cases) {
		if (cfcracker_test_case == test_case) {
			// already processed, continue with this test
			return;
		}
	}
`))

	if testCases.complete {
		builder.Write([]byte("cfcracker_test_cases.push_back(std::vector<int>());\n"))
	}

	builder.Write([]byte(
		`	if (test_case.size() == cfcracker_test_cases.back().size()) {
		assert(false && "end of test case");
	}

	std::chrono::time_point cfcracker_tp = std::chrono::system_clock::now();

	int cfcracker_x = test_case[cfcracker_test_cases.size()];
	if (cfcracker_x <= 0) {
		std::this_thread::sleep_for(std::chrono::milliseconds(1000)); // error
	}

	cfcracker_tp += std::chrono::milliseconds(100 * test_case[cfcracker_test_cases.size()]);

	while (std::chrono::system_clock::now() < cfcracker_tp) {
		// waiting
	}
	assert(false && "hello from cfcracker");
`))

	builder.Write([]byte("}"))

	builder.Write(parts.part2)

	return builder.String()
}

func crack(client *http.Client, source []byte) error {

	//var testCases []TestCase

	const submitUrl = "https://codeforces.com/contest/4/submit"
	csrf, err := findCSRF(client, submitUrl)
	if err != nil {
		return err
	}

	//<option value="89" selected="selected">GNU G++20 13.2 (64 bit, winlibs)</option>
	const GNUCXX20_ID = "89"
	const contestId = "4"
	const problemId = "A"

	resp, err := client.PostForm(fmt.Sprintf("%v?csrf_token=%v", submitUrl, csrf), url.Values{
		"csrf_token":            {csrf},
		"ftaa":                  {ftaa},
		"bfaa":                  {bfaa},
		"action":                {"submitSolutionFormSubmitted"},
		"submittedProblemIndex": {problemId},
		"programTypeId":         {GNUCXX20_ID},
		"contestId":             {contestId},
		"source":                {string(source)},
		"tabSize":               {"4"},
		"_tta":                  {"594"},
		"sourceCodeConfirmed":   {"true"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("response status code %d", resp.StatusCode))
	}

	/*
		//TODO: check errors ?
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
	*/
	return nil
}

func main() {
	const usage = "cfcracker <handle or email> <password>"
	if len(os.Args) < 3 {
		log.Fatalln(usage)
	}

	handleOrEmail := os.Args[1]
	password := os.Args[2]

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{Jar: jar}

	err = login(client, handleOrEmail, password)
	if err != nil {
		log.Fatalln(err)
	}

	source, err := os.ReadFile("solutions/watermelon.cpp")
	if err != nil {
		log.Fatalln(err)
	}

	err = crack(client, source)
	if err != nil {
		log.Fatalln(err)
	}
}
