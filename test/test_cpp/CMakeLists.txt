cmake_minimum_required(VERSION 3.16)
project(antiserial_test_cpp LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(test_cpp main.cpp)
target_include_directories(test_cpp PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

enable_testing()
set(GOLDEN_PAYLOAD "${CMAKE_CURRENT_SOURCE_DIR}/../golden/payload_v2.bin")
add_test(NAME e2e_roundtrip COMMAND test_cpp)
add_test(NAME golden_decode COMMAND test_cpp "${GOLDEN_PAYLOAD}")