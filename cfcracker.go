package main

import (
	"cfcracker/client"
	"cfcracker/compilation"
	"cfcracker/crackers"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"
)

// Possible outcomes that can be forced and their meanings:
// Incorrect answer
// Runtime error
// Idleness limit exceeded (sleep)
// TL exceeded = probably shouldn't be used for anything, maybe as sign of fatal error

// Timer cracker
// Idea: create a global start time and wait until ~100ms have passed since program startup
// before waiting out the input value to achieve more precise readings

// Idea:
// measure startup time (time taken to reach cfc_crack() call) and subtract it from the time-to-wait or the actual reading

// Idea: set the start time to +400ms or something to ignore longer runs before this test
// allocate readings from 400ms to 950 to numbers from 0 to 10 (i.e. 50ms increments)

// Idea: use memory to communicate the value
// Memory might be for passed tests only, though
// Can't figure out how the memory reporting works right now

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "cfcracker: expected 2 arguments, but got %d\n ", len(os.Args)-1)
		os.Exit(1)
	}

	handleOrEmail := os.Args[1]
	password := os.Args[2]

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalln(err)
	}
	c := &client.Client{
		Client:     http.Client{Jar: jar},
		HostUrl:    "https://y2023.contest.codeforces.com",
		ContestUrl: "https://y2023.contest.codeforces.com/group/KpnfyE63vT/contest/532234",
		LangId:     "54",     // C++17
		ContestId:  "532234", //
		ProblemId:  "2C",     //
		Cases: compilation.TestCases{
			{0, 0, 2, 2,
				1, 3, 2, 4},
			{1, 1, 3, 5,
				5, 2, 7, 4,
				2, 4, 6, 7},
			{-1000000000, -1000000000, 1000000000, 1000000000},
			{1, 1, 3, 3,
				2, 0, 2, 4},
			{1, 1, 2, 2,
				0, 0, 3, 3},
			{4, 4, 5, 5,
				4, 5, 4, 5},

			// Test 7
			{0, 2, 8, 6,
				1, 1, 7, 7,
				2, 0, 6, 8},
		},
	}

	err = c.Login(handleOrEmail, password)
	if err != nil {
		log.Fatalln(err)
	}

	source, err := os.ReadFile("solutions/C.cpp")
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
}
