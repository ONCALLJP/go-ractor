# Goractor 🚀

A robust SQL scheduling task manager that helps you automate SQL queries and deliver results to various destinations.

[![Go Report Card](https://goreportcard.com/badge/github.com/ONCALLJP/goractor)](https://goreportcard.com/report/github.com/ONCALLJP/goractor)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Features ✨

- 🕒 Schedule SQL queries with flexible timing options
- 📊 Export results in CSV format
- 🔄 Systemd integration for reliable scheduling
- 🎯 Multiple destination support
- 💼 Easy configuration management

## Installation

```bash
go install github.com/ONCALLJP/goractor/cmd/goractor@latest
```

## Quick Start 🚀

1. Create a task:
```bash
goractor task add
```

2. Install the scheduler:
```bash
goractor systemd install task1
```

3. Check status:
```bash
goractor systemd status
```

## Task Configuration Example 📝

```yaml
tasks:
  my_task:
    name: my_task
    schedule: "every_5min"  # or "daily 09:00"
    query:
      name: user_stats
      sql: |
        SELECT 
          date_trunc('day', created_at) as date,
          count(*) as total
        FROM users
        GROUP BY 1
```

## Available Commands 🛠️

- Task Management:
  ```bash
  goractor task add
  goractor task list
  goractor task remove [task-name]
  ```

- Scheduler Management:
  ```bash
  goractor systemd install [task-name]
  goractor systemd restart [task-name]
  goractor systemd disable [task-name]
  ```

- Debugging:
  ```bash
  goractor debug [task-name]
  ```

## Scheduling Options ⏰

- `every_5min`: Run every 5 minutes
- `every_hour`: Run every hour
- `daily HH:MM`: Run at specific time daily
- `weekly Mon,Wed,Fri HH:MM`: Run on specific days
- `monthly DD HH:MM`: Run on specific day of month

## Contributing 🤝

Contributions are welcome! Here are some ways you can contribute:

1. 🐛 Report bugs
2. 💡 Suggest new features
3. 🔧 Submit pull requests

## License 📄

MIT License - see the [LICENSE](LICENSE) file for details