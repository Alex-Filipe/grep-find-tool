# grep-tool

CLI tool for searching text inside files and finding files by name.

## Commands

```bash
# Search for literal text
grep-tool grep "TODO" ./src

# Case-insensitive search
grep-tool grep -i "error" ./logs

# Regex pattern
grep-tool grep -e "h.llo" .

# Control parallelism
grep-tool grep -j 8 "pattern"

# Color output
grep-tool grep --color=never "foo" .

# Find files by name (glob)
grep-tool find "*.go"
grep-tool find "test_*" tests/
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0    | Match found |
| 1    | No matches |
| 2    | Error |

## Build

```bash
make build
```

## Test

```bash
make test
```
