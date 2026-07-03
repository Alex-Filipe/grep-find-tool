# grep-tool

CLI tool for searching text inside files and finding files by name.

## Commands

```bash
# Search for text recursively (default)
grep-tool grep "TODO" ./src

# Case-insensitive search
grep-tool grep -i "error" ./logs

# Control parallelism
grep-tool grep --workers 8 "pattern"

# Find files by name (glob)
grep-tool find "*.go"
```

## Build

```bash
make build
```

## Test

```bash
make test
```
