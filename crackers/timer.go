package crackers

import (
	"cfcracker/client"
	"cfcracker/compilation"
	"fmt"
	"log"
	"math"
	"time"
)

// TimerCracker uses execution time to get the values digit by digit
// TODO: this could be made more precise (i. e RUNTIME_ERROR for digits 0-4 and WRONG_ANSWER for 5-9)
// TODO: query length of input first to free up ILE verdict
type TimerCracker struct {
	Increment   time.Duration // result is returned as x * Increment
	startupTime time.Duration // startup time is subtracted from the result
}

func (cracker *TimerCracker) elapsedToValue(elapsed time.Duration) int {
	return int(math.Round(float64(elapsed-cracker.startupTime) / float64(cracker.Increment)))
}

func (cracker *TimerCracker) GetNextValue(c *client.Client, parts compilation.Parts) (int, error) {

	csrf, err := c.FindCSRF(c.SubmitUrl())
	if err != nil {
		return 0, err
	}
	if cracker.startupTime == 0 /*|| len(c.Cases) == 1 && len(c.Cases[0]) == 0*/ {
		// TODO: measure startupTime
		cracker.startupTime = time.Millisecond * 30
	}

	signSource := parts.SignSource(c.Cases)
	sub, err := c.SendSubmission(csrf, signSource)
	if err != nil {
		return 0, err
	}

	var sign int

	if sub.Verdict == client.MemoryLimitExceeded {
		return 0, fmt.Errorf("error detected in previous values of this test")
	}
	if sub.Verdict == client.IdlenessLimitExceeded {
		return 0, client.LastTestValueError{}
	}

	if sub.Verdict == client.RuntimeError {
		sign = 1
	} else if sub.Verdict == client.WrongAnswer {
		sign = -1
	} else {
		return 0, fmt.Errorf("timer: unused verdict %v", sub.Verdict)
	}

	log.Printf("sign: %v", sign)

	result := 0
	digitNo := 0
	power := 1
	for {
		source := parts.DigitSource(
			c.Cases,
			cracker.Increment,
			digitNo,
		)
		sub, err = c.SendSubmission(csrf, source)
		if err != nil {
			return 0, err
		}
		if sub.Verdict == client.MemoryLimitExceeded {
			return 0, fmt.Errorf("mistake detected in previous values of this test")
		}
		if sub.Verdict == client.IdlenessLimitExceeded {
			return 0, client.LastTestValueError{}
		}
		result += power * cracker.elapsedToValue(sub.Time)
		power *= 10
		digitNo++
		log.Printf("number: %v", result*sign)
		if sub.Verdict == client.RuntimeError {
			continue
		}
		if sub.Verdict == client.WrongAnswer {
			break
		}
		return 0, fmt.Errorf("timer: unused verdict %v", sub.Verdict)
	}

	return result * sign, nil
}
