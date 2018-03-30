package appError

import (
	"log"
)

/* Most of the places, we will import appError package.
   So in order to log something, we need not import another package "log/fmt" and can use this function directly.
   In future, this can also be customized based on our requirement eg:- writing to different files for different case.
*/
func Debug(v ...interface{}) {
	log.Println(v...)
}
