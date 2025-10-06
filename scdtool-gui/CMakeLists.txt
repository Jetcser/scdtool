cmake_minimum_required(VERSION 3.16)  # Qt6 推荐至少 3.16+

# 项目名称
project(SogouSCDTool)

# 设置 C++ 标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

# 查找 Qt6 库
find_package(Qt6 COMPONENTS Widgets REQUIRED)

# 添加可执行文件
add_executable(scdtool
        main.cpp
        FileHandler.cpp
        SCDTool_GUI.cpp
)

# 链接库文件
target_link_libraries(SogouSCDTool
        Qt6::Widgets
)