package env_wrapper

import (
	"io/ioutil"
	"os"
	"strings"
)

func IsDockerEnv() bool {
	if _, err := os.Stat("/proc/1/cgroup"); !os.IsNotExist(err) {
		if fp, err := os.OpenFile("/proc/1/cgroup", os.O_RDONLY, 0400); err == nil {
			data, err := ioutil.ReadAll(fp)
			if err == nil {
				if strings.Contains(string(data), "docker") {
					return true
				}
			}
		}
	}
	return false
}
