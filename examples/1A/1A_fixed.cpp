#include <cstdint>
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

    uint64_t h = (n + a - 1) / a;
    uint64_t v = (m + a - 1) / a;
    std::cout << h * v << '\n';
}
