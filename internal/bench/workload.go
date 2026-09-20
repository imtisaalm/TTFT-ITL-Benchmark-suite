package bench

import "fmt"

func Prompt(base string, requestID int, sharedPrefix bool) string {
	marker := fmt.Sprintf("[benchmark-request=%d]", requestID)
	if sharedPrefix {
		return base + "
" + marker
	}
	return marker + "
" + base
}
