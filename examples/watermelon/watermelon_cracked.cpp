#include <iostream>
#include <vector>

/* CFCRACKER */ void cfc_crack(const std::vector<int> &test_case);

int main() {
    std::vector<int> test;

    int x;
    std::cin >> x;
    test.push_back(x);

#ifdef CFCRACKER
    cfc_crack(test);
#endif

    if (x % 2 == 0) {
        std::cout << "Yes";
    } else {
        std::cout << "No";
    }
    return 0;
}
