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

type TestCase struct {
	numbers  []int
	complete bool
}

type Parts struct {
	part1 []byte
	part2 []byte
	part3 []byte
}

func trimUntilNewline(s []byte) []byte {

	newlineIdx := bytes.IndexByte(s, '\n')
	if newlineIdx == -1 {
		return s
	}
	return s[newlineIdx+1:]
}

func intoParts(source []byte) (Parts, error) {
	setIndex := bytes.Index(source, []byte("// {{CFCRACKER_SET}}"))
	crackIndex := bytes.Index(source, []byte("// {{CFCRACKER_CRACK}}"))
	if setIndex == -1 {
		return Parts{}, errors.New("CFCRACKER_SET not found")
	}
	if crackIndex == -1 {
		return Parts{}, errors.New("CFCRACKER_CRACK not found")
	}

	if crackIndex < setIndex {
		return Parts{}, errors.New("CRACK before SET not allowed")
	}

	part1 := source[:setIndex]

	part2 := source[setIndex+1 : crackIndex]
	part2 = trimUntilNewline(part2)

	part3 := source[crackIndex+1:]
	part3 = trimUntilNewline(part3)
	return Parts{
			part1,
			part2,
			part3,
		},
		nil
}

func fromParts(parts Parts) string {
	var builder strings.Builder

	builder.Write(parts.part1)
	builder.Write(parts.part2)
	builder.Write(parts.part3)

	fmt.Print(builder.String())
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
