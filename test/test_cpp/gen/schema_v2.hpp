#pragma once
#include <vector>
#include <string_view>
#include <unordered_map>
#include <span>
#include <cstdint>
#include <cstring>
#include <stdexcept>
#include <optional>

// Forward Declarations
struct Geo;
struct Payload;

struct Geo {
    double lat{};
    double lng{};

    void serialize(std::vector<uint8_t>& buf) const {
        bool f0_present = this->lat != 0.0;
        bool f1_present = this->lng != 0.0;
        uint8_t b0 = 0;
        if (f0_present) {
            b0 |= 1 << 0;
        }
        if (f1_present) {
            b0 |= 1 << 1;
        }
        buf.push_back(b0);
        if (f0_present) {
            {
                double val = this->lat;
                size_t start = buf.size();
                buf.resize(start + sizeof(double));
                std::memcpy(&buf[start], &val, sizeof(double));
            }
        }
        if (f1_present) {
            {
                double val = this->lng;
                size_t start = buf.size();
                buf.resize(start + sizeof(double));
                std::memcpy(&buf[start], &val, sizeof(double));
            }
        }
    }

    void deserialize(const uint8_t* buf, size_t size, size_t& offset) {
        size_t bitmap_start = offset;
        while (true) {
            if (offset >= size) {
                throw std::runtime_error("Unexpected EOF reading bitmap");
            }
            uint8_t b = buf[offset];
            offset++;
            if ((b & 0x80) == 0) {
                break;
            }
        }
        const uint8_t* bitmap_bytes = &buf[bitmap_start];
        size_t bitmap_len = offset - bitmap_start;

        auto is_present = [&](size_t field_idx) -> bool {
            size_t byte_idx = field_idx / 7;
            size_t bit_idx = field_idx % 7;
            if (byte_idx >= bitmap_len) {
                return false;
            }
            return (bitmap_bytes[byte_idx] & (1 << bit_idx)) != 0;
        };

        if (is_present(0)) {
            if (offset + sizeof(double) > size) {
                throw std::runtime_error("Unexpected EOF");
            }
            std::memcpy(&this->lat, &buf[offset], sizeof(double));
            offset += sizeof(double);
        }

        if (is_present(1)) {
            if (offset + sizeof(double) > size) {
                throw std::runtime_error("Unexpected EOF");
            }
            std::memcpy(&this->lng, &buf[offset], sizeof(double));
            offset += sizeof(double);
        }
    }
};

struct Payload {
    int64_t id{};
    std::string_view uuid{};
    bool active{};
    std::vector<std::string_view> tags{};

    void serialize(std::vector<uint8_t>& buf) const {
        bool f0_present = this->id != 0;
        bool f1_present = !this->uuid.empty();
        bool f2_present = this->active;
        bool f3_present = !this->tags.empty();
        uint8_t b0 = 0;
        if (f0_present) {
            b0 |= 1 << 0;
        }
        if (f1_present) {
            b0 |= 1 << 1;
        }
        if (f2_present) {
            b0 |= 1 << 2;
        }
        if (f3_present) {
            b0 |= 1 << 3;
        }
        buf.push_back(b0);
        if (f0_present) {
            {
                int64_t val = this->id;
                size_t start = buf.size();
                buf.resize(start + sizeof(int64_t));
                std::memcpy(&buf[start], &val, sizeof(int64_t));
            }
        }
        if (f1_present) {
            {
                if (this->uuid.size() > 65535) {
                    throw std::runtime_error("string length exceeds uint16 maximum");
                }
                uint16_t length = static_cast<uint16_t>(this->uuid.size());
                size_t start = buf.size();
                buf.resize(start + 2 + length);
                std::memcpy(&buf[start], &length, 2);
                if (length > 0) {
                    std::memcpy(&buf[start + 2], this->uuid.data(), length);
                }
            }
        }
        if (f2_present) {
            {
                uint8_t val = this->active ? 1 : 0;
                buf.push_back(val);
            }
        }
        if (f3_present) {
            {
                if (this->tags.size() > 65535) {
                    throw std::runtime_error("list length exceeds uint16 maximum");
                }
                uint16_t count = static_cast<uint16_t>(this->tags.size());
                size_t start = buf.size();
                buf.resize(start + 2);
                std::memcpy(&buf[start], &count, 2);
                for (const auto& elem0 : this->tags) {
                    {
                        if (elem0.size() > 65535) {
                            throw std::runtime_error("string length exceeds uint16 maximum");
                        }
                        uint16_t length = static_cast<uint16_t>(elem0.size());
                        size_t start = buf.size();
                        buf.resize(start + 2 + length);
                        std::memcpy(&buf[start], &length, 2);
                        if (length > 0) {
                            std::memcpy(&buf[start + 2], elem0.data(), length);
                        }
                    }
                }
            }
        }
    }

    void deserialize(const uint8_t* buf, size_t size, size_t& offset) {
        size_t bitmap_start = offset;
        while (true) {
            if (offset >= size) {
                throw std::runtime_error("Unexpected EOF reading bitmap");
            }
            uint8_t b = buf[offset];
            offset++;
            if ((b & 0x80) == 0) {
                break;
            }
        }
        const uint8_t* bitmap_bytes = &buf[bitmap_start];
        size_t bitmap_len = offset - bitmap_start;

        auto is_present = [&](size_t field_idx) -> bool {
            size_t byte_idx = field_idx / 7;
            size_t bit_idx = field_idx % 7;
            if (byte_idx >= bitmap_len) {
                return false;
            }
            return (bitmap_bytes[byte_idx] & (1 << bit_idx)) != 0;
        };

        if (is_present(0)) {
            if (offset + sizeof(int64_t) > size) {
                throw std::runtime_error("Unexpected EOF");
            }
            std::memcpy(&this->id, &buf[offset], sizeof(int64_t));
            offset += sizeof(int64_t);
        }

        if (is_present(1)) {
            if (offset + 2 > size) {
                throw std::runtime_error("Unexpected EOF");
            }
            {
                uint16_t length;
                std::memcpy(&length, &buf[offset], 2);
                offset += 2;
                if (offset + length > size) {
                    throw std::runtime_error("Unexpected EOF");
                }
                if (length > 0) {
                    this->uuid = std::string_view(reinterpret_cast<const char*>(&buf[offset]), length);
                    offset += length;
                } else {
                    this->uuid = std::string_view();
                }
            }
        }

        if (is_present(2)) {
            if (offset + 1 > size) {
                throw std::runtime_error("Unexpected EOF");
            }
            this->active = buf[offset] != 0;
            offset += 1;
        }

        if (is_present(3)) {
            if (offset + 2 > size) {
                throw std::runtime_error("Unexpected EOF");
            }
            {
                uint16_t count;
                std::memcpy(&count, &buf[offset], 2);
                offset += 2;
                this->tags.resize(count);
                for (size_t i = 0; i < count; i++) {
                    std::string_view elem0{};
                    if (offset + 2 > size) {
                        throw std::runtime_error("Unexpected EOF");
                    }
                    {
                        uint16_t length;
                        std::memcpy(&length, &buf[offset], 2);
                        offset += 2;
                        if (offset + length > size) {
                            throw std::runtime_error("Unexpected EOF");
                        }
                        if (length > 0) {
                            elem0 = std::string_view(reinterpret_cast<const char*>(&buf[offset]), length);
                            offset += length;
                        } else {
                            elem0 = std::string_view();
                        }
                    }
                    this->tags[i] = elem0;
                }
            }
        }
    }
};

