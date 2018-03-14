package utility

import (
	"strconv"
	"strings"
)

func IntArrayToString(a []int, separator string) string {
    b := make([]string, len(a))
    for i, v := range a {
        b[i] = strconv.Itoa(v)
    }

    return strings.Join(b, separator)
}


func StringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}
