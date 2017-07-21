package generator

type Static struct {
	SourceDir      string
	ThemeDir       string
	DestinationDir string
}

func NewStatic(source, theme, des string) *Static {
	return &Static{
		ThemeDir:       theme,
		SourceDir:      source,
		DestinationDir: des,
	}
}

func (s *Static) BatchHandle() (err error) {
	err = s.CopyStaticDir()
	if err != nil {
		return
	}
	err = s.CopyPostImagesDir()
	if err != nil {
		return
	}
	return nil
}
func (s *Static) CopyStaticDir() error {
	return CopyDir(s.ThemeDir+"/static", s.DestinationDir+"/static")
}
func (s *Static) CopyPostImagesDir() error {
	return CopyDir(s.SourceDir+"/_images", s.DestinationDir+"/_images")
}
