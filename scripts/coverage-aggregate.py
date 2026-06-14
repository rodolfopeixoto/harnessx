#!/usr/bin/env python3
"""Aggregate `go tool cover -func` output by package and fail when any
core package falls below the threshold passed as argv[1] (regex) +
argv[2] (minimum percentage).
"""
import os
import re
import sys


def main() -> int:
    core_re = re.compile(sys.argv[1])
    core_min = float(sys.argv[2])
    totals: dict[str, list[float | int]] = {}
    for line in sys.stdin:
        parts = line.split()
        if len(parts) < 3:
            continue
        fileref = parts[0]
        pct = parts[-1].rstrip("%")
        try:
            pct_val = float(pct)
        except ValueError:
            continue
        pkg = os.path.dirname(fileref)
        bucket = totals.setdefault(pkg, [0.0, 0])
        bucket[0] += pct_val
        bucket[1] += 1

    failed = 0
    for pkg in sorted(totals):
        total, count = totals[pkg]
        avg = total / count if count else 0
        if core_re.search(pkg) and avg < core_min:
            print(f"✗ core {pkg:<70} {avg:5.1f}% (< {core_min}%)")
            failed = 1
    return failed


if __name__ == "__main__":
    sys.exit(main())
