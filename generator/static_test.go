package generator

import "testing"

func TestNewStatic(t *testing.T) {
	s := NewStatic("../source", "../themes/微", "../public")
	if err := s.BatchHandle(); err != nil {
		t.Error(err)
	}
}
