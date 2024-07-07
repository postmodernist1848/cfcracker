#include <iostream>
int main() {
    int n, m, a;
    std::cin >> n >> m >> a;

    int h = (n + a - 1) / a;
    int v = (m + a - 1) / a;
    std::cout << h * v << '\n';
}
