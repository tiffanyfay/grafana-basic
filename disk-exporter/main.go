package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":2112", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()
	fmt.Println("Endpoint: http://localhost:2112/metrics")

	// Create non-global registry.
	reg := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors.
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/* node_filesystem_avail_bytes{
	device="/dev/disk1s1",
	device_error="",
	fstype="apfs",
	mountpoint="/System/Volumes/Data"
}
2.3709648896e+11
*/

func recordDiskUsage(reg prometheus.Registerer) {
	usedDisk := promauto.With(reg).NewCounter(prometheus.{
		Name: "node_filesystem_used_bytes",
		Help: "Used disk space",
	})

	go func() {
		for {
			time.Sleep(2 * time.Second)
			getDiskUsage()
			log.Println("Disk usage recorded")
		}
	}()
}

func getDiskUsage() error {

	cmd := exec.Command("df")

	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Print the output
	fmt.Println(string(output)) // Get disk usage
	return nil
}

func bytesToHumanReadable(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
