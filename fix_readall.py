#!/usr/bin/env python3
import re
import sys


def fix_file(filename):
    with open(filename, "r") as f:
        lines = f.readlines()

    i = 0
    while i < len(lines):
        line = lines[i]
        # Check for body, _ := io.ReadAll(resp.Body) or respBody, _ := io.ReadAll(resp.Body)
        if "_ := io.ReadAll(resp.Body)" in line:
            indent = line[: len(line) - len(line.lstrip())]
            var_name = "body" if "body, _" in line else "respBody"

            # Find the return statement (should be next line or same block)
            # Look ahead up to 3 lines for return statement
            found_return = False
            for j in range(i + 1, min(i + 4, len(lines))):
                if "return" in lines[j] and "fmt.Errorf" in lines[j]:
                    # Extract error message
                    error_line = lines[j]
                    # Replace the line with proper error handling

                    # First, fix the io.ReadAll line
                    lines[i] = f"{indent}{var_name}, err := io.ReadAll(resp.Body)\n"

                    # Insert error check
                    lines.insert(i + 1, f"{indent}if err != nil {{\n")
                    # Need to construct proper return based on original return
                    # Extract return parts
                    match = re.search(
                        r'return (.*?), fmt\.Errorf\("([^"]*)", (.*?)\)', error_line
                    )
                    if match:
                        return_vars = match.group(1)  # e.g., "nil, \"\""
                        error_fmt = match.group(2)  # e.g., "failed to get items: %s"
                        error_args = match.group(3)  # e.g., "string(body)"

                        # Create new error for read failure
                        # Usually error_fmt has %s at end, replace with status code
                        new_error = f'"{error_fmt.replace("%s", "status %d (failed to read body: %v)")}"'
                        lines.insert(
                            i + 2,
                            f"{indent}\treturn {return_vars}, fmt.Errorf({new_error}, resp.StatusCode, err)\n",
                        )
                        lines.insert(i + 3, f"{indent}}}\n")

                        # Keep original return but with variable
                        lines[j + 3] = (
                            f'{indent}return {return_vars}, fmt.Errorf("{error_fmt}", string({var_name}))\n'
                        )
                        i += 4  # Skip inserted lines
                        found_return = True
                        break

            if not found_return:
                # Simple case - just add error check
                lines[i] = f"{indent}{var_name}, err := io.ReadAll(resp.Body)\n"
                lines.insert(i + 1, f"{indent}if err != nil {{\n")
                lines.insert(
                    i + 2,
                    f'{indent}\treturn fmt.Errorf("failed to read response body: %v", err)\n',
                )
                lines.insert(i + 3, f"{indent}}}\n")
                i += 4
        i += 1

    with open(filename, "w") as f:
        f.writelines(lines)


if __name__ == "__main__":
    fix_file(sys.argv[1])
