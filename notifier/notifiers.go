package notifier

import (
	"context"
	"log"
	"time"

	chkr "github.com/tweithoener/checker"
)

func Logging(prefix string) chkr.Notifier {
	return func(_ context.Context, name string, h chkr.History) {
		log.Printf("%s%s: %s: %s (%d times since %s)", prefix, h.State(), name, h.Message(), h.Streak(), h.Since().Local().Format("2006-01-02 15:04:05"))
	}
}

func Less(n chkr.Notifier) chkr.Notifier {
	var last time.Time
	return func(ctx context.Context, name string, h chkr.History) {
		if time.Since(last) > 1*time.Hour || h.Streak() <= 3 {
			last = time.Now()
			n(ctx, name, h)
		}
	}
}
