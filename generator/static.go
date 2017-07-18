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
	case "linux":
		cmd := exec.Command("cp", "-rf", s.SourceDir, s.DestinationDir)
		err = cmd.Run()
	case "windows":
		cmd := exec.Command("xcopy", s.SourceDir, s.DestinationDir, "/s/e/y")
		err = cmd.Run()
	default:
		cmd := exec.Command("cp", "-rf", s.SourceDir, s.DestinationDir)
		err = cmd.Run()
		// err = fmt.Errorf("Have not write %v os copy-handler", runtime.GOOS)
	}
	return
}
