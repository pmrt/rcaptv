package test

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"pedro.to/rcaptv/database"
	pg "pedro.to/rcaptv/database/postgres"
)

func SetupPostgres() (*sql.DB, *dockertest.Pool, *dockertest.Resource) {
	// Run a docker with a database for testing
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.3-alpine3.16",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=user",
			"POSTGRES_DB=name",
			"listen_addresses = '*'",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			log.Fatal("could not connect to docker. Is docker service running?")
		}
		panic(err)
	}
	res.Expire(120)

	// Prepare a connection to the db in the docker
	sto := database.New(
		pg.New(&database.StorageOptions{
			StorageHost:            res.GetBoundIP("5432/tcp"),
			StoragePort:            res.GetPort("5432/tcp"),
			StorageUser:            "user",
			StoragePassword:        "test",
			StorageDbName:          "name",
			StorageMaxIdleConns:    5,
			StorageMaxOpenConns:    10,
			StorageConnMaxLifetime: time.Hour,
			StorageConnTimeout:     60 * time.Second,
			DebugMode:              true,

			MigrationVersion: 1,
			MigrationPath:    "../database/postgres/migrations",
		}))
	db := sto.Conn()

	_, err = db.Exec(`
	INSERT INTO tracked_channels (bc_id, bc_display_name, bc_username, bc_type, pp_url, offline_pp_url, tracked_since, seen_inactive_count, enabled_status, last_modified_status, priority_lvl)
	VALUES
			('58753574', 'Zeling', 'zeling', 'partner', 'https://static-cdn.jtvnw.net/jtv_user_pictures/c4b97d39-4a23-4ab8-8d6c-f0fefe8b16f1-profile_image-300x300.png', 'https://static-cdn.jtvnw.net/jtv_user_pictures/1e0e6fd5-841b-4475-aaef-33dc05a5c738-channel_offline_image-1920x1080.jpeg', now(), 0, true, NULL, 0),
			('90075649', 'IlloJuan', 'illojuan', 'partner', 'https://static-cdn.jtvnw.net/jtv_user_pictures/37454f0e-581b-42ba-b95b-416f3113fd37-profile_image-300x300.png', 'https://static-cdn.jtvnw.net/jtv_user_pictures/ce65dafb-f633-4f70-b3a7-7e2eb3006139-channel_offline_image-1920x1080.jpg', now(), 0, true, NULL, 0);

	INSERT INTO vods (video_id, stream_id, bc_id, created_at, published_at, duration_seconds, lang, thumbnail_url, title, view_count)
	VALUES
			('1849520474', '46949794460', '58753574', '2023-06-18T15:31:56Z', '2023-06-18T15:31:56Z', 1870, 'es', 'https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/6a6512cb8facb190d27f_zeling_46949794460_1687102311//thumb/thumb0-%{width}x%{height}.jpg', '🐁 ZELING 🐁 F BELLUM🐁 Ratilla pelirroja 🐁 QUEDAN 20 DIAS DE SEASON  🚀 ( DIAMOND )  LUEGO ONLY UP LA MEJOR PASADOR A DE JUMP KINKGS', 8955),
			('1849313047', '46948572748', '58753574', '2023-06-18T08:10:51Z', '2023-06-18T08:10:51Z', 26420, 'es', 'https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/ca991ee45dbd560dcbeb_zeling_46948572748_1687075846//thumb/thumb0-%{width}x%{height}.jpg', '🐁 ZELING 🐁 F BELLUM🐁 Ratilla pelirroja 🐁 QUEDAN 20 DIAS DE SEASON  🚀 ( DIAMOND )  LUEGO ONLY UP LA MEJOR PASADOR A DE JUMP KINKGS', 51062),
			('1846472757', '46935025356', '58753574', '2023-06-14T23:21:38Z', '2023-06-14T23:21:38Z', 4000, 'es', 'https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/fa9c4ddfe074368f5a9a_zeling_46935025356_1686784894//thumb/thumb0-%{width}x%{height}.jpg', '☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡  453643 HORAS  DE STREAM', 9071),
			('1845909865', '46932511084', '58753574', '2023-06-14T07:21:23Z', '2023-06-14T07:21:23Z', 21350, 'es', 'https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/0adfa532361f7a4e4bf1_zeling_46932511084_1686727279//thumb/thumb0-%{width}x%{height}.jpg', '☣️DROPS☣️BELLUM I TIER 2 👺  RATILLA PELIRROJA VENGATIVA 💩 NOS hemos MUDADO DE BASE 😡', 39782);

	INSERT INTO vods (video_id, stream_id, bc_id, created_at, published_at, duration_seconds, lang, thumbnail_url, title, view_count)
	VALUES
			('1847800606', '46940301884', '90075649', '2023-06-16T15:36:48Z', '2023-06-16T15:36:48Z', 24770, 'es', 'https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/82d5aaf2650410948650_illojuan_46940301884_1686929802//thumb/thumb0-%{width}x%{height}.jpg', '[💀 𝙂𝙊𝙊𝙁𝙔 𝘼𝙎𝙎 𝘿𝙍𝙊𝙋𝙎 💀] DÍA 6: DÍA 1 🌈 - Bellum #6', 970227),
			('1846954069', '46936407228', '90075649', '2023-06-15T15:10:59Z', '2023-06-15T15:10:59Z', 32540, 'es', 'https://static-cdn.jtvnw.net/cf_vods/dgeft87wbj63p/40a1be7ac247f560b3b4_illojuan_46936407228_1686841854//thumb/thumb2-%{width}x%{height}.jpg', '[𝙆𝙀𝘽𝘼𝘽 𝘿𝙍𝙊𝙋𝙎] PROBAMOS LA ROG ALLY 🎮 #ad, HOY SALE EL KEBAB 🌯 Y EMPIEZA REALMENTE BELLUM 💀 - Bellum #5', 1472190),
			('1846151378', '46933669100', '90075649', '2023-06-14T16:30:20Z', '2023-06-14T16:30:20Z', 21670, 'es', 'https://static-cdn.jtvnw.net/cf_vods/d2nvs31859zcd8/66300cfbf4ed743d8246_illojuan_46933669100_1686760216/thumb/custom-9afff334-64ef-4c59-890f-b9716ff976f3-%{width}x%{height}.jpeg', '[𝙆𝙉𝙀𝙆𝙍𝙊 𝘿𝙍𝙊𝙋𝙎] 😱😱 DÍA 4: ENCONTRAREMOS METAL Y PIEDRA?????? 😱😱 - Bellum #4', 1168971),
			('1845269425', '40802865448', '90075649', '2023-06-13T15:42:01Z', '2023-06-13T15:42:01Z', 32800, 'es', 'https://static-cdn.jtvnw.net/cf_vods/d2nvs31859zcd8/05fc1b609b42ded343a2_illojuan_40802865448_1686670917/thumb/custom-7f1720b5-afdf-446e-8aa5-c2f222a13b81-%{width}x%{height}.jpeg', '[𝘿𝙍𝙊𝙋𝙎 𝘿𝙀 𝘾𝙐𝙍𝙍𝙄𝘾𝙐𝙇𝙐𝙈𝙎] ARMAS DECENTES DESBLOQUEADAS 😈 - Bellum #3', 1631286);
	`)
	if err != nil {
		panic(err)
	}
	return db, pool, res
}

func CancelPostgres(pool *dockertest.Pool, res *dockertest.Resource) error {
	return pool.Purge(res)
}
