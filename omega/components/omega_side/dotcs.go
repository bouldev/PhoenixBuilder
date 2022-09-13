package omega_side

import (
	"fmt"
	"os"
	"path"
	"phoenixbuilder/omega/utils"
)

func (o *OmegaSide) StartPureDotCSEnv(pythonPath string, portNum int) {
	dotcsDir := path.Join(o.getWorkingDir(), "dotcs")
	if !utils.IsDir(dotcsDir) {
		os.MkdirAll(dotcsDir, 0755)
	}

	o.runCmd("DotCS", fmt.Sprintf("%v %v %v", o.pythonPath, "robot.py", portNum), map[string]string{}, dotcsDir)
}
