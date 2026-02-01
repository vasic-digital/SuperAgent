#!/usr/bin/env python3
import os
import re
import sys

PATTERN = re.compile(r"^\s*defer func\(\) \{ _ = (\w+)\.Close\(\) \}\(\)\s*$")


def process_file(filepath):
    with open(filepath, "r") as f:
        lines = f.readlines()

    modified = False
    new_lines = []
    for line in lines:
        match = PATTERN.match(line)
        if match:
            var = match.group(1)
            new_line = line.replace(
                "defer func() { _ = " + var + ".Close() }()",
                "defer " + var + ".Close()",
            )
            new_lines.append(new_line)
            modified = True
            print(f"{filepath}: {line.strip()} -> {new_line.strip()}")
        else:
            new_lines.append(line)

    if modified:
        with open(filepath, "w") as f:
            f.writelines(new_lines)
    return modified


def main():
    if len(sys.argv) > 1:
        root = sys.argv[1]
    else:
        root = "."

    count = 0
    for dirpath, dirnames, filenames in os.walk(root):
        # exclude .git, vendor, etc.
        if ".git" in dirpath or "vendor" in dirpath:
            continue
        for filename in filenames:
            if filename.endswith("_test.go"):
                filepath = os.path.join(dirpath, filename)
                if process_file(filepath):
                    count += 1
    print(f"Modified {count} files")


if __name__ == "__main__":
    main()
