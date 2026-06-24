package sshvps

import (
	"strconv"
	"strings"
	"time"

	"aqnod/internal/model"
)

// metricsScript gathers everything in one round-trip (POSIX sh). Each section is
// delimited by a ##TAG line so the parser can split deterministically.
const metricsScript = `
echo "##NPROC"; nproc 2>/dev/null || echo 1
echo "##LOAD"; cat /proc/loadavg 2>/dev/null
echo "##MEM"; free -m 2>/dev/null | awk '/^Mem:/{print $2, $3}'
echo "##DISK"; df -BG / 2>/dev/null | awk 'NR==2{print $2, $3}'
echo "##UP"; uptime -p 2>/dev/null || uptime 2>/dev/null
echo "##PS"; docker ps --format '{{.Names}}|{{.Image}}|{{.Status}}' 2>/dev/null
echo "##STATS"; docker stats --no-stream --format '{{.Name}}|{{.CPUPerc}}|{{.MemUsage}}' 2>/dev/null
`

func parseMetrics(raw string) model.Vps {
	sections := splitSections(raw)

	nproc := atoiDefault(strings.TrimSpace(sections["NPROC"]), 1)
	if nproc < 1 {
		nproc = 1
	}

	load1, loadDetail := parseLoad(sections["LOAD"])
	cpu := clamp01(load1 / float64(nproc))

	ramUsed, ramTotal := parseTwoInts(sections["MEM"])
	ram := ratio(ramUsed, ramTotal)

	diskTotal, diskUsed := parseTwoG(sections["DISK"])
	disk := ratio(diskUsed, diskTotal)

	uptime := strings.TrimSpace(firstLine(sections["UP"]))
	if uptime == "" {
		uptime = "online"
	}

	containers := parseContainers(sections["PS"], sections["STATS"])

	warnings := 0
	if cpu > 0.9 {
		warnings++
	}
	if ram > 0.85 {
		warnings++
	}
	if disk > 0.9 {
		warnings++
	}
	for _, c := range containers {
		if c.Status != "running" {
			warnings++
		}
	}

	return model.Vps{
		Uptime:     uptime,
		Warnings:   warnings,
		CPU:        cpu,
		RAM:        ram,
		Disk:       disk,
		CPUDetail:  "load " + loadDetail,
		RAMDetail:  gbDetail(ramUsed, ramTotal),
		DiskDetail: strconv.Itoa(diskUsed) + " / " + strconv.Itoa(diskTotal) + " GB",
		Containers: containers,
		Logs:       deriveLogs(uptime, containers),
	}
}

func splitSections(raw string) map[string]string {
	out := map[string]string{}
	var cur string
	var b strings.Builder
	flush := func() {
		if cur != "" {
			out[cur] = b.String()
			b.Reset()
		}
	}
	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "##") {
			flush()
			cur = strings.TrimPrefix(line, "##")
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	flush()
	return out
}

func parseLoad(s string) (float64, string) {
	fields := strings.Fields(s)
	if len(fields) < 3 {
		return 0, "0.0 / 0.0 / 0.0"
	}
	l1, _ := strconv.ParseFloat(fields[0], 64)
	return l1, fields[0] + " / " + fields[1] + " / " + fields[2]
}

func parseTwoInts(s string) (used, total int) {
	f := strings.Fields(s)
	if len(f) < 2 {
		return 0, 0
	}
	total = atoiDefault(f[0], 0)
	used = atoiDefault(f[1], 0)
	return used, total
}

func parseTwoG(s string) (total, used int) {
	f := strings.Fields(s)
	if len(f) < 2 {
		return 0, 0
	}
	total = atoiDefault(strings.TrimRight(f[0], "G"), 0)
	used = atoiDefault(strings.TrimRight(f[1], "G"), 0)
	return total, used
}

func parseContainers(ps, stats string) []model.Container {
	statMap := map[string][2]string{} // name -> {cpu, mem}
	for _, line := range strings.Split(strings.TrimSpace(stats), "\n") {
		parts := strings.Split(line, "|")
		if len(parts) == 3 {
			statMap[parts[0]] = [2]string{parts[1], parts[2]}
		}
	}
	var out []model.Container
	for _, line := range strings.Split(strings.TrimSpace(ps), "\n") {
		parts := strings.Split(line, "|")
		if len(parts) < 3 || parts[0] == "" {
			continue
		}
		c := model.Container{Name: parts[0], Image: parts[1], Status: dockerStatus(parts[2]), CPU: "—", Mem: "—"}
		if s, ok := statMap[parts[0]]; ok {
			c.CPU, c.Mem = s[0], s[1]
		}
		out = append(out, c)
	}
	return out
}

func dockerStatus(s string) string {
	switch {
	case strings.HasPrefix(s, "Up"):
		return "running"
	case strings.Contains(s, "Restarting"):
		return "restarting"
	default:
		return "down"
	}
}

func deriveLogs(uptime string, containers []model.Container) []model.LogLine {
	now := time.Now().Format("15:04:05")
	logs := []model.LogLine{{Time: now, Level: "INFO", Body: "host · " + uptime}}
	for _, c := range containers {
		level, body := "OK", c.Name+" · healthy"
		if c.Status == "restarting" {
			level, body = "WARN", c.Name+" · restarting"
		} else if c.Status == "down" {
			level, body = "WARN", c.Name+" · down"
		}
		logs = append(logs, model.LogLine{Time: now, Level: level, Body: body})
	}
	if len(containers) == 0 {
		logs = append(logs, model.LogLine{Time: now, Level: "INFO", Body: "docker · nenhum container ou indisponível"})
	}
	return logs
}

// ---- numeric helpers ----

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

func ratio(used, total int) float64 {
	if total <= 0 {
		return 0
	}
	return clamp01(float64(used) / float64(total))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func gbDetail(usedMB, totalMB int) string {
	return strconv.FormatFloat(float64(usedMB)/1024, 'f', 1, 64) + " / " +
		strconv.FormatFloat(float64(totalMB)/1024, 'f', 1, 64) + " GB"
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
