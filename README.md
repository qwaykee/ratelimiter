# ratelimiter
A simple rate limiter middleware for Gin Gonic

# usage
```golang
import (
	"github.com/gin-gonic/gin"
	"github.com/qwaykee/ratelimiter"
)

func main() {
	r := gin.Default()

	r.Use(ratelimiter.Default().Middleware())

	// ...rest of the code
}
```

# advanced usage

```golang
func main() {
	r := gin.Default()

	rl := ratelimiter.New(ratelimiter.Options{
		Limit: 50,
		Rate: time.Second,
		Key: func(c *gin.Context) string {
			return c.ClientIP() // or your key logic
		},
		WhenLimitReached: func(c *gin.Context, client *ratelimiter.Client) {
			client.Ban(10 * time.Minute)
			c.String(http.StatusTooManyRequests, "You are getting banned until" + client.BannedUntil)
			c.AbortWithStatus(http.StatusTooManyRequests)
		},
	})

	r.GET("/rate-limited", rl.Middleware(), func(c *gin.Context) {
		// ...
	})
	
	r.GET("/ban", func(c *gin.Context) {
		rl.Client(rl.Key(c)).Ban(2 * time.Minute)
	})
	
	r.GET("/ban-infos", func(c *gin.Context) {
		client := rl.Client(rl.Key(c))

		c.JSON(http.StatusOK, gin.H{
			"isBanned": client.IsBanned,
			"unbanDate": client.BannedUntil,
			"visitsDuringRateWindow": client.Visits,
		})
	})

	// ...rest of the code
}
```

more informations in the [docs](https://pkg.go.dev/github.com/qwaykee/ratelimiter)