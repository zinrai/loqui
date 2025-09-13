# loqui

An interactive LogQL query builder for Grafana Loki that reduces cognitive load when searching logs.

## Why loqui?

Working with Loki's `logcli` is powerful but comes with significant cognitive overhead:

### The Pain Points

1. **Label Discovery Burden**
   - You need to remember exact label names
   - No autocomplete or hints for available labels
   - Easy to make typos that result in empty results

2. **Complex Time Range Specification**
   - `--from` and `--to` require RFC3339 format (`2025-08-14T09:00:00+09:00`)
   - High cognitive load when specifying exact time periods (e.g., "from 9 AM to 6 PM on August 14th")
   - Mental burden of converting human-readable dates/times to RFC3339 format
   - Different formats for relative (`--since`) vs absolute time ranges

3. **Label Value Guessing**
   - You need to know exact values for each label
   - No way to see what values are actually available
   - Trial and error process to find the right combination

### The Solution

`loqui` (Loki Query Interactive) solves these problems by:

- Using `fzf` for interactive label and value selection
- Converting human-friendly time formats to RFC3339 automatically
- Showing only available options at each step
- Sensible defaults - just press Enter to skip optional features
- Generating the correct `logcli` command or executing it directly with `-exec`

## Installation

```bash
$ go install github.com/yourusername/loqui@latest
```

### Prerequisites

- [logcli](https://grafana.com/docs/loki/latest/query/logcli/)
- [fzf](https://github.com/junegunn/fzf)

## Usage

### Generate Command (Default)

```bash
$ export LOKI_ADDR=http://localhost:3100
$ loqui

Select time range type:
1. Relative (e.g., 1h, 24h)
2. Absolute (specific dates)
Enter choice (1-2): 2

Enter start time (YYYY-MM-DD HH:MM or YYYY-MM-DD): 2025-08-14 09:00
Enter end time (YYYY-MM-DD HH:MM or YYYY-MM-DD): 2025-08-14 18:00

# Interactive fzf selection of labels
Select label: app

Select operator for 'app' (default: 1):
1. = (equals)
2. != (not equals)
3. =~ (regex match)
4. !~ (regex not match)
Enter number (1-4) or press Enter for default: 1

# Interactive fzf selection of values
Select value for 'app': nginx

=== Current labels ===
[SET] app="nginx"

Add more labels? (y/N): y

Select label: env
Select value for 'env': production

=== Current labels ===
[SET] app="nginx"
[SET] env="production"

Add more labels? (y/N): [Enter]

Add line filter? (y/N): y

Select line filter operator (default: 1):
1. |= (contains)
2. != (does not contain)
3. |~ (matches regex)
4. !~ (does not match regex)
Enter number (1-4) or press Enter for default: 1

Enter filter text: error

# Output:
logcli query '{app="nginx",env="production"} |= "error"' --from 2025-08-14T09:00:00+09:00 --to 2025-08-14T18:00:00+09:00
```

### Execute Directly

```bash
# Execute the query immediately instead of generating command
$ loqui -exec

# [Interactive selection same as above...]

# Direct output of logs:
2025-08-14T10:00:00+09:00 nginx: [error] connection refused
2025-08-14T10:00:01+09:00 nginx: [error] timeout occurred
...
```

### Examples

```bash
# Generate command for later use or modification
$ loqui

# Execute in a subshell (traditional way)
$ $(loqui)

# Execute directly with -exec flag
$ loqui -exec

# Pipe the results
$ loqui -exec | grep ERROR

# Use with pager
$ loqui -exec | less

# Use with cache for faster label selection
$ loqui -cache -exec
```

## Time Format Support

Instead of remembering RFC3339 format, use natural formats:

- `YYYY-MM-DD` → Automatically converts to start/end of day
- `YYYY-MM-DD HH:MM` → Converts to full RFC3339 with local timezone
- Relative times like `1h`, `24h`, `7d` for recent logs

## Options

```bash
-help        Show help message
-version     Show version
-cache       Enable label cache (faster but might show stale labels)
-exec        Execute the command immediately
```

## How It Works

1. **Time Range First**: Choose between relative (last N hours) or absolute dates
2. **Interactive Label Selection**: Use `fzf` to search and select from actual labels in your Loki instance (press Enter to skip additional labels)
3. **Smart Value Selection**: For each label, see only the values that actually exist
4. **Operator Support**: Not just equality - supports `!=`, `=~`, and `!~` for advanced queries
5. **Line Filters**: Optional - press Enter to skip
6. **Command Generation or Execution**: Outputs a ready-to-run `logcli` command or executes it directly with `-exec`

## Using Cache for Label and Label value

If you have many labels and values, enable caching for faster selection:

```bash
$ loqui -cache
```

Note: Cache is read from `~/.cache/loqui/labels.json` if it exists. Use a [loki-index-dump](https://github.com/zinrai/loki-index-dump) to manage the cache.

## Notes

For querying specific tenants in multi-tenant Loki environments, refer to the [LogCLI getting started](https://grafana.com/docs/loki/latest/query/logcli/getting-started/)

## License

This project is licensed under the [MIT License](./LICENSE).
