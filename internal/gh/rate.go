package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v84/github"

	"github.com/DragonSecurity/gomgr/internal/util"
)

func RespectRate(ctx context.Context, c *github.Client) error {
	r, _, err := c.RateLimit.Get(ctx)
	if err != nil {
		return fmt.Errorf("rate limit check: %w", err)
	}
	if r == nil {
		return nil
	}
	if core := r.GetCore(); core.Remaining < 50 {
		sleep := time.Until(core.Reset.Time) + time.Second
		util.Infof("rate-limit: sleeping %s until %s", sleep, core.Reset.Time)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
