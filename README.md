# Goractor 🚀

A robust PostgreSQL scheduling task manager that helps you automate SQL queries and deliver results to various destinations.

[![Go Report Card](https://goreportcard.com/badge/github.com/ONCALLJP/goractor)](https://goreportcard.com/report/github.com/ONCALLJP/goractor)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Features ✨

- 🕒 Schedule PostgreSQL queries with flexible timing options
- 📊 Export results in CSV format
- 🔄 Systemd integration for reliable scheduling
- 🎯 Multiple destination support
- 💼 Easy configuration management

## Installation

### Prerequisites
- Go 1.22 or higher
- PostgreSQL
- Systemd (for Linux scheduling)

### Install from source
```bash
git clone https://github.com/ONCALLJP/goractor.git
cd goractor
go build -o goractor cmd/goractor/main.go
```

### Install using go install
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

## ⚠️ Disclaimer

THIS SOFTWARE IS PROVIDED "AS IS" AND ANY EXPRESSED OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

- This tool executes SQL queries automatically. Users are responsible for validating their queries and ensuring they are safe to run.
- Users should thoroughly test their configurations before deploying to production environments.
- The authors are not responsible for any data loss, system damage, or other issues that may occur from using this tool.
- Always backup your data and test thoroughly before using in production.

### Data and Configuration Responsibility
- All configuration files are stored in `~/.goractor/`
  - `tasks.yaml`: Task definitions and schedules
  - `config.yaml`: Database connections
  - `destinations.yaml`: Destination settings (webhooks, tokens)
- **Users are solely responsible for:**
  - Backing up configuration files
  - Securing sensitive information (database credentials, API tokens)
  - Managing and transferring configurations between environments
  - Version controlling their configurations
  - Validating configurations before deployment

### Security Considerations
- Configuration files may contain sensitive information:
  - Database credentials
  - API tokens
  - Webhook URLs
- Users must ensure proper permissions and security measures
- Consider using environment variables or secure vaults for production deployments

### Execution and Data
- This tool executes SQL queries automatically
- Users are responsible for validating their queries and ensuring they are safe to run
- The authors are not responsible for:
  - Data loss or corruption
  - Leaked credentials or tokens
  - Lost or corrupted configuration files
  - System performance issues
  - Any damages resulting from automated execution

### Configuration Best Practices
```bash
# Directory structure
~/.goractor/
├── tasks.yaml      # BACKUP REQUIRED
├── config.yaml     # CONTAINS SENSITIVE DATA
└── destinations.yaml # CONTAINS SENSITIVE DATA



# Goractor 🚀

A robust PostgreSQL scheduling task manager that helps you automate SQL queries and deliver results to various destinations.

## Installation

### Prerequisites
- Go 1.22 or higher
- PostgreSQL
- Systemd (for Linux scheduling)

### Install from source
```bash
git clone https://github.com/ONCALLJP/goractor.git
cd goractor
go build -o goractor cmd/goractor/main.go
```

### Install using go install
```bash
go install github.com/ONCALLJP/goractor/cmd/goractor@latest
```

## Configuration Structure

All configurations are stored in `~/.goractor/`:
```
~/.goractor/
├── config.yaml      # Database connections
├── tasks.yaml       # Task definitions
└── destinations.yaml # Destination settings
```

### Database Configuration
```bash
# Add new database connection
goractor config add
# List configured databases
goractor config list
```

Example config.yaml:
```yaml
databases:
  db1:
    host: localhost
    port: 5432
    user: postgres
    dbname: mydb
```

### Destination Configuration
```bash
# Add new destination
goractor destination add
# List configured destinations
goractor destination list
```

Example destinations.yaml:
```yaml
destinations:
  slack1:
    type: slack
    webhook_url: https://hooks.slack.com/...
    channel: monitoring
```

## Task Management

### Creating a Task
```bash
goractor task add
```

This will prompt for:
1. Task name
2. Database selection
3. SQL query
4. Schedule configuration
5. Destination selection
6. Output format (CSV/JSON)

Example tasks.yaml:
```yaml
tasks:
  daily_stats:
    name: daily_stats
    database: db1
    schedule: "daily 09:00"
    query:
      name: user_stats
      sql: |
        SELECT 
          date_trunc('day', created_at) as date,
          count(*) as total
        FROM users
        GROUP BY 1
    destination: slack1
    output_format: csv
```

### Schedule Types
- Every 5 minutes: `every_5min`
- Every hour: `every_hour`
- Daily at specific time: `daily HH:MM`
- Weekly on specific days: `weekly Mon,Wed,Fri HH:MM`
- Monthly on specific day: `monthly DD HH:MM`

### Managing Tasks
```bash
# List all tasks
goractor task list

# Test a task
goractor task test task1

# Remove a task
goractor task remove task1
```

## Scheduler Management

```bash
# Install task schedule
goractor systemd install task1

# Check status
goractor systemd status

# Restart task
goractor systemd restart task1

# Disable task
goractor systemd disable task1
```

## Debugging

```bash
# View task details and status
goractor debug task1

# View logs
goractor log show

# Clean logs
goractor log clean
```

[Rest of README remains the same...]