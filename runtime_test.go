package middleware

import "testing"

func TestRuntime(t *testing.T) {
	nf := GetNetInfo()
	println(nf.CloseWaits)
}