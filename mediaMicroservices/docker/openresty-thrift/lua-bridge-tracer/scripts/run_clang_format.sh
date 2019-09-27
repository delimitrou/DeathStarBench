#!/bin/sh
find . -path ./3rd_party -prune -o \( -name '*.h' -or -name '*.cpp' \) \
  -exec clang-format -i {} \;
