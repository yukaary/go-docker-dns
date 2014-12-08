package utils_test

import (
	"github.com/yukaary/go-docker-dns/utils"
	"testing"
)

func TestSplit(t *testing.T) {
	array := utils.SplitAndRemoveSpace("a,b,c", ",")
	if len(array) != 3 {
		t.Errorf("Can't split string.")
	}
}

func TestSplitAndRemoveSpace(t *testing.T) {
	array := utils.SplitAndRemoveSpace("yukaary,  craft", ",")
	if array[1] != "craft" {
		t.Errorf("Can't remove space.")
	}
}
