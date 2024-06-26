package helix

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"

	"pedro.to/rcaptv/utils"
)

func TestRFC3339Timestamp(t *testing.T) {
	type testing struct {
		CreatedAt time.Time `json:"created_at"`
	}
	input := []byte(`{"created_at":"2021-04-17T17:23:52Z"}`)
	var ts *testing
	if err := json.Unmarshal(input, &ts); err != nil {
		t.Fatal(err)
	}

	want, err := time.Parse(time.RFC3339, "2021-04-17T17:23:52Z")
	if err != nil {
		t.Fatal(err)
	}

	if ts.CreatedAt != want {
		t.Fatalf("got %v, want %v", ts.CreatedAt, want)
	}

	res, err := json.Marshal(ts)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(input, res); diff != nil {
		t.Fatalf("got %v, want %v", res, input)
	}
}

func TestDeduplicate(t *testing.T) {
	t.Parallel()
	duplicated := []*Clip{
		{
			ClipID:          "FuriousTransparentWolfArsonNoSexy-w7lnpbTR8mHZQ0bN",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-04-17T17:23:52Z",
			CreatorID:       "412992173",
			CreatorName:     "maxiichh",
			DurationSeconds: 30,
			ViewCount:       36351,
		},
		{
			ClipID:          "EagerEnticingGoshawkOhMyDog-rJFkUxxaVBwcIIQJ",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-05-02T14:34:33Z",
			CreatorID:       "532931040",
			CreatorName:     "bygonleezz_27",
			DurationSeconds: 9,
			ViewCount:       28660,
		},
		{
			ClipID:          "RudeSeductiveRadishSSSsss-0_zvmvAv641XOIH5",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-05-02T14:35:55Z",
			CreatorID:       "484298161",
			CreatorName:     "mintyu0",
			DurationSeconds: 30,
			ViewCount:       20632,
		},
		{
			ClipID:          "RudeSeductiveRadishSSSsss-0_zvmvAv641XOIH5",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-05-02T14:35:55Z",
			CreatorID:       "484298161",
			CreatorName:     "mintyu0",
			DurationSeconds: 30,
			ViewCount:       20632,
		},
		{
			ClipID:          "TastyAgitatedDiscNinjaGrumpy",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2020-11-30T15:40:46Z",
			CreatorID:       "97024063",
			CreatorName:     "AleMagnoCA",
			DurationSeconds: 7,
			ViewCount:       20295,
		},
		{
			ClipID:          "TrustworthyEphemeralDunlinCharlietheUnicorn",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-01-14T16:54:11Z",
			CreatorID:       "190546482",
			CreatorName:     "F0RCE______",
			DurationSeconds: 59,
			ViewCount:       18964,
		},
		{
			ClipID:          "TrustworthyEphemeralDunlinCharlietheUnicorn",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-01-14T16:54:11Z",
			CreatorID:       "190546482",
			CreatorName:     "F0RCE______",
			DurationSeconds: 59,
			ViewCount:       18964,
		},
		{
			ClipID:          "EnergeticApatheticSlothStrawBeary-IQ1ny3mCS4S-_Hr4",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2022-01-08T21:06:37Z",
			CreatorID:       "235578058",
			CreatorName:     "Keygon8",
			DurationSeconds: 43,
			ViewCount:       16617,
		},
	}

	want := []*Clip{
		{
			ClipID:          "FuriousTransparentWolfArsonNoSexy-w7lnpbTR8mHZQ0bN",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-04-17T17:23:52Z",
			CreatorID:       "412992173",
			CreatorName:     "maxiichh",
			DurationSeconds: 30,
			ViewCount:       36351,
		},
		{
			ClipID:          "EagerEnticingGoshawkOhMyDog-rJFkUxxaVBwcIIQJ",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-05-02T14:34:33Z",
			CreatorID:       "532931040",
			CreatorName:     "bygonleezz_27",
			DurationSeconds: 9,
			ViewCount:       28660,
		},
		{
			ClipID:          "RudeSeductiveRadishSSSsss-0_zvmvAv641XOIH5",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-05-02T14:35:55Z",
			CreatorID:       "484298161",
			CreatorName:     "mintyu0",
			DurationSeconds: 30,
			ViewCount:       20632,
		},
		{
			ClipID:          "TastyAgitatedDiscNinjaGrumpy",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2020-11-30T15:40:46Z",
			CreatorID:       "97024063",
			CreatorName:     "AleMagnoCA",
			DurationSeconds: 7,
			ViewCount:       20295,
		},
		{
			ClipID:          "TrustworthyEphemeralDunlinCharlietheUnicorn",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-01-14T16:54:11Z",
			CreatorID:       "190546482",
			CreatorName:     "F0RCE______",
			DurationSeconds: 59,
			ViewCount:       18964,
		},
		{
			ClipID:          "EnergeticApatheticSlothStrawBeary-IQ1ny3mCS4S-_Hr4",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2022-01-08T21:06:37Z",
			CreatorID:       "235578058",
			CreatorName:     "Keygon8",
			DurationSeconds: 43,
			ViewCount:       16617,
		},
	}

	got := Deduplicate(duplicated, func(c *Clip) string {
		return c.ClipID
	})

	if len(got) != len(want) {
		t.Fatal("unexpected length")
	}

	if diff := deep.Equal(got, want); diff != nil {
		for _, line := range diff {
			t.Log(line)
		}
		t.Fatal("expected values to be equal")
	}
}

func TestDeduplicateConflict(t *testing.T) {
	t.Parallel()
	offset := 300
	duplicated := []*Clip{
		{
			ClipID:          "FuriousTransparentWolfArsonNoSexy-w7lnpbTR8mHZQ0bN",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-04-17T17:23:52Z",
			CreatorID:       "412992173",
			CreatorName:     "maxiichh",
			DurationSeconds: 30,
			ViewCount:       36351,
		},
		{
			ClipID:          "RudeSeductiveRadishSSSsss-0_zvmvAv641XOIH5",
			BroadcasterID:   "58753574",
			VideoID:         "video1",
			CreatedAt:       "2021-05-02T14:35:55Z",
			CreatorID:       "484298161",
			CreatorName:     "mintyu0",
			DurationSeconds: 30,
			ViewCount:       200,
		},
		{
			ClipID:          "RudeSeductiveRadishSSSsss-0_zvmvAv641XOIH5",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-05-02T14:35:55Z",
			CreatorID:       "484298161",
			CreatorName:     "mintyu0",
			DurationSeconds: 30,
			ViewCount:       1,
		},
		{
			ClipID:          "TastyAgitatedDiscNinjaGrumpy",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2020-11-30T15:40:46Z",
			CreatorID:       "97024063",
			CreatorName:     "AleMagnoCA",
			DurationSeconds: 7,
			ViewCount:       20295,
		},
		{
			ClipID:           "TrustworthyEphemeralDunlinCharlietheUnicorn",
			BroadcasterID:    "58753574",
			VideoID:          "",
			CreatedAt:        "2021-01-14T16:54:11Z",
			CreatorID:        "190546482",
			CreatorName:      "F0RCE______",
			DurationSeconds:  59,
			ViewCount:        10,
			VODOffsetSeconds: &offset,
		},
		{
			ClipID:          "TrustworthyEphemeralDunlinCharlietheUnicorn",
			BroadcasterID:   "58753574",
			VideoID:         "video2",
			CreatedAt:       "2021-01-14T16:54:11Z",
			CreatorID:       "190546482",
			CreatorName:     "F0RCE______",
			DurationSeconds: 59,
			ViewCount:       2000,
		},
	}
	clips := Deduplicate(duplicated,
		func(c *Clip) string {
			return c.ClipID
		},
		func(a *Clip, b *Clip) *Clip {
			a.VODOffsetSeconds = utils.Coalesce(a.VODOffsetSeconds, b.VODOffsetSeconds)
			a.VideoID = utils.CoalesceString(a.VideoID, b.VideoID)
			a.ViewCount = utils.Max(a.ViewCount, b.ViewCount)
			return a
		},
	)
	want := []*Clip{
		{
			ClipID:          "FuriousTransparentWolfArsonNoSexy-w7lnpbTR8mHZQ0bN",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2021-04-17T17:23:52Z",
			CreatorID:       "412992173",
			CreatorName:     "maxiichh",
			DurationSeconds: 30,
			ViewCount:       36351,
		},
		{
			ClipID:          "RudeSeductiveRadishSSSsss-0_zvmvAv641XOIH5",
			BroadcasterID:   "58753574",
			VideoID:         "video1",
			CreatedAt:       "2021-05-02T14:35:55Z",
			CreatorID:       "484298161",
			CreatorName:     "mintyu0",
			DurationSeconds: 30,
			ViewCount:       200,
		},
		{
			ClipID:          "TastyAgitatedDiscNinjaGrumpy",
			BroadcasterID:   "58753574",
			VideoID:         "",
			CreatedAt:       "2020-11-30T15:40:46Z",
			CreatorID:       "97024063",
			CreatorName:     "AleMagnoCA",
			DurationSeconds: 7,
			ViewCount:       20295,
		},
		{
			ClipID:           "TrustworthyEphemeralDunlinCharlietheUnicorn",
			BroadcasterID:    "58753574",
			VideoID:          "video2",
			CreatedAt:        "2021-01-14T16:54:11Z",
			CreatorID:        "190546482",
			CreatorName:      "F0RCE______",
			DurationSeconds:  59,
			ViewCount:        2000,
			VODOffsetSeconds: &offset,
		},
	}

	spew.Dump(clips)
	if len(clips) != 4 {
		t.Fatalf("expected 4 clips, got %d", len(clips))
	}

	if diff := deep.Equal(clips, want); diff != nil {
		t.Fatal(diff)
	}
}

func TestHelixCredentials(t *testing.T) {
	t.Parallel()
	cid, cs := os.Getenv("TEST_CLIENT_ID"), os.Getenv("TEST_CLIENT_SECRET")
	if cid == "" || cs == "" {
		t.Skip("WARNING: TEST_CLIENT_ID and TEST_CLIENT_SECRET environment variables needed for this test, skipping. Re-run test with required environment variables.")
	}

	hx := New(&HelixOpts{
		Creds: ClientCreds{
			ClientID:     cid,
			ClientSecret: cs,
		},
		APIUrl:           os.Getenv("TWITCH_API_URL"),
		EventsubEndpoint: "/eventsub",
	})

	if hx.defaultClient == nil {
		t.Fatal("client is empty")
	}

	endpoint := fmt.Sprintf("/users?login=%s", "alexelcapo")
	req, err := http.NewRequest("GET", hx.APIUrl()+endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Client-Id", hx.ClientID())

	resp, err := hx.defaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	wantJSON := []byte(`{"data":[{"id":"36138196","login":"alexelcapo","display_name":"alexelcapo","type":"","broadcaster_type":"partner","description":"Nací en el 87 y me gusta jugar a videojuegos.","profile_image_url":"https://static-cdn.jtvnw.net/jtv_user_pictures/78528288-6216-4e21-872b-7f415b602a9a-profile_image-300x300.png","offline_image_url":"https://static-cdn.jtvnw.net/jtv_user_pictures/bf455aac-4ce9-4daa-94a0-c6c0a1b2500d-channel_offline_image-1920x1080.png","view_count":79789494,"created_at":"2012-09-12T21:24:26Z"}]}`)
	// Check some fields that we know will most likely never change
	var got, want struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(wantJSON, &want); err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(got.Data[0], want.Data[0]); diff != nil {
		t.Fatal(diff)
	}

	if resp.Request.Header.Get("Authorization") == "" {
		t.Fatal("expected authorization request header to not be empty")
	}
}

func TestHelixCreateEventsubSubscription(t *testing.T) {
	t.Parallel()
	const (
		broadcasterid = "1234"
		cb            = "http://localhost/webhook"
		secret        = "thisisanososecretsecret"
	)

	wantJson := `{"type":"stream.online","version":"1","condition":{"broadcaster_user_id":"1234"},"transport":{"method":"webhook","callback":"http://localhost/webhook","secret":"thisisanososecretsecret"}}` + string('\n')

	var body []byte
	sv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Log(err)
		}
		body = b
	}))
	defer sv.Close()
	hx := &Helix{
		opts: &HelixOpts{
			APIUrl:           sv.URL,
			EventsubEndpoint: "/eventsub",
		},
		defaultClient: sv.Client(),
	}
	err := hx.CreateEventsubSubscription(&Subscription{
		Type:    SubStreamOnline,
		Version: "1",
		Condition: &Condition{
			BroadcasterUserID: broadcasterid,
		},
		Transport: &Transport{
			Method:   "webhook",
			Callback: cb,
			Secret:   secret,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, want := string(body), wantJson
	if got != want {
		t.Fatalf("got:\n\n%s (%d)\nwant:\n\n%s (%d)", got, len(got), want, len(want))
	}
}

func TestUntilRatelimitReset(t *testing.T) {
	t.Parallel()
	now := time.Now()
	resetAt := now.Add(time.Second * 10).Unix()
	reset, err := untilRatelimitReset(fmt.Sprint(resetAt), now)
	if err != nil {
		t.Fatal(err)
	}
	diff := reset - (time.Second * time.Duration(10))
	if diff.Abs() > time.Second {
		t.Fatal("expected reset delay to be within 1s of expected value")
	}
}

func TestHelixRateLimitedResiliency(t *testing.T) {
	t.Parallel()
	const (
		broadcasterid = "1234"
		cb            = "http://localhost/webhook"
		secret        = "thisisanososecretsecret"
	)

	resetAfter := time.Duration(3) * time.Second
	reqs := 0
	attempts := 0
	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		reqs++
		if attempts == 0 {
			attempts++
			now := time.Now()
			resp.Header().Set("Ratelimit-Reset", fmt.Sprint(now.Add(resetAfter).Unix()))
			resp.Header().Set("Date", now.UTC().Format(http.TimeFormat))
			resp.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer sv.Close()
	hx := &Helix{
		opts: &HelixOpts{
			APIUrl:           sv.URL,
			EventsubEndpoint: "/eventsub",
		},
		defaultClient: sv.Client(),
	}

	start := time.Now()
	err := hx.CreateEventsubSubscription(&Subscription{
		Type:    SubStreamOnline,
		Version: "1",
		Condition: &Condition{
			BroadcasterUserID: broadcasterid,
		},
		Transport: &Transport{
			Method:   "webhook",
			Callback: cb,
			Secret:   secret,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	end := time.Now()

	if reqs != 2 {
		t.Fatal("expected exactly 2 requests to server")
	}
	took := end.Sub(start)
	diff := took - (resetAfter + time.Second)
	if diff.Abs() > time.Millisecond*100 {
		t.Fatal("expected reset delay to be within 100ms of expected value")
	}
}

func TestHelixPagination(t *testing.T) {
	t.Parallel()
	clipsJson := [...][]byte{
		[]byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:33:59Z","creator_id":"809288340","creator_name":"NiviVT","duration":9,"embed_url":"https://clips.twitch.tv/embed?clip=CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","game_id":"21779","id":"CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3MsHosfc3X3iPWfF-7FUIQ/AT-cm%7C3MsHosfc3X3iPWfF-7FUIQ-preview-480x272.jpg","title":"apagando Windows","url":"https://clips.twitch.tv/CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:34:53Z","creator_id":"809288340","creator_name":"NiviVT","duration":14.9,"embed_url":"https://clips.twitch.tv/embed?clip=FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","game_id":"21779","id":"FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/WP4c9ZjMKjjwjxhL9E89_g/AT-cm%7CWP4c9ZjMKjjwjxhL9E89_g-preview-480x272.jpg","title":"CUIDADO NIÑO","url":"https://clips.twitch.tv/FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:45:51Z","creator_id":"574315409","creator_name":"kiseorr","duration":20.6,"embed_url":"https://clips.twitch.tv/embed?clip=GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/u_8_ToJKlmSQMXTyOcahsA/AT-cm%7Cu_8_ToJKlmSQMXTyOcahsA-preview-480x272.jpg","title":"KEK","url":"https://clips.twitch.tv/GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","video_id":"","view_count":1000,"vod_offset":null}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX0"}}`),
		[]byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T12:40:06Z","creator_id":"67005639","creator_name":"rodrifyify","duration":21.8,"embed_url":"https://clips.twitch.tv/embed?clip=GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","game_id":"21779","id":"GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/j0F2AU2edE0sKnlb0KQXRg/AT-cm%7Cj0F2AU2edE0sKnlb0KQXRg-preview-480x272.jpg","title":"ELM Y ZELING CORAZON ROTO :(","url":"https://clips.twitch.tv/GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T13:24:36Z","creator_id":"95615188","creator_name":"Thalekith","duration":18.7,"embed_url":"https://clips.twitch.tv/embed?clip=CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","game_id":"21779","id":"CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/93_qsOtPY1tWxsyzpJqz8Q/AT-cm%7C93_qsOtPY1tWxsyzpJqz8Q-preview-480x272.jpg","title":"Da gusto entrar al stream y que te reciban así","url":"https://clips.twitch.tv/CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T13:14:27Z","creator_id":"80189286","creator_name":"BestLeeMorocco","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","game_id":"21779","id":"LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/G70oOxvgyI9zKsb6xUAzHQ/46919399260-offset-14476-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","video_id":"","view_count":1000,"vod_offset":null}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX1"}}`),
		[]byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T09:36:22Z","creator_id":"429634600","creator_name":"fabi42218","duration":30,"embed_url":"https://clips.twitch.tv/embed?clip=CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","game_id":"21779","id":"CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/8uy-Ah8RpX3mIXq0mNKing/46919399260-offset-1388-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T12:06:20Z","creator_id":"80767915","creator_name":"daniurlol","duration":26,"embed_url":"https://clips.twitch.tv/embed?clip=RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","game_id":"21779","id":"RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/plgZ6J7mZBhbQ5lVsnU_Ig/46923904428-offset-11834-preview-480x272.jpg","title":"me la shaco","url":"https://clips.twitch.tv/RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","video_id":"","view_count":1000,"vod_offset":null}],"pagination":{"cursor":""}}`),
	}

	reqs := 0
	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		reqs++
		if got := r.URL.Query().Get("broadcaster_id"); got != "58753574" {
			t.Fatalf("bad broadcaster_id got: %s, want %s", got, "58753574")
		}
		if got := r.URL.Query().Get("first"); got != "100" {
			t.Fatalf("bad first got: %s, want %s", got, "100")
		}

		if !r.URL.Query().Has("after") {
			resp.Write(clipsJson[0])
		} else if r.URL.Query().Get("after") == "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX0" {
			resp.Write(clipsJson[1])
		} else if r.URL.Query().Get("after") == "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX1" {
			resp.Write(clipsJson[2])
		}
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		defaultClient: sv.Client(),
	}
	clipsResp, err := hx.Clips(&ClipsParams{
		BroadcasterID:            "58753574",
		StopViewsThreshold:       8,
		ViewsThresholdWindowSize: 3,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []*Clip{
		{
			ClipID: "CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh",
		},
		{
			ClipID: "FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-",
		},
		{
			ClipID: "GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw",
		},
		{
			ClipID: "GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG",
		},
		{
			ClipID: "CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx",
		},
		{
			ClipID: "LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd",
		},
		{
			ClipID: "CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe",
		},
		{
			ClipID: "RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i",
		},
	}
	for i, clip := range clipsResp.Clips {
		got, want := clip.ClipID, want[i].ClipID
		if got != want {
			t.Fatalf("unexpected clip id got: %s, want %s", got, want)
		}
	}

	if reqs != 3 {
		t.Fatalf("expected 3 requests, got %d", reqs)
	}
}

func TestHelixDeduplicatedPagination(t *testing.T) {
	t.Parallel()
	clipsJson := [...][]byte{
		[]byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:33:59Z","creator_id":"809288340","creator_name":"NiviVT","duration":9,"embed_url":"https://clips.twitch.tv/embed?clip=CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","game_id":"21779","id":"CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3MsHosfc3X3iPWfF-7FUIQ/AT-cm%7C3MsHosfc3X3iPWfF-7FUIQ-preview-480x272.jpg","title":"apagando Windows","url":"https://clips.twitch.tv/CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:34:53Z","creator_id":"809288340","creator_name":"NiviVT","duration":14.9,"embed_url":"https://clips.twitch.tv/embed?clip=FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","game_id":"21779","id":"FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/WP4c9ZjMKjjwjxhL9E89_g/AT-cm%7CWP4c9ZjMKjjwjxhL9E89_g-preview-480x272.jpg","title":"CUIDADO NIÑO","url":"https://clips.twitch.tv/FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:45:51Z","creator_id":"574315409","creator_name":"kiseorr","duration":20.6,"embed_url":"https://clips.twitch.tv/embed?clip=GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/u_8_ToJKlmSQMXTyOcahsA/AT-cm%7Cu_8_ToJKlmSQMXTyOcahsA-preview-480x272.jpg","title":"KEK","url":"https://clips.twitch.tv/GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","video_id":"","view_count":1000,"vod_offset":null}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX0"}}`),
		[]byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T12:40:06Z","creator_id":"67005639","creator_name":"rodrifyify","duration":21.8,"embed_url":"https://clips.twitch.tv/embed?clip=GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/j0F2AU2edE0sKnlb0KQXRg/AT-cm%7Cj0F2AU2edE0sKnlb0KQXRg-preview-480x272.jpg","title":"ELM Y ZELING CORAZON ROTO :(","url":"https://clips.twitch.tv/GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T13:24:36Z","creator_id":"95615188","creator_name":"Thalekith","duration":18.7,"embed_url":"https://clips.twitch.tv/embed?clip=CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","game_id":"21779","id":"CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/93_qsOtPY1tWxsyzpJqz8Q/AT-cm%7C93_qsOtPY1tWxsyzpJqz8Q-preview-480x272.jpg","title":"Da gusto entrar al stream y que te reciban así","url":"https://clips.twitch.tv/CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T13:14:27Z","creator_id":"80189286","creator_name":"BestLeeMorocco","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","game_id":"21779","id":"LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/G70oOxvgyI9zKsb6xUAzHQ/46919399260-offset-14476-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","video_id":"","view_count":1000,"vod_offset":null}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX1"}}`),
		[]byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T09:36:22Z","creator_id":"429634600","creator_name":"fabi42218","duration":30,"embed_url":"https://clips.twitch.tv/embed?clip=CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","game_id":"21779","id":"CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/8uy-Ah8RpX3mIXq0mNKing/46919399260-offset-1388-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","video_id":"","view_count":1000,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T12:06:20Z","creator_id":"80767915","creator_name":"daniurlol","duration":26,"embed_url":"https://clips.twitch.tv/embed?clip=RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","game_id":"21779","id":"RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/plgZ6J7mZBhbQ5lVsnU_Ig/46923904428-offset-11834-preview-480x272.jpg","title":"me la shaco","url":"https://clips.twitch.tv/RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","video_id":"","view_count":1000,"vod_offset":null}],"pagination":{"cursor":""}}`),
	}

	reqs := 0
	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		reqs++
		if got := r.URL.Query().Get("broadcaster_id"); got != "58753574" {
			t.Fatalf("bad broadcaster_id got: %s, want %s", got, "58753574")
		}
		if got := r.URL.Query().Get("first"); got != "100" {
			t.Fatalf("bad first got: %s, want %s", got, "100")
		}

		if !r.URL.Query().Has("after") {
			resp.Write(clipsJson[0])
		} else if r.URL.Query().Get("after") == "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX0" {
			resp.Write(clipsJson[1])
		} else if r.URL.Query().Get("after") == "eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX1" {
			resp.Write(clipsJson[2])
		}
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		defaultClient: sv.Client(),
	}
	clipsResp, err := hx.Clips(&ClipsParams{
		BroadcasterID:            "58753574",
		StopViewsThreshold:       8,
		ViewsThresholdWindowSize: 3,
		SkipDeduplication:        true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(clipsResp.Clips) != 8 {
		t.Fatalf("expected 8 clips with SkipDeduplication=true, got:%d", len(clipsResp.Clips))
	}

	clipsResp, err = hx.Clips(&ClipsParams{
		BroadcasterID:            "58753574",
		StopViewsThreshold:       8,
		ViewsThresholdWindowSize: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []*Clip{
		{
			ClipID: "CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh",
		},
		{
			ClipID: "FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-",
		},
		{
			ClipID: "GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw",
		},
		{
			ClipID: "CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx",
		},
		{
			ClipID: "LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd",
		},
		{
			ClipID: "CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe",
		},
		{
			ClipID: "RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i",
		},
	}
	got := clipsResp.Clips
	if len(got) != len(want) {
		t.Fatal("unexpected length")
	}
	for i, clip := range clipsResp.Clips {
		got, want := clip.ClipID, want[i].ClipID
		if got != want {
			t.Fatalf("unexpected clip id got: %s, want %s", got, want)
		}
	}

	if reqs != 6 {
		t.Fatalf("expected 3 requests, got %d", reqs)
	}
}
