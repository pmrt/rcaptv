package cookie

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Cookie struct {
	claims url.Values
}

func (c *Cookie) Get(claimKey ClaimKey) string {
	return c.claims.Get(string(claimKey))
}

func (c *Cookie) GetTime(claimKey ClaimKey) time.Time {
	str := c.claims.Get(string(claimKey))
	if str == "" {
		return time.Time{}
	}

	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return time.Time{}
	}
	return t
}

func (c *Cookie) Set(claimKey ClaimKey, claimValue string) *Cookie {
	c.claims.Set(string(claimKey), claimValue)
	return c
}

func (c *Cookie) SetTime(claimKey ClaimKey, t time.Time) *Cookie {
	c.claims.Set(string(claimKey), t.Format(time.RFC3339))
	return c
}

func (c *Cookie) SetInt64(claimKey ClaimKey, i int64) *Cookie {
	c.claims.Set(string(claimKey), strconv.FormatInt(i, 10))
	return c
}

func (c *Cookie) AddInt64(claimKey ClaimKey, i int64) *Cookie {
	c.claims.Add(string(claimKey), strconv.FormatInt(i, 10))
	return c
}

func (c *Cookie) GetInt64(claimKey ClaimKey) int64 {
	str := c.claims.Get(string(claimKey))
	if str == "" {
		return 0
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func (c *Cookie) Add(claimKey ClaimKey, claimValue string) *Cookie {
	c.claims.Add(string(claimKey), claimValue)
	return c
}

func (c *Cookie) AddTime(claimKey ClaimKey, t time.Time) *Cookie {
	c.claims.Add(string(claimKey), t.Format(time.RFC3339))
	return c
}

func (c *Cookie) String() string {
	return c.claims.Encode()
}

func (c *Cookie) IsEmpty() bool {
	return c.String() == ""
}

func FromString(cookie string) *Cookie {
	if cookie == "" {
		return New()
	}

	claims, err := url.ParseQuery(cookie)
	if err != nil {
		return New()
	}
	return &Cookie{
		claims: claims,
	}
}

func New() *Cookie {
	return &Cookie{
		claims: url.Values{},
	}
}

func Fiber(c *fiber.Ctx, k string) *Cookie {
	return FromString(c.Cookies(k))
}
