cmake_minimum_required(VERSION 3.16)

project(scdviewer LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Qt6 REQUIRED COMPONENTS Core Widgets Gui)

set(SOURCES
    scdviewer.cpp
    SCDInfoRead.cpp
)

set(HEADERS
    SCDInfoRead.h
)

add_executable(scdviewer ${SOURCES} ${HEADERS})

target_link_libraries(scdviewer PRIVATE Qt6::Core Qt6::Widgets Qt6::Gui)
