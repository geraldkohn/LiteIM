package cronjob

import (
	"fmt"
	"testing"
	"time"
)

func TestJobRun(t *testing.T) {
	j := NewJober("job-1", func() {
		fmt.Printf("%d", time.Now().Unix())
	}, time.Duration(5 * time.Second), 0)
	j.Run()
}
