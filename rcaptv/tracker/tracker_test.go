package tracker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pedro.to/rcaptv/helix"
)

func TestDeepFetchClips(t *testing.T) {
	/*
																	X = incomplete
									X
			|-------------------- 168h (A) --------------------|   lvl1
		              X
			|-------- 0-83h (B) -----||------- 84-168h (C) ----|   lvl2

			|- 0-41h (D)-||-42-83h(E)|                             lvl3

			- A: A single clip with lots of views to force IsComplete=false
			- B: 6 clips with lots of views to force IsComplete=false
			- C: 6 clips with 1 view, IsComplete=true
			- D and E: 3 clips each, repeated from A and B but truncated with 1 view
			IsComplete=true. In a real case it would just return more clips with low
			views

			So we should get D, E and C clips with views that meet the
			average view window threshold
	*/
	jsonA := []byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T00:01:00Z","creator_id":"809288340","creator_name":"NiviVT","duration":14.9,"embed_url":"https://clips.twitch.tv/embed?clip=FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","game_id":"21779","id":"FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/WP4c9ZjMKjjwjxhL9E89_g/AT-cm%7CWP4c9ZjMKjjwjxhL9E89_g-preview-480x272.jpg","title":"CUIDADO NIÑO","url":"https://clips.twitch.tv/FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","video_id":"","view_count":925,"vod_offset":null}],"pagination":{"cursor":""}}`)
	jsonB := []byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T00:01:00Z","creator_id":"809288340","creator_name":"NiviVT","duration":14.9,"embed_url":"https://clips.twitch.tv/embed?clip=FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","game_id":"21779","id":"FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/WP4c9ZjMKjjwjxhL9E89_g/AT-cm%7CWP4c9ZjMKjjwjxhL9E89_g-preview-480x272.jpg","title":"CUIDADO NIÑO","url":"https://clips.twitch.tv/FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","video_id":"","view_count":925,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:45:51Z","creator_id":"574315409","creator_name":"kiseorr","duration":20.6,"embed_url":"https://clips.twitch.tv/embed?clip=GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/u_8_ToJKlmSQMXTyOcahsA/AT-cm%7Cu_8_ToJKlmSQMXTyOcahsA-preview-480x272.jpg","title":"KEK","url":"https://clips.twitch.tv/GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","video_id":"","view_count":829,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-05T03:40:06Z","creator_id":"67005639","creator_name":"rodrifyify","duration":21.8,"embed_url":"https://clips.twitch.tv/embed?clip=GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","game_id":"21779","id":"GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/j0F2AU2edE0sKnlb0KQXRg/AT-cm%7Cj0F2AU2edE0sKnlb0KQXRg-preview-480x272.jpg","title":"ELM Y ZELING CORAZON ROTO :(","url":"https://clips.twitch.tv/GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","video_id":"","view_count":507,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:24:36Z","creator_id":"95615188","creator_name":"Thalekith","duration":18.7,"embed_url":"https://clips.twitch.tv/embed?clip=CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","game_id":"21779","id":"CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/93_qsOtPY1tWxsyzpJqz8Q/AT-cm%7C93_qsOtPY1tWxsyzpJqz8Q-preview-480x272.jpg","title":"Da gusto entrar al stream y que te reciban así","url":"https://clips.twitch.tv/CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","video_id":"","view_count":385,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:44:27Z","creator_id":"80189286","creator_name":"BestLeeMorocco","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","game_id":"21779","id":"LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/G70oOxvgyI9zKsb6xUAzHQ/46919399260-offset-14476-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","video_id":"","view_count":269,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-07T05:36:22Z","creator_id":"429634600","creator_name":"fabi42218","duration":30,"embed_url":"https://clips.twitch.tv/embed?clip=CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","game_id":"21779","id":"CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/8uy-Ah8RpX3mIXq0mNKing/46919399260-offset-1388-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","video_id":"","view_count":217,"vod_offset":null}],"pagination":{"cursor":""}}`)
	jsonC := []byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-08T12:14:55Z","creator_id":"648661011","creator_name":"kevinknight619","duration":10.1,"embed_url":"https://clips.twitch.tv/embed?clip=TsundereOddSrirachaDendiFace-MifT_BPTFolNTpGN","game_id":"21779","id":"TsundereOddSrirachaDendiFace-MifT_BPTFolNTpGN","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/Tne9OTIn0_wT0xaJTU6RjQ/AT-cm%7CTne9OTIn0_wT0xaJTU6RjQ-preview-480x272.jpg","title":"Manoling? ","url":"https://clips.twitch.tv/TsundereOddSrirachaDendiFace-MifT_BPTFolNTpGN","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-08T13:32:23Z","creator_id":"183425734","creator_name":"OaksitoUwu","duration":5,"embed_url":"https://clips.twitch.tv/embed?clip=CloudyVictoriousPanKappaClaus-7_3O0CV6-V3-NUaF","game_id":"21779","id":"CloudyVictoriousPanKappaClaus-7_3O0CV6-V3-NUaF","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/BbV3eu_9ztAlevld_jVGhQ/AT-cm%7CBbV3eu_9ztAlevld_jVGhQ-preview-480x272.jpg","title":"zelin roleadora del año ","url":"https://clips.twitch.tv/CloudyVictoriousPanKappaClaus-7_3O0CV6-V3-NUaF","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-08T15:28:25Z","creator_id":"571226401","creator_name":"kucusumusu_","duration":6,"embed_url":"https://clips.twitch.tv/embed?clip=ColorfulCarelessLemurWholeWheat-7muTGbovfSdq8YyQ","game_id":"21779","id":"ColorfulCarelessLemurWholeWheat-7muTGbovfSdq8YyQ","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/kFHUzO5rkxm6L30navGMkA/AT-cm%7CkFHUzO5rkxm6L30navGMkA-preview-480x272.jpg","title":"wtf con la e","url":"https://clips.twitch.tv/ColorfulCarelessLemurWholeWheat-7muTGbovfSdq8YyQ","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-10T13:23:10Z","creator_id":"809288340","creator_name":"NiviVT","duration":21.7,"embed_url":"https://clips.twitch.tv/embed?clip=ShortBlightedRadicchioKappaClaus-n1E58A7CPN0PTRNo","game_id":"21779","id":"ShortBlightedRadicchioKappaClaus-n1E58A7CPN0PTRNo","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/g7TWRJ9ROfAOwjzdD00TaA/AT-cm%7Cg7TWRJ9ROfAOwjzdD00TaA-preview-480x272.jpg","title":"PLATA","url":"https://clips.twitch.tv/ShortBlightedRadicchioKappaClaus-n1E58A7CPN0PTRNo","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-10T14:12:37Z","creator_id":"652653268","creator_name":"lxjuanl","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=AmazonianHeartlessCheetahPRChase-TSuVHmK3_JC3x36B","game_id":"21779","id":"AmazonianHeartlessCheetahPRChase-TSuVHmK3_JC3x36B","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/XdARuE0rlR5it_yQt8d-gQ/AT-cm%7CXdARuE0rlR5it_yQt8d-gQ-preview-480x272.jpg","title":"Ideas pa san valentin","url":"https://clips.twitch.tv/AmazonianHeartlessCheetahPRChase-TSuVHmK3_JC3x36B","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-10T14:15:37Z","creator_id":"652653268","creator_name":"lxjuanl","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=AmazonianHeartlessCheetahPRChase-TSuVHmK3_JC3x36B","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/XdARuE0rlR5it_yQt8d-gQ/AT-cm%7CXdARuE0rlR5it_yQt8d-gQ-preview-480x272.jpg","title":"Ideas pa san valentin","url":"https://clips.twitch.tv/AmazonianHeartlessCheetahPRChase-TSuVHmK3_JC3x36B","video_id":"","view_count":1,"vod_offset":null}],"pagination":{"cursor":""}}`)
	jsonD := []byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T00:01:00Z","creator_id":"809288340","creator_name":"NiviVT","duration":14.9,"embed_url":"https://clips.twitch.tv/embed?clip=FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","game_id":"21779","id":"FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/WP4c9ZjMKjjwjxhL9E89_g/AT-cm%7CWP4c9ZjMKjjwjxhL9E89_g-preview-480x272.jpg","title":"CUIDADO NIÑO","url":"https://clips.twitch.tv/FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-04T11:45:51Z","creator_id":"574315409","creator_name":"kiseorr","duration":20.6,"embed_url":"https://clips.twitch.tv/embed?clip=GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","game_id":"21779","id":"GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/u_8_ToJKlmSQMXTyOcahsA/AT-cm%7Cu_8_ToJKlmSQMXTyOcahsA-preview-480x272.jpg","title":"KEK","url":"https://clips.twitch.tv/GlutenFreeCourteousPineappleUncleNox-gkqWZJAxdPI5xqGw","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-05T03:40:06Z","creator_id":"67005639","creator_name":"rodrifyify","duration":21.8,"embed_url":"https://clips.twitch.tv/embed?clip=GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","game_id":"21779","id":"GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/j0F2AU2edE0sKnlb0KQXRg/AT-cm%7Cj0F2AU2edE0sKnlb0KQXRg-preview-480x272.jpg","title":"ELM Y ZELING CORAZON ROTO :(","url":"https://clips.twitch.tv/GenerousGrossHyenaAllenHuhu-TZ50TSwqeVvQpBdG","video_id":"","view_count":1,"vod_offset":null}],"pagination":{"cursor":""}}`)
	jsonE := []byte(`{"data":[{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:24:36Z","creator_id":"95615188","creator_name":"Thalekith","duration":18.7,"embed_url":"https://clips.twitch.tv/embed?clip=CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","game_id":"21779","id":"CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/93_qsOtPY1tWxsyzpJqz8Q/AT-cm%7C93_qsOtPY1tWxsyzpJqz8Q-preview-480x272.jpg","title":"Da gusto entrar al stream y que te reciban así","url":"https://clips.twitch.tv/CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-06T13:44:27Z","creator_id":"80189286","creator_name":"BestLeeMorocco","duration":28,"embed_url":"https://clips.twitch.tv/embed?clip=LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","game_id":"21779","id":"LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/G70oOxvgyI9zKsb6xUAzHQ/46919399260-offset-14476-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/LuckyBrainyNikudonSMOrc-G-DIj3MqxvrFQDMd","video_id":"","view_count":1,"vod_offset":null},{"broadcaster_id":"58753574","broadcaster_name":"Zeling","created_at":"2023-06-07T05:36:22Z","creator_id":"429634600","creator_name":"fabi42218","duration":30,"embed_url":"https://clips.twitch.tv/embed?clip=CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","game_id":"21779","id":"CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/8uy-Ah8RpX3mIXq0mNKing/46919399260-offset-1388-preview-480x272.jpg","title":"😈Ratilla pelirroja GANANDO TODO 1 VS 9 😈  DIA DE NO ENFADOS I AKALI 1 VS 37 dias de SEASON","url":"https://clips.twitch.tv/CoyCogentLapwingOhMyDog-HxGlfeYherSY0qKe","video_id":"","view_count":1,"vod_offset":null}],"pagination":{"cursor":""}}`)
	emptyJson := []byte(`{"data":[],"pagination":{"cursor":""}}`)

	startA, err := time.Parse(time.RFC3339, "2023-06-04T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	endA := startA.Add(168 * time.Hour)

	startB := startA
	endB := startB.Add(84 * time.Hour)

	startC := endB
	endC := endA

	startD := startA
	endD := startD.Add(42 * time.Hour)

	startE := endD
	endE := endB

	reqs := 0
	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		reqs++
		s := r.URL.Query().Get("started_at")
		e := r.URL.Query().Get("ended_at")
		bid := r.URL.Query().Get("broadcaster_id")
		start, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		end, err := time.Parse(time.RFC3339, e)
		if err != nil {
			t.Fatal(err)
		}
		if bid != "58753574" {
			t.Fatalf("expected bid:58753574, got: %s", bid)
		}

		if startA.Equal(start) && endA.Equal(end) {
			resp.Write(jsonA)
		} else if startB.Equal(start) && endB.Equal(end) {
			resp.Write(jsonB)
		} else if startC.Equal(start) && endC.Equal(end) {
			resp.Write(jsonC)
		} else if startD.Equal(start) && endD.Equal(end) {
			resp.Write(jsonD)
		} else if startE.Equal(start) && endE.Equal(end) {
			resp.Write(jsonE)
		} else {
			t.Fatalf("unexpected date start:%s end:%s", s, e)
			resp.Write(emptyJson)
		}
	}))
	defer sv.Close()

	hx := helix.NewWithoutExchange(&helix.HelixOpts{
		APIUrl: sv.URL,
	}, sv.Client())
	tracker := &Tracker{
		ctx:                      context.Background(),
		hx:                       hx,
		ClipTrackingMaxDeepLevel: 3,
		ClipTrackingWindowHours:  168,
		ClipViewThreshold:        8,
		ClipViewWindowSize:       3,
	}

	clips, err := tracker.deepFetchClips("58753574", 1, startA, endA)
	if err != nil {
		t.Fatal(err)
	}

	if reqs != 5 {
		t.Fatalf("unexpected number of requests, got:%d, want:5", reqs)
	}

	want := []helix.Clip{
		// From D
		{ClipID: "FriendlyUninterestedLlamaTriHard-mMwqOPCPGEv0Tz3-"},
		// From E
		{ClipID: "CogentSpunkyChinchillaPJSugar-609jW1bGzLOkmrPx"},
		// From C
		{ClipID: "TsundereOddSrirachaDendiFace-MifT_BPTFolNTpGN"},
	}
	for i, clip := range clips {
		if got := clip.ClipID; got != want[i].ClipID {
			t.Fatalf("unexpected clip, got:%s want:%s", got, want[i].ClipID)
		}
	}
}

func TestFetchVods(t *testing.T) {
	vodsJson := []byte(`{"data":[{"created_at":"2023-06-14T23:21:38Z","description":"","duration":"1h6m40s","id":"1846472757","language":"es","muted_segments":null,"published_at":"2023-06-14T23:21:38Z","stream_id":"46935025356","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM","type":"archive","url":"https://www.twitch.tv/videos/1846472757","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":7865,"viewable":"public"},{"created_at":"2023-06-14T07:21:23Z","description":"","duration":"5h55m50s","id":"1845909865","language":"es","muted_segments":null,"published_at":"2023-06-14T07:21:23Z","stream_id":"46932511084","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡","type":"archive","url":"https://www.twitch.tv/videos/1845909865","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":39695,"viewable":"public"},{"created_at":"2023-06-13T07:21:20Z","description":"","duration":"8h34m10s","id":"1845060937","language":"es","muted_segments":null,"published_at":"2023-06-13T07:21:20Z","stream_id":"39674758421","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e50667e1d13ad2e09b4f_zeling_39674758421_1686640875//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I TIER 2  RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡 stream CORTO","type":"archive","url":"https://www.twitch.tv/videos/1845060937","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":53928,"viewable":"public"},{"created_at":"2023-06-12T17:05:21Z","description":"","duration":"2h9m50s","id":"1844500473","language":"es","muted_segments":null,"published_at":"2023-06-12T17:05:21Z","stream_id":"39673161157","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/2da3c11328ff9ec12e1d_zeling_39673161157_1686589516//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844500473","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":17792,"viewable":"public"},{"created_at":"2023-06-12T06:37:33Z","description":"","duration":"7h56m10s","id":"1844222546","language":"es","muted_segments":null,"published_at":"2023-06-12T06:37:33Z","stream_id":"39671417045","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/e4dae1bdae5f2a390746_zeling_39671417045_1686551846//thumb/thumb0-%{width}x%{height}.jpg","title":"☣️DROPS☣️BELLUM I RATILLA PELIRROJA VENGATIVA 💩 NOS HAN RAIDEADO 😡","type":"archive","url":"https://www.twitch.tv/videos/1844222546","user_id":"58753574","user_login":"zeling","user_name":"Zeling","view_count":58274,"viewable":"public"}],"pagination":{"cursor":"eyJiIjpudWxsLCJhIjp7Ik9mZnNldCI6NX19"}}`)
	wantQuery := "first=1&period=week&type=archive&user_id=58753574"
	bid := "58753574"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != wantQuery {
			t.Fatalf("bad query got: %s, want %s", r.URL.RawQuery, wantQuery)
		}
		resp.Write(vodsJson)
	}))
	defer sv.Close()

	hx := helix.NewWithoutExchange(&helix.HelixOpts{
		APIUrl: sv.URL,
	}, sv.Client())
	tracker := &Tracker{
		ctx:               context.Background(),
		hx:                hx,
		lastVIDByStreamer: make(lastVODTable, 20),
	}

	// test empty lastVODs table
	_, err := tracker.FetchVods(bid)
	if err != nil {
		t.Fatal(err)
	}
	want := "1846472757"
	if got := tracker.lastVIDByStreamer[bid]; got != want {
		t.Fatalf("expected lastVOD to be %s, got %s", want, got)
	}

	// test again with a older lastVOD in the table
	wantQuery = "first=100&period=week&type=archive&user_id=58753574"
	tracker.lastVIDByStreamer.Set(bid, "1845060937")
	vods, err := tracker.FetchVods(bid)
	if err != nil {
		t.Fatal(err)
	}
	if len(vods) != 3 {
		t.Fatalf("expected exactly 3 vods, got %d", len(vods))
	}
	wantVods := []string{"1846472757", "1845909865", "1845060937"}
	for i, vod := range vods {
		got := vod.VideoID
		want := wantVods[i]
		if got != want {
			t.Fatalf("expected vod %d to be %s, got %s", i, want, got)
		}
	}
}
