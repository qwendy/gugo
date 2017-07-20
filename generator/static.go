package generator

import (
	"os/exec"
	"runtime"
)

type Static struct {
	SourceDir      string
	DestinationDir string
}

func NewStatic(source, des string) *Static {
	return &Static{
		SourceDir:      source,
		DestinationDir: des,
	}
}

func (s *Static) BatchHandle() (err error) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("rmdir.exe", "/s/q", s.DestinationDir)
		err = cmd.Run()
		if err != nil {
			// cmd := exec.Command("rm", "-rf", s.DestinationDir)
			// err = cmd.Run()
			// if err != nil {
			// 	return
			// }
			// cmd = exec.Command("cp", "-rf", s.SourceDir, s.DestinationDir)
			// err = cmd.Run()
			return
		}
		cmd = exec.Command("xcopy.exe", s.SourceDir, s.DestinationDir, "/s/e/y")
		err = cmd.Run()
	default:
		cmd := exec.Command("rm", "-rf", s.DestinationDir)
		err = cmd.Run()
		if err != nil {
			return
		}
		cmd = exec.Command("cp", "-rf", s.SourceDir, s.DestinationDir)
		err = cmd.Run()
		// err = fmt.Errorf("Have not write %v os copy-handler", runtime.GOOS)
	}
	return
}
