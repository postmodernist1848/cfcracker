#include <vector>
#include <chrono>
#include <thread>
#include <cassert>
#include <iostream>
#include <vector>

// create a set of inputs that have been processed already
// if not in this set, wait n * 100 ms
std::vector<std::vector<int>> cfcracker_test_cases {};
void crack(const std::vector<int> &test_case) {
	for (auto &cfcracker_test_case : cfcracker_test_cases) {
		if (cfcracker_test_case == test_case) {
			// already processed, continue with this test
			return;
		}
	}
	if (test_case.size() == cfcracker_test_cases.back().size()) {
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
}

int main() {

    std::vector<int> test;

    int x;
    std::cin >> x;
    test.push_back(x);

    crack(test);

    if (x % 2 == 0) {
        std::cout << "Yes";
    } else {
        std::cout << "No";
    }
}
