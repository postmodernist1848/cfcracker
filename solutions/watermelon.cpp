#include <iostream>

// create a set of inputs that have been processed already
// {{CFCRACKER_SET}}

int main() {

    // if not in this set, wait n * 100 ms
    // {{CFCRACKER_CRACK}}

    int x;
    std::cin >> x;
    if (x % 2 == 0) {
        std::cout << "Yes";
    } else {
        std::cout << "No";
    }
}
