#include <iostream>
#include <vector>

/* CFCRACKER */ void cfc_crack(const std::vector<int> &test_case);

int main() {
    std::vector<int> test;

    int x;
    std::cin >> x;
    test.push_back(x);

    cfc_crack(test);

    if (x % 2 == 0) {
        std::cout << "Yes";
    } else {
        std::cout << "No";
    }
    return 0;
}
