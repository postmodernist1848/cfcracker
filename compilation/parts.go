package compilation

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func RandString(n int) string {
	const CHA = "abcdefghijklmnopqrstuvwxyz0123456789"

	b := make([]byte, n)
	for i := range b {
		b[i] = CHA[rand.Intn(len(CHA))]
	}
	return string(b)
}

type TestCases [][]int

func (testCases TestCases) CompileToVector(builder *strings.Builder) {
	// test_cases vector
	builder.Write([]byte("std::vector<std::vector<int>> cfc_test_cases {"))
	for _, testCase := range testCases {
		builder.WriteByte('{')
		for _, x := range testCase {
			builder.Write(strconv.AppendInt([]byte{}, int64(x), 10))
			builder.WriteByte(',')
		}
		builder.WriteByte('}')
		builder.WriteByte(',')
	}
	builder.Write([]byte("};\n"))
}

type Parts struct {
	part1 []byte
	part2 []byte
}

func New(source []byte) (Parts, error) {
	magic := []byte("/* CFCRACKER */")
	const crackerSignature = " void cfc_crack(const std::vector<int> &test_case);"
	signature := []byte(crackerSignature)

	index := bytes.Index(source, magic)
	if index == -1 {
		return Parts{}, errors.New("CFCRACKER magic not found")
	}

	after := source[index+len(magic):]
	if !bytes.HasPrefix(after, signature) {
		return Parts{}, errors.New("cfc_crack signature incorrect")
	}

	part1 := source[:index]
	part2 := source[index+len(magic)+len(signature):]

	return Parts{
			part1,
			part2,
		},
		nil
}

// skipTests is sourceCode to skip known tests
const skipTests = `	for (size_t i = 0; i < cfc_test_cases.size() - 1; ++i) {
		if (cfc_test_cases[i] == test_case) {
			// already processed, continue with this test
			return;
		}
	}
`

const checkErrors = `
	if (test_case.size() <= cfc_test_cases.back().size()) {
		// error detected, commit suicide
		while (true) {
			char *p = (char *)malloc(1024 * 1024);
			std::cout << (int)p[0];
		}
	}
	for (size_t i = 0; i < cfc_test_cases.back().size(); ++i) {
		if (test_case[i] != cfc_test_cases.back()[i]) {
			while (true) {
				char *p = (char *)malloc(1024 * 1024);
				std::cout << (int)p[0];
			}
		}
	}
`
const reportEndOfTest = `
	if (test_case == cfc_test_cases.back()) {
		std::cerr << "end of test case\n";
		// force idleness_limit_exceeded
		std::this_thread::sleep_for(std::chrono::seconds(4));
	}
`

// LTSource constructs a source for checking if current test value is less than n
// <             RUNTIME_ERROR
// >=            WRONG_ANSWER
// end of test - IDLENESS_LIMIT_EXCEEDED
func (parts Parts) LTSource(cases TestCases, n int) string {
	var builder strings.Builder

	const headers = `#include <vector>
#include <thread>
#include <chrono>
#include <cassert>
`

	builder.Write([]byte(headers))
	builder.Write(parts.part1)
	builder.Write([]byte("void cfc_crack(const std::vector<int> &test_case) {\n"))

	builder.WriteString(fmt.Sprintf("int cfc_n = %v;", n))

	builder.WriteString("//" + RandString(20) + "\n")
	cases.CompileToVector(&builder)
	builder.Write([]byte(skipTests))
	builder.Write([]byte(checkErrors))
	builder.Write([]byte(reportEndOfTest))

	builder.Write([]byte(`	int cfc_x = test_case[cfc_test_cases.back().size()];
	if (cfc_x < cfc_n) {
		std::cerr << "less than\n";
		// force runtime error
		exit(1);
	} else {
		std::cerr << "greater than or equal to\n";
		// force wrong answer
		std::cout << "fejl31413l13joidojdoidaklfjasidfjoi311kjnciaonsof31asdoiuqetpynklcnvzq[]\n";
		exit(0);
	}
}`))
	builder.Write(parts.part2)
	return builder.String()
}

// SignSource constructs a source for getting the sign of current test value.
//
// >= 0 - RUNTIME_ERROR
// < 0  - WRONG_ANSWER
// TODO: also return 'length' of number
func (parts Parts) SignSource(testCases TestCases) string {
	var builder strings.Builder

	const headers = `#include <vector>
#include <thread>
#include <chrono>
#include <cassert>
`

	builder.Write([]byte(headers))
	builder.Write(parts.part1)

	// crack function
	builder.Write([]byte("void cfc_crack(const std::vector<int> &test_case) {\n"))

	builder.WriteString("//" + RandString(20) + "\n")
	testCases.CompileToVector(&builder)
	builder.Write([]byte(skipTests))
	builder.Write([]byte(checkErrors))
	builder.Write([]byte(reportEndOfTest))

	builder.Write([]byte(
		`
		int x = test_case[cfc_test_cases.back().size()];
	if (x >= 0) {
		exit(1);
	}
	std::cout << "fejl31413l13joidojdoidaklfjasidfjoi311kjnciaonsof31asdoiuqetpynklcnvzq[]\n";
	exit(0);
}`))

	builder.Write(parts.part2)
	return builder.String()
}

// DigitSource constructs a source for getting nth digit of current test value
//
// last non-zero digit of number - WRONG_ANSWER
// not last digit - RUNTIME_ERROR
// end of test - IDLENESS_LIMIT_EXCEEDED
// mistake detected = MEMORY_LIMIT_EXCEEDED
func (parts Parts) DigitSource(testCases TestCases, increment time.Duration, digitNo int) string {
	var builder strings.Builder

	const headers = `#include <vector>
#include <thread>
#include <chrono>
#include <cassert>
`

	builder.Write([]byte(headers))
	builder.Write(parts.part1)

	// crack function
	builder.Write([]byte("void cfc_crack(const std::vector<int> &test_case) {\n"))

	builder.WriteString("//" + RandString(20) + "\n")
	testCases.CompileToVector(&builder)
	builder.Write([]byte(skipTests))
	builder.Write([]byte(checkErrors))
	builder.Write([]byte(reportEndOfTest))

	// constants
	// TODO:
	// const cfc_global_start?

	builder.WriteString(fmt.Sprintf("int n = %v;", digitNo))

	builder.WriteString(
		fmt.Sprintf("const int cfc_increment = %v;\n",
			increment.Milliseconds()))

	builder.Write([]byte("	"))

	builder.Write([]byte(
		`
		int x = test_case[cfc_test_cases.back().size()];
		int power = 1;
		for (int i = 0; i < n; ++i) {
			power *= 10;
		}
		int x_digit = (std::abs(x)/power)%10;

	std::clock_t start = std::clock();
    while (1000.0 * (std::clock() - start) / CLOCKS_PER_SEC < x_digit * cfc_increment) { /* wait */ }
	if (std::abs(x) / (power * 10) == 0) {
		std::cout << "fejl31413l13joidojdoidaklfjasidfjoi311kjnciaonsof31asdoiuqetpynklcnvzq[]\n";
		exit(0);
	}
    exit(1);

}`))

	builder.Write(parts.part2)
	return builder.String()
}
