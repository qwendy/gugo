package generator

import "testing"

func TestNewStatic(t *testing.T) {
	s := NewStatic("../themes/微/static", "../public/static")
	if err := s.BatchHandle(); err != nil {
		t.Error(err)
	}
}
