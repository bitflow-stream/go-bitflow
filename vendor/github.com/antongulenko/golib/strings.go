package golib

import (
	"github.com/lunixbochs/vtclean"
	"golang.org/x/text/unicode/norm"
)

func StringLength(str string) (strlen int) {
	str = vtclean.Clean(str, false)
	var ia norm.Iter
	ia.InitString(norm.NFKD, str)
	for ; !ia.Done(); ia.Next() {
		strlen += 1
	}
	// Alternative:
	// strlen = utf8.RuneCountInString(str)
	return
}
