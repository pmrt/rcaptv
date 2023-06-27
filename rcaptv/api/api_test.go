package api

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/nsf/jsondiff"
	"pedro.to/rcaptv/test"
)

var db *sql.DB

func TestMain(m *testing.M) {
	conn, pool, res := test.SetupPostgres()
	db = conn

	// Run tests
	code := m.Run()

	if err := test.CancelPostgres(pool, res); err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func TestVods(t *testing.T) {
	t.Parallel()
	wantJson := []byte(`{"data":{"vods":[{"id":"1847800606","user_id":"90075649","stream_id":"46940301884","created_at":"2023-06-16T15:36:48Z","published_at":"2023-06-16T15:36:48Z","language":"es","title":"[💀 𝙂𝙊𝙊𝙁𝙔 𝘼𝙎𝙎 𝘿𝙍𝙊𝙋𝙎 💀] DÍA 6: DÍA 1 🌈 - Bellum #6","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/82d5aaf2650410948650_illojuan_46940301884_1686929802//thumb/thumb0-%{width}x%{height}.jpg","view_count":970227,"duration_seconds":24770},{"id":"1846954069","user_id":"90075649","stream_id":"46936407228","created_at":"2023-06-15T15:10:59Z","published_at":"2023-06-15T15:10:59Z","language":"es","title":"[𝙆𝙀𝘽𝘼𝘽 𝘿𝙍𝙊𝙋𝙎] PROBAMOS LA ROG ALLY 🎮 #ad, HOY SALE EL KEBAB 🌯 Y EMPIEZA REALMENTE BELLUM 💀 - Bellum #5","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/40a1be7ac247f560b3b4_illojuan_46936407228_1686841854//thumb/thumb2-%{width}x%{height}.jpg","view_count":1472190,"duration_seconds":32540},{"id":"1846151378","user_id":"90075649","stream_id":"46933669100","created_at":"2023-06-14T16:30:20Z","published_at":"2023-06-14T16:30:20Z","language":"es","title":"[𝙆𝙉𝙀𝙆𝙍𝙊 𝘿𝙍𝙊𝙋𝙎] 😱😱 DÍA 4: ENCONTRAREMOS METAL Y PIEDRA?????? 😱😱 - Bellum #4","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/d2nvs31859zcd8/66300cfbf4ed743d8246_illojuan_46933669100_1686760216/thumb/custom-9afff334-64ef-4c59-890f-b9716ff976f3-%{width}x%{height}.jpeg","view_count":1168971,"duration_seconds":21670},{"id":"1845269425","user_id":"90075649","stream_id":"40802865448","created_at":"2023-06-13T15:42:01Z","published_at":"2023-06-13T15:42:01Z","language":"es","title":"[𝘿𝙍𝙊𝙋𝙎 𝘿𝙀 𝘾𝙐𝙍𝙍𝙄𝘾𝙐𝙇𝙐𝙈𝙎] ARMAS DECENTES DESBLOQUEADAS 😈 - Bellum #3","thumbnail_url":"https://static-cdn.jtvnw.net/cf_vods/d2nvs31859zcd8/05fc1b609b42ded343a2_illojuan_40802865448_1686670917/thumb/custom-7f1720b5-afdf-446e-8aa5-c2f222a13b81-%{width}x%{height}.jpeg","view_count":1631286,"duration_seconds":32800}]},"errors":[]}`)

	api := &API{
		db: db,
	}

	app := fiber.New()
	app.Get("/vods", api.Vods)

	params := url.Values{}
	params.Add("bid", "90075649")
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/vods?%s", params.Encode()),
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected http 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	opts := jsondiff.DefaultConsoleOptions()
	if res, diff := jsondiff.Compare(body, wantJson, &opts); res != jsondiff.FullMatch {
		t.Fatal(diff)
	}
}

func TestVodsEmpty(t *testing.T) {
	t.Parallel()
	wantJson := []byte(`{"data":{"vods":[]},"errors":["Missing bid"]}`)

	api := &API{
		db: db,
	}

	app := fiber.New()
	app.Get("/vods", api.Vods)

	req := httptest.NewRequest(
		"GET",
		"/vods",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected http 400, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	opts := jsondiff.DefaultConsoleOptions()
	if res, diff := jsondiff.Compare(body, wantJson, &opts); res != jsondiff.FullMatch {
		t.Fatal(diff)
	}
}

func TestVodsUnknownBID(t *testing.T) {
	t.Parallel()
	wantJson := []byte(`{"data":{"vods":[]},"errors":[]}`)

	api := &API{
		db: db,
	}

	app := fiber.New()
	app.Get("/vods", api.Vods)

	params := url.Values{}
	params.Add("bid", "1234")
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/vods?%s", params.Encode()),
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected http 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	opts := jsondiff.DefaultConsoleOptions()
	if res, diff := jsondiff.Compare(body, wantJson, &opts); res != jsondiff.FullMatch {
		t.Fatal(diff)
	}
}
