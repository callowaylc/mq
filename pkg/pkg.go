package pkg

import (
	"fmt"
)

// constants ////////////////////////////////////

const PROJECT string = "github.com/callowaylc/mq"

// functions ////////////////////////////////////

func Trace(function, pkg string) string {
  return fmt.Sprintf(
    "%s#%s.%s@%s", PROJECT, "main", function, pkg,
  )
}
