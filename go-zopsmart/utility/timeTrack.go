package utility

import (
	"log"
	"time"
)

/* Will help us track the time taken. Use it at any places
	 eg:- defer timeTrack(time.Now(), "factorial")
*/	 
func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}