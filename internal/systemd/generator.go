package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ONCALLJP/goractor/internal/task"
)

type ServiceGenerator struct {
	projectDir string
	serviceDir string
}

func NewServiceGenerator() *ServiceGenerator {
	// Get current working directory as project directory
	wd, err := os.Getwd()
	if err != nil {
		wd = "/home/ubuntu/goractor" // fallback
	}

	return &ServiceGenerator{
		projectDir: wd,
		serviceDir: "/etc/systemd/system",
	}
}

func (g *ServiceGenerator) GenerateService(t *task.Task) error {
	// Generate service file
	serviceContent := g.generateServiceFile(t)
	servicePath := filepath.Join(g.serviceDir, fmt.Sprintf("goractor-%s.service", t.Name))

	// Generate timer file
	timerContent := g.generateTimerFile(t)
	timerPath := filepath.Join(g.serviceDir, fmt.Sprintf("goractor-%s.timer", t.Name))

	// Create temporary files
	tmpServicePath := filepath.Join(os.TempDir(), fmt.Sprintf("goractor-%s.service", t.Name))
	tmpTimerPath := filepath.Join(os.TempDir(), fmt.Sprintf("goractor-%s.timer", t.Name))

	// Write to temporary files first
	if err := os.WriteFile(tmpServicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write temporary service file: %w", err)
	}
	if err := os.WriteFile(tmpTimerPath, []byte(timerContent), 0644); err != nil {
		return fmt.Errorf("failed to write temporary timer file: %w", err)
	}

	// Use sudo to move files and setup service
	commands := [][]string{
		{"mv", tmpServicePath, servicePath},
		{"mv", tmpTimerPath, timerPath},
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", fmt.Sprintf("goractor-%s.timer", t.Name)},
		{"systemctl", "start", fmt.Sprintf("goractor-%s.timer", t.Name)},
		{"touch", "/var/log/goractor.log", "/var/log/goractor.error.log"},
		{"chown", os.Getenv("USER") + ":" + os.Getenv("USER"), "/var/log/goractor.log", "/var/log/goractor.error.log"},
	}

	for _, cmd := range commands {
		command := exec.Command("sudo", cmd...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("failed to execute command %v: %w", cmd, err)
		}
	}

	return nil
}

func (g *ServiceGenerator) generateServiceFile(t *task.Task) string {

	currentUser := os.Getenv("USER")
	homeDir := os.Getenv("HOME")
	// Get absolute path to the binary
	binaryPath := filepath.Join(homeDir, "goractor", "goractor")

	return fmt.Sprintf(`[Unit]
Description=Goractor %s
After=network.target

[Service]
Type=oneshot
WorkingDirectory=%s/goractor
ExecStart=%s task run %s
User=%s
Environment="HOME=%s"
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
StandardOutput=append:/var/log/goractor.log
StandardError=append:/var/log/goractor.error.log

[Install]
WantedBy=multi-user.target
`, t.Name, homeDir, binaryPath, t.Name, currentUser, homeDir)
}

func (g *ServiceGenerator) generateTimerFile(t *task.Task) string {
	timerType, timerValue := convertScheduleToSystemd(t.Schedule, t.Timezone)

	if timerType == "OnUnitActiveSec" {
		return fmt.Sprintf(`[Unit]
Description=Timer for Goractor %s

[Timer]
OnBootSec=0min
OnUnitActiveSec=%s
Persistent=true

[Install]
WantedBy=timers.target
`, t.Name, timerValue)
	}

	return fmt.Sprintf(`[Unit]
Description=Timer for Goractor %s

[Timer]
OnCalendar=%s
Persistent=true

[Install]
WantedBy=timers.target
`, t.Name, timerValue)
}

func convertScheduleToSystemd(schedule, timezone string) (string, string) {
	parts := strings.Fields(schedule)
	scheduleType := parts[0]

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc, err = time.LoadLocation("Local")
		if err != nil {
			loc = time.UTC
		}
	}

	switch scheduleType {
	case "every_5min":
		return "OnUnitActiveSec", "5min"

	case "every_hour":
		return "OnCalendar", "*:00:00"

	case "daily":
		if len(parts) != 2 {
			return "OnUnitActiveSec", "5min"
		}
		timeStr := parts[1]
		scheduledTime, err := time.ParseInLocation("15:04", timeStr, loc)
		if err != nil {
			return "OnUnitActiveSec", "5min"
		}
		return "OnCalendar", fmt.Sprintf("*-*-* %s:00 %s",
			scheduledTime.Format("15:04"), timezone)

	case "weekly":
		if len(parts) != 3 {
			return "OnUnitActiveSec", "5min"
		}
		days := parts[1]
		timeStr := parts[2]

		// Handle special cases
		switch days {
		case "Monday-Friday":
			days = "Mon..Fri"
		case "Saturday,Sunday":
			days = "Sat,Sun"
		}

		scheduledTime, err := time.ParseInLocation("15:04", timeStr, loc)
		if err != nil {
			return "OnUnitActiveSec", "5min"
		}
		return "OnCalendar", fmt.Sprintf("%s *-*-* %s:00 %s",
			days, scheduledTime.Format("15:04"), timezone)

	case "monthly":
		if len(parts) != 3 {
			return "OnUnitActiveSec", "5min"
		}
		day := parts[1]
		timeStr := parts[2]

		scheduledTime, err := time.ParseInLocation("15:04", timeStr, loc)
		if err != nil {
			return "OnUnitActiveSec", "5min"
		}
		return "OnCalendar", fmt.Sprintf("*-*-%s %s:00 %s",
			day, scheduledTime.Format("15:04"), timezone)

	default:
		return "OnUnitActiveSec", "5min"
	}
}

func (g *ServiceGenerator) RemoveService(taskName string) error {
	commands := [][]string{
		{"systemctl", "stop", fmt.Sprintf("goractor-%s.timer", taskName)},
		{"systemctl", "disable", fmt.Sprintf("goractor-%s.timer", taskName)},
		{"rm", filepath.Join(g.serviceDir, fmt.Sprintf("goractor-%s.service", taskName))},
		{"rm", filepath.Join(g.serviceDir, fmt.Sprintf("goractor-%s.timer", taskName))},
		{"systemctl", "daemon-reload"},
	}

	for _, cmd := range commands {
		command := exec.Command("sudo", cmd...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("failed to execute command %v: %w", cmd, err)
		}
	}

	return nil
}
