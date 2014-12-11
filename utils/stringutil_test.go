package utils_test

import (
	"github.com/yukaary/go-docker-dns/utils"
	"strings"
	"testing"
)

func TestJoinSingleString(t *testing.T) {
	key := genKeyFromKeychain("crafter", "yukaary")
	if key != "yukaary" {
		t.Errorf("Can't generate exepected.")
	}
}

func TestJoinStrings(t *testing.T) {
	key := genKeyFromKeychain("craft", "yukari", "yuzuki", "darkmaster")
	if key != "yukari/yuzuki/darkmaster" {
		t.Errorf("Can't generate exepected.")
	}
}

func genKeyFromKeychain(value string, key ...string) string {
	return strings.Join(key, "/")
}

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

func TestSplitScaledHostname(t *testing.T) {
	var name, scale string
	name, scale = utils.SplitScaledHostname("backend_10")
	if name != "backend" || scale != "10" {
		t.Errorf("Splitting is not working well")
	}

	name, scale = utils.SplitScaledHostname("rails_backend_10")
	if name != "rails_backend" || scale != "10" {
		t.Errorf("Splitting is not working well")
	}

	name, scale = utils.SplitScaledHostname("rails_1_backend_10")
	if name != "rails_1_backend" || scale != "10" {
		t.Errorf("Splitting is not working well")
	}

	name, scale = utils.SplitScaledHostname("rails_backend")
	if name != "rails_backend" || scale != "" {
		t.Errorf("Splitting is not working well")
	}
}
