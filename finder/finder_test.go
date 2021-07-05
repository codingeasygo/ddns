package finder

import (
	"fmt"
	"testing"
)

func TestWeb(t *testing.T) {
	web := NewFinderServer("test", "https://api.ipify.org/?format=json", "/ip")
	fmt.Println(web.Discover("", ""))
}
