package api

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/nsf/jsondiff"
	"pedro.to/rcaptv/helix"
	"pedro.to/rcaptv/test"
)

var db *sql.DB
	var respClipsJson = []byte(`{"data":[{"id":"VainShakingYakinikuAllenHuhu-Qi-dMVyhPUGuJSwX","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:53:07Z","creator_id":"106212238","creator_name":"RagerBomb","title":"JAJAJAJAJAJA WELO ","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/9SO_A0IvnTFcoT1BriUsFw/46977536044-offset-24336-preview-480x272.jpg","duration":26,"view_count":2403,"vod_offset":24312},{"id":"LivelyIronicStaplePeanutButterJellyTime-s8wN4d_iRj_h7JT1","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:45:42Z","creator_id":"186415580","creator_name":"senor_simio","title":"SkainPutereando","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/FtC1-LPMDeHf4kY_4eyU9Q/AT-cm%7CFtC1-LPMDeHf4kY_4eyU9Q-preview-480x272.jpg","duration":13.5,"view_count":1555,"vod_offset":23837},{"id":"HelplessShyTermiteHassaanChop-27ZCgBgsnWMEoKea","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:24:38Z","creator_id":"554249414","creator_name":"0jota0","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/DIsr2tu2QoAYLJ1jBKCnmg/46977536044-offset-1026-preview-480x272.jpg","duration":28,"view_count":1140,"vod_offset":1002},{"id":"BoredTiredFrogSuperVinlin-F7t0wDliOTQgi2zs","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:50:51Z","creator_id":"241257498","creator_name":"AL3xSG_","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/q6zAPRe7dMA5Yjk0IIts7A/46977536044-offset-24200-preview-480x272.jpg","duration":26,"view_count":428,"vod_offset":24176},{"id":"DiligentCrispyMonkeyStoneLightning-6QIyNhVPxXOGBjHX","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:24:56Z","creator_id":"181528808","creator_name":"DNeverland","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3NIsDOkCMjYVT8VzhgYx1w/46977536044-offset-1044-preview-480x272.jpg","duration":28,"view_count":353,"vod_offset":1020},{"id":"FancyArbitraryLemurNomNom-Se5QiZSFesQpBPaE","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T16:29:27Z","creator_id":"164052966","creator_name":"muphasa_","title":"la cama","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/2xhcBUnuUImUhXrEKQT2hA/AT-cm%7C2xhcBUnuUImUhXrEKQT2hA-preview-480x272.jpg","duration":19.7,"view_count":249,"vod_offset":4872},{"id":"OpenPlumpDugongCmonBruh-ty2YkBc_j_o9vkyg","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:33:15Z","creator_id":"453377434","creator_name":"deny_box","title":"Uy","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/_o4M8LctmJAXP4_6Pb5MoA/AT-cm%7C_o4M8LctmJAXP4_6Pb5MoA-preview-480x272.jpg","duration":13.4,"view_count":197,"vod_offset":8711},{"id":"ImpartialTameRadishKeepo-96xOc8uwzM11DHDO","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:58:21Z","creator_id":"480623333","creator_name":"arenass95","title":"TITITITIN","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/hIuWMPesTnPLsMzNdQxwzg/46977536044-offset-3050-preview-480x272.jpg","duration":28,"view_count":161,"vod_offset":3026},{"id":"OnerousDignifiedSageBudBlast-IHCxunhU61blOeFB","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:29:46Z","creator_id":"89501893","creator_name":"Don_Samus","title":":(","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/GfMWQCBwsbp4VOtW5TnhQQ/AT-cm%7CGfMWQCBwsbp4VOtW5TnhQQ-preview-480x272.jpg","duration":38.5,"view_count":143,"vod_offset":22897},{"id":"LazyTrappedYogurtKappaWealth-lVKhImTlCx9p5c5r","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:06:23Z","creator_id":"174732498","creator_name":"Arcaniine","title":"El kenkro esquizo no existe...  Kenkro esquizo:","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/XQaubZVSoSJk5QsfaeXOaw/AT-cm%7CXQaubZVSoSJk5QsfaeXOaw-preview-480x272.jpg","duration":10.8,"view_count":134,"vod_offset":21227},{"id":"TrappedOutstandingWaterDatBoi-rocFCspnvIT2jDcD","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T23:19:07Z","creator_id":"145900332","creator_name":"Leonel0614","title":"Por el orto 2023","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/exRHwq6HBQrKZ7yE9ow3Aw/AT-cm%7CexRHwq6HBQrKZ7yE9ow3Aw-preview-480x272.jpg","duration":5,"view_count":133,"vod_offset":29482},{"id":"ThirstyAgitatedCheddarTF2John-mR6ZhI9FLB7XwRxY","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T23:38:01Z","creator_id":"183681572","creator_name":"ByStrago","title":"YAMETE KUDASAI","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/Vicc1JkemYSPRSwe18EcWA/AT-cm%7CVicc1JkemYSPRSwe18EcWA-preview-480x272.jpg","duration":6.5,"view_count":87,"vod_offset":30618},{"id":"SuspiciousDullAniseArsonNoSexy-A08gHf5hkwFyMNpL","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:21:13Z","creator_id":"97225836","creator_name":"Vsfernandez29","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/un29agNPIB3-YZMBcjDaSw/46977536044-offset-22416-preview-480x272.jpg","duration":30,"view_count":69,"vod_offset":22392},{"id":"SleepyRudeTomatoKappaRoss-tNSP7YtloCPY7Mnq","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T20:44:42Z","creator_id":"101376231","creator_name":"HahaVids","title":"Se pica","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/XlpHvLTUj-2LZaAj2OzJ6g/46977536044-offset-20232-preview-480x272.jpg","duration":26,"view_count":67,"vod_offset":20208},{"id":"WanderingExuberantWrenDatBoi-FpEPi_IgbhrVMR9X","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:30:19Z","creator_id":"145293169","creator_name":"sovietik_","title":"360 MI NIÑO","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3l-2wBaRD4vyIG4a1TEfVw/AT-cm%7C3l-2wBaRD4vyIG4a1TEfVw-preview-480x272.jpg","duration":7.7,"view_count":51,"vod_offset":26543},{"id":"SucculentProudMarrowPanicVis-skGfMRbMxVUCu-xy","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:00:50Z","creator_id":"515013566","creator_name":"D4KN0","title":"asjkkajjkasas","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/HlwPULJBnzEWzMVVArvscg/AT-cm%7CHlwPULJBnzEWzMVVArvscg-preview-480x272.jpg","duration":32.3,"view_count":51,"vod_offset":21169},{"id":"CautiousViscousSrirachaKeyboardCat-QrarxC4wtpzR_Pgs","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:34:05Z","creator_id":"167206569","creator_name":"itsLyrian","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/Nc_hhEK2q8QuMLac0fz4LA/46977536044-offset-23194-preview-480x272.jpg","duration":26,"view_count":41,"vod_offset":23170},{"id":"FaithfulSingleOryxVoteNay-2wx9JR1lRIfP8Urw","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:36:57Z","creator_id":"59179902","creator_name":"Apoken_","title":"y se la pela","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/jL3rr9sMfSUdZUSHz_FK8A/AT-cm%7CjL3rr9sMfSUdZUSHz_FK8A-preview-480x272.jpg","duration":21.4,"view_count":29,"vod_offset":1744},{"id":"LongPiliableSangBuddhaBar-QPmbozZCqqZOjSV5","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:35:18Z","creator_id":"599850528","creator_name":"afri6","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/sydlH0D5m6M2QqqmTo0leQ/46977536044-offset-8866-preview-480x272.jpg","duration":28,"view_count":29,"vod_offset":8842},{"id":"BoredPlacidPicklesBCouch-TK_odDyG7399f-Uq","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:39:19Z","creator_id":"150027413","creator_name":"manuubb","title":"knekro se caga encima","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/1i3YO0yv6HQ8hu9_fX_lKg/AT-cm%7C1i3YO0yv6HQ8hu9_fX_lKg-preview-480x272.jpg","duration":5.2,"view_count":19,"vod_offset":27083},{"id":"AcceptableOddShrimpCharlieBitMe-cbZvmbIDd5LSo5Sg","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:24:48Z","creator_id":"111370378","creator_name":"DTE99","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/kXTxHnG2b5myA3eS0qOUoA/46977536044-offset-8236-preview-480x272.jpg","duration":26,"view_count":17,"vod_offset":8212},{"id":"DifficultHorribleHyenaOneHand-qAvs5TyXzR-I5bgG","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:32:15Z","creator_id":"252090554","creator_name":"xLykkos","title":"u can stop me man","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/iDVt-6gL7OU8l80fiOW4Bg/AT-cm%7CiDVt-6gL7OU8l80fiOW4Bg-preview-480x272.jpg","duration":25.2,"view_count":13,"vod_offset":26646},{"id":"ObservantColdMarjoramNinjaGrumpy-iFhdsmRyzsIbHBI2","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:34:34Z","creator_id":"468966674","creator_name":"arestenshi","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/C-XwB4hYXbJkcs7CGxwSNg/46977536044-offset-8822-preview-480x272.jpg","duration":26,"view_count":13,"vod_offset":8798},{"id":"RockyEnergeticDonkeyCharlieBitMe-DtHuBP6jUOuQCxHs","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T16:52:50Z","creator_id":"196594651","creator_name":"GelatinaLatina","title":"uy va dios mio!!!","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/87iUmA7R7gT-o1DrkAsR7A/AT-cm%7C87iUmA7R7gT-o1DrkAsR7A-preview-480x272.jpg","duration":10.2,"view_count":11,"vod_offset":6311},{"id":"WanderingLittleShrewBleedPurple-iRR1-d8hQw_mabtL","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T20:14:38Z","creator_id":"526433364","creator_name":"Lillos8","title":".","game_id":"21779","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/VQq69YkUdfnfIUMP5mm7Sw/AT-cm%7CVQq69YkUdfnfIUMP5mm7Sw-preview-480x272.jpg","duration":11.7,"view_count":11,"vod_offset":18418},{"id":"FriendlyCourageousBunnySoBayed-trbgdal5ZIN51oT0","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T18:02:47Z","creator_id":"98385514","creator_name":"Neceo","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/FuQIM7a24HWpeWSDedaE9A/46977536044-offset-10516-preview-480x272.jpg","duration":26,"view_count":11,"vod_offset":10492},{"id":"KitschyHonorableQuailPogChamp-faoCKtkhzejEGB8A","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:39:30Z","creator_id":"275210834","creator_name":"rumo9","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/12VUI32qAirZIc3oxBgfAA/46977536044-offset-23518-preview-480x272.jpg","duration":26,"view_count":9,"vod_offset":23494},{"id":"DirtyAntsyRatWutFace-J3Zl0vrAyNT9d8Is","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T23:03:37Z","creator_id":"115348532","creator_name":"petehockxd","title":"ella quiere p*nga","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/pyitvCd2ZdL6wMjU_XfCLA/46977536044-offset-28564-preview-480x272.jpg","duration":28,"view_count":9,"vod_offset":28540},{"id":"WealthyImpossiblePoultryNotATK-56DwZyDToGToLUzM","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:59:47Z","creator_id":"245790334","creator_name":"K0rVuX_x","title":"unai","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/GVuB-Oeg8Gcg06Bk6C1MzA/AT-cm%7CGVuB-Oeg8Gcg06Bk6C1MzA-preview-480x272.jpg","duration":15.5,"view_count":9,"vod_offset":28316},{"id":"PowerfulSmoothRutabagaPipeHype-ZY5pNzNbjazSzzno","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:51:15Z","creator_id":"221050357","creator_name":"kikefiga","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/TlV6o2WOaTczW42s0s856A/46977536044-offset-2624-preview-480x272.jpg","duration":28,"view_count":7,"vod_offset":2600},{"id":"RelentlessHeartlessRaccoonDancingBaby-G5DrpOnCUddr3_4V","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:39:13Z","creator_id":"93340335","creator_name":"elvisjak1","title":"Me representa","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/VneUKRuQ66TpwlHXa9_-FQ/AT-cm%7CVneUKRuQ66TpwlHXa9_-FQ-preview-480x272.jpg","duration":20.2,"view_count":7,"vod_offset":9076},{"id":"BrightPleasantPoxPRChase-cM0fsvSTeHcBkD_C","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:49:00Z","creator_id":"230022270","creator_name":"delnorte89","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/s1o0Z3DNH7L0k3h3AZA_wg/46977536044-offset-2488-preview-480x272.jpg","duration":28,"view_count":7,"vod_offset":2464}],"pagination":{"cursor":""}}`)
var wantClipsJson = []byte(`{"data":{"clips":[{"id":"VainShakingYakinikuAllenHuhu-Qi-dMVyhPUGuJSwX","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:53:07Z","creator_id":"106212238","creator_name":"RagerBomb","title":"JAJAJAJAJAJA WELO ","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/9SO_A0IvnTFcoT1BriUsFw/46977536044-offset-24336-preview-480x272.jpg","duration":26,"view_count":2403,"vod_offset":24312},{"id":"LivelyIronicStaplePeanutButterJellyTime-s8wN4d_iRj_h7JT1","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:45:42Z","creator_id":"186415580","creator_name":"senor_simio","title":"SkainPutereando","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/FtC1-LPMDeHf4kY_4eyU9Q/AT-cm%7CFtC1-LPMDeHf4kY_4eyU9Q-preview-480x272.jpg","duration":13.5,"view_count":1555,"vod_offset":23837},{"id":"HelplessShyTermiteHassaanChop-27ZCgBgsnWMEoKea","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:24:38Z","creator_id":"554249414","creator_name":"0jota0","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/DIsr2tu2QoAYLJ1jBKCnmg/46977536044-offset-1026-preview-480x272.jpg","duration":28,"view_count":1140,"vod_offset":1002},{"id":"BoredTiredFrogSuperVinlin-F7t0wDliOTQgi2zs","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:50:51Z","creator_id":"241257498","creator_name":"AL3xSG_","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/q6zAPRe7dMA5Yjk0IIts7A/46977536044-offset-24200-preview-480x272.jpg","duration":26,"view_count":428,"vod_offset":24176},{"id":"DiligentCrispyMonkeyStoneLightning-6QIyNhVPxXOGBjHX","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:24:56Z","creator_id":"181528808","creator_name":"DNeverland","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3NIsDOkCMjYVT8VzhgYx1w/46977536044-offset-1044-preview-480x272.jpg","duration":28,"view_count":353,"vod_offset":1020},{"id":"FancyArbitraryLemurNomNom-Se5QiZSFesQpBPaE","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T16:29:27Z","creator_id":"164052966","creator_name":"muphasa_","title":"la cama","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/2xhcBUnuUImUhXrEKQT2hA/AT-cm%7C2xhcBUnuUImUhXrEKQT2hA-preview-480x272.jpg","duration":19.7,"view_count":249,"vod_offset":4872},{"id":"OpenPlumpDugongCmonBruh-ty2YkBc_j_o9vkyg","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:33:15Z","creator_id":"453377434","creator_name":"deny_box","title":"Uy","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/_o4M8LctmJAXP4_6Pb5MoA/AT-cm%7C_o4M8LctmJAXP4_6Pb5MoA-preview-480x272.jpg","duration":13.4,"view_count":197,"vod_offset":8711},{"id":"ImpartialTameRadishKeepo-96xOc8uwzM11DHDO","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:58:21Z","creator_id":"480623333","creator_name":"arenass95","title":"TITITITIN","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/hIuWMPesTnPLsMzNdQxwzg/46977536044-offset-3050-preview-480x272.jpg","duration":28,"view_count":161,"vod_offset":3026},{"id":"OnerousDignifiedSageBudBlast-IHCxunhU61blOeFB","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:29:46Z","creator_id":"89501893","creator_name":"Don_Samus","title":":(","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/GfMWQCBwsbp4VOtW5TnhQQ/AT-cm%7CGfMWQCBwsbp4VOtW5TnhQQ-preview-480x272.jpg","duration":38.5,"view_count":143,"vod_offset":22897},{"id":"LazyTrappedYogurtKappaWealth-lVKhImTlCx9p5c5r","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:06:23Z","creator_id":"174732498","creator_name":"Arcaniine","title":"El kenkro esquizo no existe...  Kenkro esquizo:","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/XQaubZVSoSJk5QsfaeXOaw/AT-cm%7CXQaubZVSoSJk5QsfaeXOaw-preview-480x272.jpg","duration":10.8,"view_count":134,"vod_offset":21227},{"id":"TrappedOutstandingWaterDatBoi-rocFCspnvIT2jDcD","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T23:19:07Z","creator_id":"145900332","creator_name":"Leonel0614","title":"Por el orto 2023","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/exRHwq6HBQrKZ7yE9ow3Aw/AT-cm%7CexRHwq6HBQrKZ7yE9ow3Aw-preview-480x272.jpg","duration":5,"view_count":133,"vod_offset":29482},{"id":"ThirstyAgitatedCheddarTF2John-mR6ZhI9FLB7XwRxY","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T23:38:01Z","creator_id":"183681572","creator_name":"ByStrago","title":"YAMETE KUDASAI","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/Vicc1JkemYSPRSwe18EcWA/AT-cm%7CVicc1JkemYSPRSwe18EcWA-preview-480x272.jpg","duration":6.5,"view_count":87,"vod_offset":30618},{"id":"SuspiciousDullAniseArsonNoSexy-A08gHf5hkwFyMNpL","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:21:13Z","creator_id":"97225836","creator_name":"Vsfernandez29","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/un29agNPIB3-YZMBcjDaSw/46977536044-offset-22416-preview-480x272.jpg","duration":30,"view_count":69,"vod_offset":22392},{"id":"SleepyRudeTomatoKappaRoss-tNSP7YtloCPY7Mnq","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T20:44:42Z","creator_id":"101376231","creator_name":"HahaVids","title":"Se pica","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/XlpHvLTUj-2LZaAj2OzJ6g/46977536044-offset-20232-preview-480x272.jpg","duration":26,"view_count":67,"vod_offset":20208},{"id":"WanderingExuberantWrenDatBoi-FpEPi_IgbhrVMR9X","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:30:19Z","creator_id":"145293169","creator_name":"sovietik_","title":"360 MI NIÑO","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/3l-2wBaRD4vyIG4a1TEfVw/AT-cm%7C3l-2wBaRD4vyIG4a1TEfVw-preview-480x272.jpg","duration":7.7,"view_count":51,"vod_offset":26543},{"id":"SucculentProudMarrowPanicVis-skGfMRbMxVUCu-xy","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:00:50Z","creator_id":"515013566","creator_name":"D4KN0","title":"asjkkajjkasas","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/HlwPULJBnzEWzMVVArvscg/AT-cm%7CHlwPULJBnzEWzMVVArvscg-preview-480x272.jpg","duration":32.3,"view_count":51,"vod_offset":21169},{"id":"CautiousViscousSrirachaKeyboardCat-QrarxC4wtpzR_Pgs","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:34:05Z","creator_id":"167206569","creator_name":"itsLyrian","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/Nc_hhEK2q8QuMLac0fz4LA/46977536044-offset-23194-preview-480x272.jpg","duration":26,"view_count":41,"vod_offset":23170},{"id":"FaithfulSingleOryxVoteNay-2wx9JR1lRIfP8Urw","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:36:57Z","creator_id":"59179902","creator_name":"Apoken_","title":"y se la pela","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/jL3rr9sMfSUdZUSHz_FK8A/AT-cm%7CjL3rr9sMfSUdZUSHz_FK8A-preview-480x272.jpg","duration":21.4,"view_count":29,"vod_offset":1744},{"id":"LongPiliableSangBuddhaBar-QPmbozZCqqZOjSV5","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:35:18Z","creator_id":"599850528","creator_name":"afri6","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/sydlH0D5m6M2QqqmTo0leQ/46977536044-offset-8866-preview-480x272.jpg","duration":28,"view_count":29,"vod_offset":8842},{"id":"BoredPlacidPicklesBCouch-TK_odDyG7399f-Uq","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:39:19Z","creator_id":"150027413","creator_name":"manuubb","title":"knekro se caga encima","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/1i3YO0yv6HQ8hu9_fX_lKg/AT-cm%7C1i3YO0yv6HQ8hu9_fX_lKg-preview-480x272.jpg","duration":5.2,"view_count":19,"vod_offset":27083},{"id":"AcceptableOddShrimpCharlieBitMe-cbZvmbIDd5LSo5Sg","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:24:48Z","creator_id":"111370378","creator_name":"DTE99","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/kXTxHnG2b5myA3eS0qOUoA/46977536044-offset-8236-preview-480x272.jpg","duration":26,"view_count":17,"vod_offset":8212},{"id":"DifficultHorribleHyenaOneHand-qAvs5TyXzR-I5bgG","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:32:15Z","creator_id":"252090554","creator_name":"xLykkos","title":"u can stop me man","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/iDVt-6gL7OU8l80fiOW4Bg/AT-cm%7CiDVt-6gL7OU8l80fiOW4Bg-preview-480x272.jpg","duration":25.2,"view_count":13,"vod_offset":26646},{"id":"ObservantColdMarjoramNinjaGrumpy-iFhdsmRyzsIbHBI2","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:34:34Z","creator_id":"468966674","creator_name":"arestenshi","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/C-XwB4hYXbJkcs7CGxwSNg/46977536044-offset-8822-preview-480x272.jpg","duration":26,"view_count":13,"vod_offset":8798},{"id":"RockyEnergeticDonkeyCharlieBitMe-DtHuBP6jUOuQCxHs","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T16:52:50Z","creator_id":"196594651","creator_name":"GelatinaLatina","title":"uy va dios mio!!!","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/87iUmA7R7gT-o1DrkAsR7A/AT-cm%7C87iUmA7R7gT-o1DrkAsR7A-preview-480x272.jpg","duration":10.2,"view_count":11,"vod_offset":6311},{"id":"WanderingLittleShrewBleedPurple-iRR1-d8hQw_mabtL","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T20:14:38Z","creator_id":"526433364","creator_name":"Lillos8","title":".","game_id":"21779","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/VQq69YkUdfnfIUMP5mm7Sw/AT-cm%7CVQq69YkUdfnfIUMP5mm7Sw-preview-480x272.jpg","duration":11.7,"view_count":11,"vod_offset":18418},{"id":"FriendlyCourageousBunnySoBayed-trbgdal5ZIN51oT0","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T18:02:47Z","creator_id":"98385514","creator_name":"Neceo","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/FuQIM7a24HWpeWSDedaE9A/46977536044-offset-10516-preview-480x272.jpg","duration":26,"view_count":11,"vod_offset":10492},{"id":"KitschyHonorableQuailPogChamp-faoCKtkhzejEGB8A","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T21:39:30Z","creator_id":"275210834","creator_name":"rumo9","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/12VUI32qAirZIc3oxBgfAA/46977536044-offset-23518-preview-480x272.jpg","duration":26,"view_count":9,"vod_offset":23494},{"id":"DirtyAntsyRatWutFace-J3Zl0vrAyNT9d8Is","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T23:03:37Z","creator_id":"115348532","creator_name":"petehockxd","title":"ella quiere p*nga","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/pyitvCd2ZdL6wMjU_XfCLA/46977536044-offset-28564-preview-480x272.jpg","duration":28,"view_count":9,"vod_offset":28540},{"id":"WealthyImpossiblePoultryNotATK-56DwZyDToGToLUzM","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T22:59:47Z","creator_id":"245790334","creator_name":"K0rVuX_x","title":"unai","game_id":"491487","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/GVuB-Oeg8Gcg06Bk6C1MzA/AT-cm%7CGVuB-Oeg8Gcg06Bk6C1MzA-preview-480x272.jpg","duration":15.5,"view_count":9,"vod_offset":28316},{"id":"PowerfulSmoothRutabagaPipeHype-ZY5pNzNbjazSzzno","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:51:15Z","creator_id":"221050357","creator_name":"kikefiga","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/TlV6o2WOaTczW42s0s856A/46977536044-offset-2624-preview-480x272.jpg","duration":28,"view_count":7,"vod_offset":2600},{"id":"RelentlessHeartlessRaccoonDancingBaby-G5DrpOnCUddr3_4V","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T17:39:13Z","creator_id":"93340335","creator_name":"elvisjak1","title":"Me representa","game_id":"509658","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/VneUKRuQ66TpwlHXa9_-FQ/AT-cm%7CVneUKRuQ66TpwlHXa9_-FQ-preview-480x272.jpg","duration":20.2,"view_count":7,"vod_offset":9076},{"id":"BrightPleasantPoxPRChase-cM0fsvSTeHcBkD_C","broadcaster_id":"152633332","video_id":"1856303227","created_at":"2023-06-26T15:49:00Z","creator_id":"230022270","creator_name":"delnorte89","title":"ONLY CUAJADO| Empezamos la semana sin tilteos  | !eneba","game_id":"245018539","language":"es","thumbnail_url":"https://clips-media-assets2.twitch.tv/s1o0Z3DNH7L0k3h3AZA_wg/46977536044-offset-2488-preview-480x272.jpg","duration":28,"view_count":7,"vod_offset":2464}]},"errors":[]}`)


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

func TestClips(t *testing.T) {
	t.Parallel()
	bid := "152633332"
	start := "2023-06-26T15:07:30Z"
	end := "2023-06-27T00:46:30Z"

	sv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, r *http.Request) {
		if at := r.URL.Query().Get("started_at"); at != start {
			t.Fatalf("expected started_at %q, got %q", start, at)
		}
		if at := r.URL.Query().Get("ended_at"); at != end {
			t.Fatalf("expected ended_at %q, got %q", end, at)
		}
		if bcId := r.URL.Query().Get("broadcaster_id"); bcId != bid {
			t.Fatalf("expected bid %q, got %q", bid, bcId)
		}
		resp.Write(respClipsJson)
	}))
	hx := helix.NewWithoutExchange(&helix.HelixOpts{
		APIUrl: sv.URL,
	}, sv.Client())
	api := &API{
		db: db,
		hx: hx,
	}

	app := fiber.New()
	app.Get("/clips", api.Clips)

	params := url.Values{}
	params.Add("bid", bid)
	params.Add("started_at", start)
	params.Add("ended_at", end)
	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/clips?%s", params.Encode()),
		nil,
	)
	resp, err := app.Test(req, 10*1000)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(body))
	if resp.StatusCode != 200 {
		t.Fatalf("expected http 200, got %d", resp.StatusCode)
	}


	opts := jsondiff.DefaultConsoleOptions()
	if res, diff := jsondiff.Compare(wantClipsJson, body, &opts); res != jsondiff.FullMatch {
		t.Fatal(diff)
	}
}

func TestClipsEmpty(t *testing.T) {
	t.Parallel()
	wantJson := []byte(`{"data":{"clips":[]},"errors":["Missing bid"]}`)

	api := &API{
		db: db,
	}

	app := fiber.New()
	app.Get("/clips", api.Clips)

	req := httptest.NewRequest(
		"GET",
		"/clips",
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
	if res, diff := jsondiff.Compare(wantJson, body, &opts); res != jsondiff.FullMatch {
		t.Fatal(diff)
	}
}