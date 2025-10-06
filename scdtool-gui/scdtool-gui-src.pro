TEMPLATE = app
TARGET = scdtool
QT += core gui widgets
CONFIG += c++17

SOURCES += main.cpp \
           FileHandler.cpp \
           SCDTool_GUI.cpp

HEADERS += FileHandler.h \
           SCDTool_GUI.h

FORMS += SCDTool_GUI.ui

# 可选：生成目录
DESTDIR = build
