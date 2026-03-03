#!/usr/bin/env python3
import re
import sys


def fix_io_readall(filename):
    with open(filename, "r") as f:
        lines = f.readlines()

    i = 0
    while i < len(lines):
        line = lines[i]
        # Match body, _ := io.ReadAll(resp.Body) or respBody, _ := io.ReadAll(resp.Body)
        if "_ := io.ReadAll(resp.Body)" in line:
            indent = line[: len(line) - len(line.lstrip())]
            var_name = "body" if "body, _" in line else "respBody"

            # Find the return statement (should be within next few lines)
            for j in range(i + 1, min(i + 5, len(lines))):
                if "return" in lines[j] and "fmt.Errorf" in lines[j]:
                    # Extract the error message format
                    error_line = lines[j]

                    # Simple pattern: return ..., fmt.Errorf("message: %s", string(var_name))
                    # We'll replace with proper error handling

                    # Fix the io.ReadAll line
                    lines[i] = f"{indent}{var_name}, err := io.ReadAll(resp.Body)\n"

                    # Insert error check
                    lines.insert(i + 1, f"{indent}if err != nil {{\n")

                    # Create new error message
                    # Extract error message template
                    match = re.search(r'fmt\.Errorf\("([^"]*)"', error_line)
                    if match:
                        error_fmt = match.group(1)
                        # Replace %s with status code placeholder
                        new_error = error_fmt.replace(
                            "%s", "status %d (failed to read body: %v)"
                        )

                        # Extract return prefix
                        return_match = re.match(
                            r"(\s*return\s+)([^,]+)(,\s*)", error_line
                        )
                        if return_match:
                            return_prefix = return_match.group(1)
                            return_vars = return_match.group(2)
                            return_comma = return_match.group(3)

                            lines.insert(
                                i + 2,
                                f'{indent}\t{return_prefix}{return_vars}{return_comma}fmt.Errorf("{new_error}", resp.StatusCode, err)\n',
                            )
                            lines.insert(i + 3, f"{indent}}}\n")

                            # Update original error line with variable
                            lines[j + 4] = error_line.replace(
                                f"string({var_name})"
                                if f"string({var_name})" in error_line
                                else "string(body)",
                                f"string({var_name})",
                            )
                            i += 4
                            break
                    else:
                        # Fallback
                        lines.insert(
                            i + 2,
                            f'{indent}\treturn fmt.Errorf("failed to read response body: %v", err)\n',
                        )
                        lines.insert(i + 3, f"{indent}}}\n")
                        i += 4
                    break
        i += 1

    with open(filename, "w") as f:
        f.writelines(lines)


if __name__ == "__main__":
    fix_io_readall(sys.argv[1])
