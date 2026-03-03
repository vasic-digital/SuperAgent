#!/usr/bin/env python3
import sys
import re


def add_nolint(filename):
    with open(filename, "r") as f:
        lines = f.readlines()

    i = 0
    while i < len(lines):
        if "_ := io.ReadAll(resp.Body)" in lines[i]:
            indent = lines[i][: len(lines[i]) - len(lines[i].lstrip())]
            # Check if previous line already has nolint
            if i > 0 and "nolint:errcheck" in lines[i - 1]:
                i += 1
                continue
            # Insert nolint comment before the line
            comment = f"{indent}//nolint:errcheck // error reading response body for error message\n"
            lines.insert(i, comment)
            i += 2  # Skip the line we just inserted and the original line
        else:
            i += 1

    with open(filename, "w") as f:
        f.writelines(lines)


if __name__ == "__main__":
    add_nolint(sys.argv[1])
