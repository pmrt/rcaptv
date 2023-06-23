package helix

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestHelixClip(t *testing.T) {
	t.Parallel()
	clipsJson := []byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:33:59Z","creator_id":"809288340","creator_name":"NiviVT","duration":9,"embed_url":"https://clips.twitch.tv/embed?clip=CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","game_id":"21779","id":"CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3MsHosfc3X3iPWfF-7FUIQ/AT-cm%7C3MsHosfc3X3iPWfF-7FUIQ-preview-480x272.jpg","title":"apagando Windows","url":"https://clips.twitch.tv/CoweringDreamyOrcaGingerPower-x9zdfeI9Z8X7sVQh","video_id":"","view_count":10,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:34:53Z","creator_id":"809288340","creator_name":"NiviVT","duration":14.9,"embed_url":"https://clips.twitch.tv/embed?clip=FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","game_id":"21779","id":"FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/WP4c9ZjMKjjwjxhL9E89_g/AT-cm%7CWP4c9ZjMKjjwjxhL9E89_g-preview-480x272.jpg","title":"CUIDADO NIÑO","url":"https://clips.twitch.tv/FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","video_id":"","view_count":10,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:45:51Z","creator_id":"574315409","creator_name":"kiseorr","duration":20.6,"embed_url":"https://clips.twitch.tv/embed?clip=GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/u_8_ToJKlmSQMXTyOcahsA/AT-cm%7Cu_8_ToJKlmSQMXTyOcahsA-preview-480x272.jpg","title":"KEK","url":"https://clips.twitch.tv/GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","video_id":"","view_count":10,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T12:40:06Z","creator_id":"67005639","creator_name":"rodrifyify","duration":21.8,"embed_url":"https://clips.twitch.tv/embed?clip=GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","game_id":"21779","id":"GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/j0F2AU2edE0sKnlb0KQXRg/AT-cm%7Cj0F2AU2edE0sKnlb0KQXRg-preview-480x272.jpg","title":"ELM Y ZELING CORAZON ROTO :(","url":"https://clips.twitch.tv/GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","video_id":"","view_count":8,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T13:24:36Z","creator_id":"95615188","creator_name":"Thalekith","duration":18.7,"embed_url":"https://clips.twitch.tv/embed?clip=CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","game_id":"21779","id":"CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/93_qsOtPY1tWxsyzpJqz8Q/AT-cm%7C93_qsOtPY1tWxsyzpJqz8Q-preview-480x272.jpg","title":"Da gusto entrar al stream y que te reciban así","url":"https://clips.twitch.tv/CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","video_id":"","view_count":6,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T13:14:27Z","creator_id":"80189286","creator_name":"BestLeeMorocco","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","game_id":"21779","id":"LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/G70oOxvgyI9zKsb6xUAzHQ/46919399260-offset-14476-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","video_id":"","view_count":3,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T09:36:22Z","creator_id":"429634600","creator_name":"fabi42218","duration":30,"embed_url":"https://clips.twitch.tv/embed?clip=CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","game_id":"21779","id":"CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/8uy-Ah8RpX3mIXq0mNKing/46919399260-offset-1388-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T12:06:20Z","creator_id":"80767915","creator_name":"daniurlol","duration":26,"embed_url":"https://clips.twitch.tv/embed?clip=RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","game_id":"21779","id":"RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/plgZ6J7mZBhbQ5lVsnU_Ig/46923904428-offset-11834-preview-480x272.jpg","title":"me la shaco","url":"https://clips.twitch.tv/RacyResilientRhinocerosOSkomodo-nuZkGFtDmVWydT8i","video_id":"","view_count":1,"vod_offset":null}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7IkN1cnNvciI6Ik9BPT0ifX0"}}`)
	wantQuery := "broadcaster_id=58753574&first=100"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(clipsJson)
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		c: sv.Client(),
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
	}
	for i, clip := range clipsResp.Clips {
		got, want := clip.ClipID, want[i].ClipID
		if got != want {
			t.Fatalf("unexpected clip id got: %s, want %s", got, want)
		}
	}
}

func TestHelixClipEmpty(t *testing.T) {
	t.Parallel()
	clipsJson := []byte(`{"data":[],"pagination":{}}`)
	wantQuery := "broadcaster_id=58753574&first=100"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(clipsJson)
	}))
	defer sv.Close()

	hx := &Helix{
		opts: &HelixOpts{
			APIUrl: sv.URL,
		},
		c: sv.Client(),
	}
	clipsResp, err := hx.Clips(&ClipsParams{
		BroadcasterID:            "58753574",
		StopViewsThreshold:       8,
		ViewsThresholdWindowSize: 3,
	})
	if err != nil {
		if !errors.Is(err, ErrItemsEmpty) {
			t.Fatal(err)
		}
	} else {
		t.Fatal("expected err to be ErrItemsEmpty")
	}

	if len(clipsResp.Clips) != 0 {
		t.Fatal("expected 0 clips")
	}
}

func TestHelixClipDates(t *testing.T) {
	t.Parallel()
	// Note: this test is very unstable because if the streamer deletes some clip
	// it will fail Maybe think something better if that happens
	cid, cs := os.Getenv("TEST_CLIENT_ID"), os.Getenv("TEST_CLIENT_SECRET")
	if cid == "" || cs == "" {
		t.Skip("WARNING: TEST_CLIENT_ID and TEST_CLIENT_SECRET environment variables needed for this test, skipping. Re-run test with required environment variables.")
	}

	hx := New(&HelixOpts{
		APIUrl: "https://api.twitch.tv/helix",
		Creds: ClientCreds{
			ClientID:     cid,
			ClientSecret: cs,
		},
	})
	startedAt, err := time.Parse(time.RFC3339, "2023-06-04T11:45:51Z")
	if err != nil {
		t.Fatal(err)
	}
	endedAt, err := time.Parse(time.RFC3339, "2023-06-04T14:45:51Z")
	if err != nil {
		t.Fatal(err)
	}
	clipsResp, err := hx.Clips(&ClipsParams{
		BroadcasterID:            "58753574",
		StopViewsThreshold:       (385 + 507 + 823) / 3,
		ViewsThresholdWindowSize: 3,
		StartedAt:                startedAt,
		EndedAt:                  endedAt,
		First:                    5,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []*Clip{
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
	}
	for i, clip := range clipsResp.Clips {
		got, want := clip.ClipID, want[i].ClipID
		if got != want {
			t.Fatalf("unexpected clip id got: %s, want %s", got, want)
		}
	}
}
