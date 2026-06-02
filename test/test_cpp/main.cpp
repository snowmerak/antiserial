#include "gen/schema_v2.hpp"

#include <cstdint>
#include <cstring>
#include <fstream>
#include <iostream>
#include <iterator>
#include <string>
#include <vector>

static std::vector<uint8_t> read_file(const std::string& path) {
    std::ifstream in(path, std::ios::binary);
    if (!in) {
        throw std::runtime_error("failed to open file: " + path);
    }
    return std::vector<uint8_t>(std::istreambuf_iterator<char>(in), {});
}

static bool bytes_equal(const std::vector<uint8_t>& a, const std::vector<uint8_t>& b) {
    return a.size() == b.size() && std::memcmp(a.data(), b.data(), a.size()) == 0;
}

static bool tags_match(const std::vector<std::string_view>& tags) {
    return tags.size() == 2 && tags[0] == "go" && tags[1] == "rust";
}

static int verify_golden(const std::string& path) {
    try {
        const auto data = read_file(path);
        Payload p{};
        size_t offset = 0;
        p.deserialize(data.data(), data.size(), offset);
        if (offset != data.size()) {
            std::cerr << "did not consume full buffer\n";
            return 1;
        }
        if (p.id != 1234567890 || p.uuid != "abc" || !p.active || !tags_match(p.tags)) {
            std::cerr << "field mismatch\n";
            return 1;
        }
        std::cout << "C++ golden verify: PASSED\n";
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "golden verify error: " << e.what() << "\n";
        return 1;
    }
}

static int e2e_roundtrip() {
    try {
        Payload p{};
        p.id = 1234567890;
        p.uuid = "abc";
        p.active = true;
        p.tags = {std::string_view("go"), std::string_view("rust")};

        std::vector<uint8_t> serialized;
        p.serialize(serialized);

        const std::vector<uint8_t> expected = {
            0x0F,
            0xD2, 0x02, 0x96, 0x49, 0x00, 0x00, 0x00, 0x00,
            0x03, 0x00, 0x61, 0x62, 0x63,
            0x01,
            0x02, 0x00,
            0x02, 0x00, 0x67, 0x6F,
            0x04, 0x00, 0x72, 0x75, 0x73, 0x74,
        };

        if (!bytes_equal(serialized, expected)) {
            std::cerr << "C++ wire format mismatch\n";
            return 1;
        }

        Payload decoded{};
        size_t offset = 0;
        decoded.deserialize(serialized.data(), serialized.size(), offset);
        if (offset != serialized.size()) {
            std::cerr << "C++ deserialize did not consume full buffer\n";
            return 1;
        }
        if (decoded.id != p.id || decoded.uuid != p.uuid || decoded.active != p.active ||
            !tags_match(decoded.tags)) {
            std::cerr << "C++ fields mismatch after roundtrip\n";
            return 1;
        }

        std::cout << "C++ E2E Verification: PASSED\n";
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "e2e error: " << e.what() << "\n";
        return 1;
    }
}

int main(int argc, char* argv[]) {
    if (argc == 2) {
        return verify_golden(argv[1]);
    }
    if (argc != 1) {
        std::cerr << "usage: test_cpp [<golden.bin>]\n";
        return 1;
    }
    return e2e_roundtrip();
}