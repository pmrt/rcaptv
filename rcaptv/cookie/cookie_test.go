package cookie

import (
	"testing"
	"time"
)

func TestCookieAddSet(t *testing.T) {
	c := New()

	c.Add("token", "abcd")

	want := "token=abcd"
	got := c.String()
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}

	c.Add("access", "1234")
	want = "access=1234&token=abcd"
	got = c.String()
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}

	c.Set("token", "efgh")
	want = "access=1234&token=efgh"
	got = c.String()
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}

func TestCookieFromString(t *testing.T) {
	s := "token=abcd&access=1234"
	c := FromString(s)

	want := "access=1234&token=abcd"
	got := c.String()
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}

	want = "abcd"
	got = c.Get("token")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}

	want = "1234"
	got = c.Get("access")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}

func TestCookieTime(t *testing.T) {
	s := "expiry=2023-06-26T15%3A07%3A30Z"
	c := FromString(s)

	ts, err := time.Parse(time.RFC3339, "2023-06-26T15:07:30Z")
	if err != nil {
		t.Fatal(err)
	}

	want := ts
	got := c.GetTime("expiry")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}

	ts2, err := time.Parse(time.RFC3339, "2023-06-26T16:07:30Z")
	if err != nil {
		t.Fatal(err)
	}
	c.AddTime("created", ts2)
	want = ts2
	got = c.GetTime("created")
	if got != want {
		t.Fatalf("want %s got %s", want, got)
	}
}
