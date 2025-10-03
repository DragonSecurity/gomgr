package gh

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/v75/github"
)

func RespectRate(ctx context.Context, c *github.Client) error {
	r, _, err := c.RateLimits(ctx)
	if err != nil {
		return nil
	}
	if core := r.GetCore(); core.Remaining < 50 {
		sleep := time.Until(core.Reset.Time)
		log.Printf("rate-limit: sleeping until %s", core.Reset.Time)
		time.Sleep(sleep + time.Second)
	}
	return nil
}
