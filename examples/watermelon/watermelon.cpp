#include <iostream>

int main() {

    int x;
    std::cin >> x;

    if (x % 2 == 0) {
        std::cout << "Yes";
    } else {
        std::cout << "No";
    }
    return 0;
}
