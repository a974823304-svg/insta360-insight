package mock

import "time"

// timeNow 单独提出来便于单测时 mock
var timeNow = time.Now
