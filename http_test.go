package middleware

import (
	"regexp"
	"testing"
)

func TestRegisterDefaultIndex(t *testing.T) {
	//RegisterDefaultIndex("", nil, "", nil, nil, "", true)
}

func TestDistReg(t *testing.T) {
	exp := regexp.MustCompile(`\.html$|\.js$|\.jsx$|\.ts$|\.tsx$|\.css$|\.svg$|\.icon$|\.ico$|\.png$|\.jpg$|\.jpeg$|\.gif$`)
	println(exp.MatchString("index.jsx"))
}
