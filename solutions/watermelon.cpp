#include <iostream>
#include <vector>

// create a set of inputs that have been processed already
// if not in this set, wait n * 100 ms
/* CFCRACKER */ void crack(const std::vector<int> &test_case);

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
