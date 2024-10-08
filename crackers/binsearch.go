package crackers

import (
	"fmt"
	"github.com/postmodernist1848/cfcracker/client"
	"github.com/postmodernist1848/cfcracker/compilation"
)

type BinSearchCracker struct {
	Low  int // inclusive
	High int // non-inclusive
}

func (binSearchCracker *BinSearchCracker) GetNextValue(c *client.Client, parts compilation.Parts) (int, error) {
	csrf, err := c.FindCSRF(c.SubmitUrl())
	if err != nil {
		return 0, err
	}

	l, r := binSearchCracker.Low, binSearchCracker.High
	// l <= x
	// r > x
	for r-l > 1 {
		m := (l + r) / 2
		res, err := c.Submit(csrf, parts.LTSource(c.Cases, m))
		if err != nil {
			return 0, err
		}
		if res.Verdict == client.MemoryLimitExceeded {
			return 0, client.ValueError{}
		}
		if res.Verdict == client.IdlenessLimitExceeded {
			return 0, client.TestEndError{}
		}

		if res.Verdict == client.RuntimeError {
			// x < m
			r = m
		} else if res.Verdict == client.WrongAnswer {
			// m <= x
			l = m
		} else {
			return 0, fmt.Errorf("binsearch: unused verdict %v", res.Verdict)
		}
	}
	return l, nil
}
