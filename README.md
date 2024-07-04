# Codeforces Tests Cracker

### CF Tests Cracker is a tool that brute forces every test in a Codeforces problem up until (and including) the one that your solution cannot pass.

### Disclaimer: this tool is not intended (and cannot be used) for cheating or hacking a contest.
The reason why I made it was that I couldn't figure out one single test case in a problem that bothered me a lot.

## Installation
```console
$ go install github.com/postmodernist1848/cfcracker/cmd/cfcracker
$ $GOPATH/bin/cfcracker -help
```
## How to use it

Let's say you have your solution for https://codeforces.com/problemset/problem/1/A at [solutions/1A.cpp](solutions/1A.cpp).
(In this simplified example the tests are actually available on CodeForces,
and you get sanitizer output, but it's just a demo).

First, you need to create a config:
```console
$ cfcracker -create-config solutions/1A.json
Created sample config at 1A.json
```
###### TODO: automatic config creation
Fill it with:
```json
{
    "contest_url": "https://codeforces.com/problemset",
    "lang_id": "54",
    "contest_id": "1",
    "problem_id": "A",
    "test_cases": [
        [6, 6, 4],
        []
    ]
}
```

You can provide known test cases (e.g. examples) in test_cases or just leave it empty.

###### **NOTE: empty [] at the end means that previous test ended and should not be cracked.**

Then, modify your source code by adding a call to `cfc_crack` function that will work on the test case, which is to be provided as `std::vector<int>`.
Be sure to specify the signature exactly (with same whitespace, too):
```c++
/* CFCRACKER */ void cfc_crack(const std::vector<int> &test_case);
```

The declaration with a preceding `/* CFCRACKER */` comment will be checked for correctness and replaced with a definition.
Then, call the function on a vector of inputs that you consider part of the test case (for example, you can omit size of input and only add important parts).
You can use the `CFCRACKER` macro that is added by CFCracker in generated code to still be able to send submissions as normal.

[solutions/1A_crack.cpp](solutions/1A_crack.cpp):
```c++
#include <vector>
/* CFCRACKER */ void cfc_crack(const std::vector<int> &test_case);

#include <iostream>
int main() {
    int n, m, a;
    std::cin >> n >> m >> a;
    std::vector<int> test_case {n, m, a};

#ifdef CFCRACKER
    cfc_crack(test_case);
#endif

    int h = (n + a - 1) / a;
    int v = (m + a - 1) / a;
    std::cout << h * v << '\n';
}
```

Start cracking with:
###### Tip: you can use HISTCONTROL=ignorespace option in bash to disable saving the command (and password) to history
```console
$  cfcracker -source solutions/1A_crack.cpp -config solutions/1A.json postmodernist1848 password
```
You'll see the logs of cracking and the test cases will be printed:
```console
Logged in as postmodernist1848
2024/07/04 11:33:25 268736575: Test 2: RUNTIME_ERROR (46ms)
2024/07/04 11:33:25 sign: 1
2024/07/04 11:33:28 268736599: Test 2: WRONG_ANSWER (124ms)
2024/07/04 11:33:28 number: 1
[[6, 6, 4] [1]]
2024/07/04 11:33:35 268736613: Test 2: RUNTIME_ERROR (30ms)
2024/07/04 11:33:35 sign: 1
2024/07/04 11:33:40 268736633: Test 2: WRONG_ANSWER (156ms)
2024/07/04 11:33:40 number: 1
[[6, 6, 4] [1, 1]]
2024/07/04 11:33:50 268736646: Test 2: RUNTIME_ERROR (61ms)
2024/07/04 11:33:50 sign: 1
2024/07/04 11:33:57 268736666: Test 2: WRONG_ANSWER (124ms)
2024/07/04 11:33:57 number: 1
[[6, 6, 4], [1, 1, 1]]
2024/07/04 11:34:17 268736679: Test 2: IDLENESS_LIMIT_EXCEEDED (46ms)
2024/07/04 11:34:23 268736726: Test 3: RUNTIME_ERROR (46ms)
2024/07/04 11:34:23 sign: 1
2024/07/04 11:34:26 268736742: Test 3: WRONG_ANSWER (218ms)
2024/07/04 11:34:26 number: 2
[[6, 6, 4], [1, 1, 1], [2]]
```

Eventually, you reach the limit:
```console
2024/07/04 11:35:46 268736950: Test 4: WRONG_ANSWER (124ms)
2024/07/04 11:35:46 number: 1
[[6, 6, 4], [1, 1, 1], [2, 1, 1], [1, 2, 1]]
cfcracker: You can submit no more than 20 times per 5 minutes
```
Just rerun it after some time with the newly acquired test cases.
###### TODO: add auto saving

You may encounter errors, which can be fixed by removing the test case and reading it again or trying to guess the real value (normally, it's the incorrect value +/- 1)

### Common signs of errors in the test cases read:
- Logs say "Error detected in last value. Retrying..." - CFCracker will remove last value and try to query it again.
- CFCracker terminates saying that there was a wrong answer on test n, but the test passes without cracker.

Sooner or later you'll find that the test contained an overflow edge case:
```console
[[6, 6, 4], [1, 1, 1], [2, 1, 1], [1, 2, 1], [2, 2, 1], [2, 1, 2], [1, 1, 3], [2, 3, 4], [1000000000, 1000000000]]
2024/07/04 13:13:46 268751101: Test 9: RUNTIME_ERROR (77ms)
2024/07/04 13:13:46 sign: 1
2024/07/04 13:13:49 268751129: Test 9: WRONG_ANSWER (140ms)
2024/07/04 13:13:49 number: 1
[[6, 6, 4], [1, 1, 1], [2, 1, 1], [1, 2, 1], [2, 2, 1], [2, 1, 2], [1, 1, 3], [2, 3, 4], [1000000000, 1000000000, 1]]
2024/07/04 13:14:13 268751138: Test 9: IDLENESS_LIMIT_EXCEEDED (46ms)
[[6, 6, 4], [1, 1, 1], [2, 1, 1], [1, 2, 1], [2, 2, 1], [2, 1, 2], [1, 1, 3], [2, 3, 4], [1000000000, 1000000000, 1], []]
2024/07/04 13:14:19 268751189: Test 9: WRONG_ANSWER (62ms)
cfcracker: wrong answer on test 9
```
So you finally [fix your solution](solutions/1A_fixed.cpp).

## How it works

There are two methods of "cracking" implemented currently. \
The most efficient one is the Timer Cracker. It uses the execution time report of a submission to extract a numeric value. \
E.g. wait for 100ms * first digit and save the result. \
Currently only 1 second or greater time limits are supported. \
Only numeric values (64 bit signed integers) can be reported. \
First, the test_cases are compiled to `std::vector<std::vector<int>> cfc_test_cases`.  \
Let's refer to `test_case` argument of `cfc_crack` as "this test":
- When this test is found in `cfc_test_cases`, we skip it if it's not the last one so execution continues on normally.
- When this test is found in `cfc_test_cases` and it's the last one, the program sleeps to get IDLENESS_LIMIT_EXCEEDED verdict.
- When this test is not found in `cfc_test_cases`, we report the value by sleeping (if it's the Timer Cracker).
- If all elements of `cfc_test_cases.back()` do not match this test, MEMORY_LIMIT_EXCEEDED verdict is created to enable auto correcting errors.
