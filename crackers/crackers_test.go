package crackers

import (
	"bytes"
	"cfcracker/compilation"
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

var extendedTesting = len(os.Getenv("EXTENDED_TESTING")) > 0

// compileAndRun compiles source and runs it on input as stdin.
// Returns execution time, stdout as string and Cmd.Run() error
func compileAndRun(t *testing.T, source string, input string) (time.Duration, string, error) {
	err := os.WriteFile("../solutions/out.cpp", []byte(source), 0666)
	if err != nil {
		t.Fatalf("could not write to file: %v", err)
	}
	cmd := exec.Command("clang++", "-std=c++17", "-Wall", "-Wextra", "-Wpedantic", "../solutions/out.cpp",
		"-o", "../solutions/a.out")
	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			t.Logf("%s", output)
		}
		t.Fatal(err)
	}
	if len(output) > 0 {
		log.Println(string(output))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, "../solutions/a.out")

	cmd.Stdin = strings.NewReader(input)
	stdout := bytes.NewBuffer(make([]byte, 0))
	cmd.Stdout = stdout

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Now().Sub(start)
	return elapsed, stdout.String(), err
}

func partsFromFile(t *testing.T, path string) compilation.Parts {
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts, err := compilation.NewParts(source)
	if err != nil {
		t.Fatal(err)
	}

	return parts
}

func TestConstructDigitSource(t *testing.T) {

	parts := partsFromFile(t, "../solutions/watermelon.cpp")
	testCases := compilation.TestCases{
		{},
	}
	const increment = 75 * time.Millisecond
	constructed := parts.DigitSource(testCases, increment, 0)

	doTest := func(input int) {
		elapsed, stdout, err := compileAndRun(t, constructed, strconv.Itoa(input))
		if err != nil {
			t.Fatal(err)
		}

		if stdout == "" {
			t.Fatal("stdout empty. Expected wrong answer")
		}

		cracker := &TimerCracker{
			Increment: increment,
		}

		log.Println(elapsed)
		if cracker.elapsedToValue(elapsed) != input {
			t.Fatalf("value: %v", cracker.elapsedToValue(elapsed))
		}
	}

	if extendedTesting {
		for input := 0; input <= 9; input++ {
			doTest(input)
		}
	} else {
		doTest(8)
	}
}

func TestErrorDetection(t *testing.T) {
	parts := partsFromFile(t, "../solutions/C.cpp")

	testCases := compilation.TestCases{
		{0, 0, 2, 2, 1, 3, 2, 4},
		{1, 1, 0},
	}
	input := "3\n1 1 3 5\n5 2 7 4\n2 4 6 7\n"

	const increment = 75 * time.Millisecond
	constructed := parts.DigitSource(testCases, increment, 0)

	_, stdout, err := compileAndRun(t, constructed, input)

	if err == nil {
		t.Fatal("error should have been detected. Stdout: ", stdout)
	}
}

func TestMultiDigit(t *testing.T) {
	parts := partsFromFile(t, "../solutions/C.cpp")

	testCases := compilation.TestCases{
		{0, 0, 2, 2, 1, 3, 2, 4},
		{1, 1, 3, 5, 5, 2, 7, 4, 2, 4, 6, 7},
		{},
	}
	input := "3\n653 1 3 5\n5 2 7 4\n2 4 6 7\n"

	const increment = 75 * time.Millisecond

	cracker := &TimerCracker{
		Increment: increment,
	}

	constructed := parts.DigitSource(testCases, increment, 0)
	elapsed, stdout, err := compileAndRun(t, constructed, input)

	if cracker.elapsedToValue(elapsed) != 3 {
		t.Fatalf("value: %v (3 expected)", cracker.elapsedToValue(elapsed))
	}
	if err == nil {
		t.Fatal("expected RUNTIME_ERROR - first digit. Stdout: ", stdout)
	}

	constructed = parts.DigitSource(testCases, increment, 1)
	elapsed, stdout, err = compileAndRun(t, constructed, input)

	if err == nil {
		t.Fatal("expected RUNTIME_ERROR - second digit. Stdout: ", stdout)
	}
	if cracker.elapsedToValue(elapsed) != 5 {
		t.Fatalf("value: %v (5 expected)", cracker.elapsedToValue(elapsed))
	}

	constructed = parts.DigitSource(testCases, increment, 2)
	elapsed, stdout, err = compileAndRun(t, constructed, input)
	if cracker.elapsedToValue(elapsed) != 6 {
		t.Fatalf("value: %v (6 expected)", cracker.elapsedToValue(elapsed))
	}
	if err != nil {
		t.Fatal("expected WRONG_ANSWER -- last digit", stdout)
	}
}

func TestSignSource(t *testing.T) {
	parts := partsFromFile(t, "../solutions/C.cpp")

	testCases := compilation.TestCases{
		{0, 0, 2, 2, 1, 3, 2, 4},
		{1, 1, 3, 5, 5, 2, 7, 4, 2, 4, 6, 7},
		{},
	}

	input := "3\n653 1 3 5\n5 2 7 4\n2 4 6 7\n"

	constructed := parts.SignSource(testCases)
	_, stdout, err := compileAndRun(t, constructed, input)

	if err == nil {
		t.Fatal("expected RUNTIME_ERROR for positive input. Stdout: ", stdout)
	}

	input = "3\n-53 1 3 5\n5 2 7 4\n2 4 6 7\n"

	constructed = parts.SignSource(testCases)
	_, stdout, err = compileAndRun(t, constructed, input)

	if err != nil {
		t.Fatal("expected WRONG_ANSWER for negative input.")
	}

}

func TestConstructSourceLT(t *testing.T) {

	for i, testCases := range []compilation.TestCases{
		{{8}, {}},
		{{8}, {3}},
	} {
		parts := partsFromFile(t, "../solutions/watermelon.cpp")
		constructed := parts.LTSource(testCases, 10)

		_, stdout, err := compileAndRun(t, constructed, "3")
		if err == nil {
			t.Fatalf("program ended with exit code 0:\n%v", stdout)
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if ws.Signaled() {
					sig := ws.Signal()
					if i == 1 && sig == syscall.SIGKILL {
						continue
					}
				} else {
					if i == 0 && ws.ExitStatus() == 1 {
						continue
					}
				}

			}
		}
		t.Fatalf("%v: incorrect error reason: %v", testCases, exitError)
	}
}

func TestGetValue(t *testing.T) {
	cracker := TimerCracker{
		Increment:   100 * time.Millisecond,
		startupTime: 30 * time.Millisecond,
	}
	v := cracker.elapsedToValue(156 * time.Millisecond)
	if v != 1 {
		t.Fatalf("elapsedToValue: %v", v)
	}
	v = cracker.elapsedToValue(125 * time.Millisecond)
	if v != 1 {
		t.Fatalf("elapsedToValue: %v", v)
	}
}
