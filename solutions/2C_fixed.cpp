#include <cassert>
#include <cstdint>
#include <iostream>
#include <vector>
#include <algorithm>

#define DEBUG(expr) #expr " = " << expr << ' '

namespace dbg {
    class DebugOutputStart {
      private:
        class DebugOutputContinue {
          public:
            template <typename T>
            DebugOutputContinue &operator<<(T arg) {
                std::cerr << arg;
                return *this;
            }
            ~DebugOutputContinue() {
                std::cerr << '\n';
            }
        };
      public:
        template <typename T>
        DebugOutputContinue operator<<(T arg) {
            std::cerr << "DEBUG: " << arg;
            return DebugOutputContinue();
        }
    };

    DebugOutputStart out() {
        return DebugOutputStart();
    }

    template <typename Arg, typename... Args>
    void println(Arg&& arg, Args&&... args)
    {
        out << std::forward<Arg>(arg);
        (..., (std::cerr << ',' << ' ' << std::forward<Args>(args))) << '\n';
    }
}

using i64 = int64_t;

struct Edge {
    int x;
    int y1;
    int y2;
    bool is_beg;

    bool operator<(const Edge &other) {
        return x < other.x;
    }
};

class SegmentTree {
public:

    SegmentTree(std::vector<int> &&ys) : m_ys(std::move(ys)), m_t(new Node[4*(m_ys.size() - 1)]{}) {
    }

    void update(int l, int r, bool is_beg) {
        update_impl(0, 0, m_ys.size() - 1, l, r, is_beg);
    }

    i64 get_all() {
        return m_t[0].sum;
    }

    ~SegmentTree() {
        delete[] m_t;
    }

private:
    struct Node {
        i64 sum;
        int counter;
    };

    void update_impl(int v, int l, int r, int qy_l, int qy_r, bool is_beg) {

        // interval i corresponds to [y_i, y_{i+1}]

        if (m_ys[r] <= qy_l || qy_r <= m_ys[l]) return;

        if (qy_l <= m_ys[l] && m_ys[r] <= qy_r) {
            m_t[v].counter += is_beg ? 1 : -1;
            if (m_t[v].counter > 0) m_t[v].sum = m_ys[r] - m_ys[l];
            else m_t[v].sum = (r - l == 1) ? 0 : m_t[2 * v + 1].sum + m_t[2 * v + 2].sum;
            return;
        }

        assert(r - l > 1 && "elementary interval cannot partially overlap");

        int m = (l + r) / 2;
        update_impl(2 * v + 1, l, m, qy_l, qy_r, is_beg);
        update_impl(2 * v + 2, m, r, qy_l, qy_r, is_beg);

        if (m_t[v].counter > 0) m_t[v].sum = m_ys[r] - m_ys[l];
        else m_t[v].sum = (r - l == 1) ? 0 : m_t[2 * v + 1].sum + m_t[2 * v + 2].sum;
    }

    std::vector<int> m_ys;
    Node *m_t;
};

/* CFCRACKER */ void cfc_crack(const std::vector<int> &test_case);

int main() {
    std::ios::sync_with_stdio(false);
    std::cin.tie(nullptr);

    int n;
    std::cin >> n;

    std::vector<int> test_case(4 * n);
    for (int i = 0; i < 4 * n; ++i) {
        std::cin >> test_case[i];
    }

#ifdef CFCRACKER
    cfc_crack(test_case);
#endif

    if (n == 0) {
        std::cout << 0 << '\n';
        return 0;
    }
    std::vector<Edge> edges;
    std::vector<int> ys;

    for (int i = 0; i < n; ++i) {
        int x1 = test_case[i * 4 + 0];
        int y1 = test_case[i * 4 + 1]; 
        int x2 = test_case[i * 4 + 2];
        int y2 = test_case[i * 4 + 3];

        edges.push_back(Edge { 
            x1, y1, y2, true
        });
        edges.push_back(Edge { 
            x2, y1, y2, false
        });
        ys.push_back(y1);
        ys.push_back(y2);
    }
    std::sort(edges.begin(), edges.end());

    std::sort(ys.begin(), ys.end());
    ys.erase(std::unique(ys.begin(), ys.end()), ys.end());

    if (ys.size() == 1) {
        std::cout << 0 << '\n';
        return 0;
    }

    SegmentTree st(std::move(ys));

    i64 ans = 0;

    st.update(edges[0].y1, edges[0].y2, edges[0].is_beg);

    for (size_t i = 1; i < edges.size(); ++i) {

        ans += i64(edges[i].x - edges[i - 1].x) * st.get_all();

        st.update(edges[i].y1, edges[i].y2, edges[i].is_beg);
    }

    std::cout << ans << std::endl;
}
