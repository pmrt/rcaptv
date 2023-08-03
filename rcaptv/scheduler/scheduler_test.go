package scheduler

import (
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/montanaflynn/stats"

	"pedro.to/rcaptv/utils"
)

var (
	usernames_fixture string = "silentdreamer12,mountainmaverick,thunderstrider,dragonflydancer,musicalalchemy,adventurerider,starlightwhisper,caffeinecommander,nightfallwarrior,artisticinsomnia,bookishwanderer,magicalmysteries,silverfoxhound,sundrenchedwaves,fashionisto101,wildhearteddreamer,rainbowtrailblazer,nightingalemelody,coffeeloverforever,mountainwanderlust,thunderstormseeker,dragonrider13,mysterymindbender,beachbummerparadise,whimsicalwinds,rockingstar13,moonlitwonder,sunshinewarrior,treasurehunterextra,galaxysurferx,whisperingwillow23,silverliningseeker,pixelpoet23,warriorprincessx,enchantedforever,fairydustwanderer,artisticflairista,gamingaddict999,musicmaverickx,adrenalinejunkie21,mindfulmeditations,sunsetchaser23,wonderlustful13,dancingqueen24x,adventureawaitsme,silentobserver365,natureloverforever,gamechanger999,songbirdmelodies,inspiredbymusicx,storytellerinside,sunflowergurl,whimsicalwanderings,rainbowdreamer365,mindfulmomentsx,coffeeandbookworm,moonlitnightshade13,creativesoul777,dreamweaver2023,magicalifehappens,mountainexplorer365,thunderstruck88x,dragonheartedbeast,wanderingstargazer,beachcomberatheart,daydreambeliever21,whisperingwind23,rocknrollstar999,moonbeammagicx,sunshineandroses23,treasurehunterxtra,cosmicwanderer22x,enchantedwatersfall,fairytalesanddreamsx,artisticjourneyx,gamingwizardx,musicandlyricsrock,adventuroushearted,serenityseeker365,sunsetloverx,wonderlanddreamsx,danceunderthestars24,adventurouspirate,silenthillscreams,naturescapewanderer,gameonforever,songbirdmelodist,inspiretheworldx,mindfulmomentum365,sunflowerbreezex,whimsicaljourney365,rainbowglimmerx,mindfulnessmattersx,coffeecultureaddict,moonlitserenadex,creativityunleashedx,dreamcatcher2023,magicalrealmx,mountainscapeadventurer,thunderstormchaserx,dragonflydreamsx,wanderingmindful,beachvibesatheart,daydreambeliever365,whisperingsecretsx,rockstarinmyveins999,moonlitdreamsx,sunshineadventuresx,treasurehunterextraordinairex,cosmicwanderlustx,enchantedforestwalksx,fairytalesunfoldingx,artisticexpressionx,gamingchampionx,musicislife365,adventuresawaitme,serenereflections365,sunsetserenadex,wonderlandexplorerx,danceandbeheard24,adventurousjourneyx,silentdreamer12,mountainmaverick,thunderstrider,dragonflydancer,musicalalchemy,adventurerider,starlightwhisper,caffeinecommander,nightfallwarrior,artisticinsomnia,bookishwanderer,magicalmysteries,silverfoxhound,sundrenchedwaves,fashionisto101,wildhearteddreamer,rainbowtrailblazer,nightingalemelody,coffeeloverforever,mountainwanderlust,thunderstormseeker,dragonrider13,mysterymindbender,beachbummerparadise,whimsicalwinds,rockingstar13,moonlitwonder,sunshinewarrior,treasurehunterextra,galaxysurferx,whisperingwillow23,silverliningseeker,pixelpoet23,warriorprincessx,enchantedforever,fairydustwanderer,artisticflairista,gamingaddict999,musicmaverickx,adrenalinejunkie21,mindfulmeditations,sunsetchaser23,wonderlustful13,dancingqueen24x,adventureawaitsme,silentobserver365,natureloverforever,gamechanger999,songbirdmelodies,inspiredbymusicx,storytellerinside,sunflowergurl,whimsicalwanderings,rainbowdreamer365,mindfulmomentsx,coffeeandbookworm,moonlitnightshade13,creativesoul777,dreamweaver2023,magicalifehappens,mountainexplorer365,thunderstruck88x,dragonheartedbeast,wanderingstargazer,beachcomberatheart,daydreambeliever21,whisperingwind23,rocknrollstar999,moonbeammagicx,sunshineandroses23,treasurehunterxtra,cosmicwanderer22x,enchantedwatersfall,fairytalesanddreamsx,artisticjourneyx,gamingwizardx,musicandlyricsrock,adventuroushearted,serenityseeker365,sunsetloverx,wonderlanddreamsx,danceunderthestars24,adventurouspirate,silenthillscreams,naturescapewanderer,gameonforever,songbirdmelodist,inspiretheworldx,mindfulmomentum365,sunflowerbreezex,whimsicaljourney365,rainbowglimmerx,mindfulnessmattersx,coffeecultureaddict,moonlitserenadex,creativityunleashedx,dreamcatcher2023,magicalrealmx,mountainscapeadventurer,thunderstormchaserx,dragonflydreamsx,wanderingmindful,beachvibesatheart,daydreambeliever365,whisperingsecretsx,rockstarinmyveins999,moonlitdreamsx,sunshineadventuresx,treasurehunterextraordinairex,cosmicwanderlustx,enchantedforestwalksx,fairytalesunfoldingx,artisticexpressionx,gamingchampionx,musicislife365,adventuresawaitme,serenereflections365,sunsetserenadex,wonderlandexplorerx,danceandbeheard24,adventurousjourneyx,serenejourney365,sunshineandwhiskers,treasurehunterxplorer,cosmicwanderlust365,enchanteddreamerx,fairytalesandmagic,artisticadventurer,gaminglegendx,musicmadness365,adventurespiritxplorer,serenityseeker365x,sunsetwandererx,wonderlanddreamsxplorer,dancingundermoonlight,adventurousheartxplorer,silentdreamweaver,mountainmarauder,thunderstriderx,dragonflydancex,musicalwanderlust,adventureriderx,starlightwhisperx,caffeinecommanderx,nightfallwarriorx,artisticinsomniac,bookishwandererx,magicalmysteriesx,silverfoxhoundx,sundrenchedwavesx,fashionisto101x,wildhearteddreamerx,rainbowtrailblazerx,nightingalemelodyx,coffeeloverforeverx,mountainwanderlustx,thunderstormseekerx,dragonrider13x,mysterymindbenderx,beachbummerparadisex,whimsicalwindsx,rockingstar13x,moonlitwonderx,sunshinewarriorx,treasurehunterextrax,galaxysurferxplorer,whisperingwillow23x,silverliningseekerx,pixelpoet23x,warriorprincessxplorer,enchantedforeverx,fairydustwandererx,artisticflairistax,gamingaddict999x,musicmaverickxplorer,adrenalinejunkie21x,mindfulmeditationsx,sunsetchaser23x,wonderlustful13x,dancingqueen24xplorer,adventureawaitsmex,silentobserver365x,natureloverforeverx,gamechanger999x,songbirdmelodiesx,inspiredbymusicxplorer,storytellerinsidex,sunflowergurlx,whimsicalwanderingsx,rainbowdreamer365x,mindfulmomentsxplorer,coffeeandbookwormx,moonlitnightshade13x,creativesoul777x,dreamweaver2023x,magicalifehappensx,mountainexplorer365x,thunderstruck88xplorer,dragonheartedbeastx,wanderingstargazerx,beachcomberatheartx,daydreambeliever21x,whisperingwind23xplorer,rocknrollstar999x,moonbeammagicxplorer,sunshineandroses23x,treasurehunterxtraordinaire,cosmicwanderer22xplorer,enchantedwatersfallx,fairytalesanddreamsxplorer,artisticjourneyxplorer,gamingwizardxplorer,musicandlyricsrockx,adventurousheartedxplorer,serenityseeker365xplorer,sunsetloverxplorer,wonderlanddreamsxplorer,danceunderthestars24x,adventurouspiratex,silenthillscreamsx,naturescapewandererx,gameonforeverx,songbirdmelodistx,inspiretheworldxplorer,mindfulmomentum365x,sunflowerbreezexplorer,whimsicaljourney365x,rainbowglimmerxplorer,mindfulnessmattersxplorer,coffeecultureaddictx,moonlitserenadexplorer,creativityunleashedxplorer,dreamcatcher2023x,magicalrealmxplorer,mountainscapeadventurerx,thunderstormchaserxplorer,dragonflydreamsxplorer,wanderingmindfulx,beachvibesatheartx,daydreambeliever365x,whisperingsecretsxplorer,rockstarinmyveins999x,moonlitdreamsxplorer,sunshineadventuresxplorer,treasurehunterextraordinairexplorer,cosmicwanderlustxplorer,enchantedforestwalksxplorer,fairytalesunfoldingxplorer,artisticexpressionxplorer,gamingchampionxplorer,musicislife365x,adventuresawaitmex,serenereflections365xplorer,sunsetserenadexplorer,wonderlandexplorerx,danceandbeheard24x,adventurousjourneyxplorer,serenejourney365x,sunshineandwhiskersxplorer,treasurehunterxplorerx,cosmicwanderlust365xplorer,enchanteddreamerxplorer,fairytalesandmagicxplorer,artisticadventurerxplorer,gaminglegendxplorer,musicmadness365xplorer,adventurespiritxplorer,serenityseeker365xplorer,sunsetwandererxplorer,wonderlanddreamsxplorer,dancingundermoonlightx,adventurousheartxplorer,sparksfly,phantomsoul,fizzpop,neonwhisper,glimmeringeyes,mysticspell,velvetshadow,whisperingwind,rhythmicpulse,silentsparkle,azureflame,crimsonblaze,silverstrike,lunarbeam,sunfire,crystaldream,serenadesoul,melodicwhisper,eternalbliss,enchantedaura,shimmeringmist,magicalwhirl,starlightgaze,aurorabreeze,whimsicalwisp,enchantedecho,soothingserenade,illusionarycharm,rainbowglow,harmoniousmelody,whisperingwillow,sparklingtwilight,luminescentgleam,mysticdusk,dreamweaver,goldenglitter,silentwhisper,crystalmoon,sirenssong,cosmicdawn,velvetspell,whisperingillusion,shadowdance,emberflare,stellarwish,goldenflame,crimsonheart,silverdusk,seraphicwhisper,etherealmoon,whisperingglimmer,soothingharmony,illusivewonder,rhythmicwhisper,whisperingembers,serenademagic,whimsicalgaze,enchantedsoul,sparklingdream,celestialwhisper,harmoniousglow,whisperingecho,moonlitmelody,lustrousdusk,dreamcrafter,twilightserenade,luminescentwhisper,mysticillumination,aurorablush,melodicwhirl,enchantedessence,whisperingrainbow,magicalwhisper,starlitbreeze,mysticalwhisper,whispersofeternity,lucidtwilight,crystalwhisper,sirenswhisper,celestialecho,whisperingflame,emberwhisper,stellarwhisper,goldenserenade,crimsonwhisper,silverwhisper,seraphicwhisper,etherealwhisper,whisperingdream,illusivemelody,rhythmicwhisper,whisperingembers,serenademagic,whimsicalwhisper,enchantedsoul,sparklingwhisper,celestialwhisper,harmoniouswhisper,whisperingecho,moonlitwhisper,lustrouswhisper,dreamwhisper,twilightwhisper,luminescentwhisper,mysticwhisper,aurorawhisper,melodicwhisper,enchantedwhisper,whisperingrainbow,magicalwhisper,starlitwhisper,mysticalwhisper,whispersofeternity,lucidwhisper,crystalwhisper,sirenswhisper,celestialecho,whisperingflame,emberwhisper,stellarwhisper,goldenwhisper,crimsonwhisper,silverwhisper,seraphicwhisper,etherealwhisper,whisperingdream,illusivemelody,rhythmicwhisper,whisperingembers,serenademagic,whimsicalwhisper,enchantedsoul,sparklingwhisper,celestialwhisper,harmoniouswhisper,whisperingecho,moonlitwhisper,lustrouswhisper,dreamwhisper,twilightwhisper,luminescentwhisper,mysticwhisper,aurorawhisper,melodicwhisper,enchantedwhisper,whisperingrainbow,magicalwhisper,starlitwhisper,mysticalwhisper,whispersofeternity,lucidwhisper,crystalwhisper,sirenswhisper,celestialecho,whisperingflame,emberwhisper,stellarwhisper,goldenwhisper,crimsonwhisper,silverwhisper,seraphicwhisper,etherealwhisper,whisperingdream,illusivemelody,rhythmicwhisper,whisperingembers,serenademagic,whimsicalwhisper,enchantedsoul,sparklingwhisper,celestialwhisper,harmoniouswhisper,whisperingecho,moonlitwhisper,lustrouswhisper,dreamwhisper,twilightwhisper,luminescentwhisper,mysticwhisper,aurorawhisper,melodicwhisper,enchantedwhisper,whisperingrainbow,magicalwhisper,starlitwhisper,mysticalwhisper,whispersofeternity,lucidwhisper,crystalwhisper,sirenswhisper,celestialecho,whisperingflame,emberwhisper,stellarwhisper,goldenwhisper,crimsonwhisper,silverwhisper,seraphicwhisper,etherealwhisper,whisperingdream,illusivemelody,rhythmicwhisper,whisperingembers,serenademagic,whimsicalwhisper,enchantedsoul,sparklingwhisper,celestialwhisper,harmoniouswhisper,whisperingecho,moonlitwhisper,lustrouswhisper,dreamwhisper,twilightwhisper,luminescentwhisper,mysticwhisper,aurorawhisper,melodicwhisper,enchantedwhisper,whisperingrainbow,magicalwhisper,starlitwhisper,mysticalwhisper,whispersofeternity,lucidwhisper,crystalwhisper,sirenswhisper,celestialecho,whisperingflame,emberwhisper,stellarwhisper,goldenwhisper,crimsonwhisper,silverwhisper,seraphicwhisper,etherealwhisper"
	usernamesWithDups        = strings.Split(usernames_fixture, ",")
	usernames                = deduplicate(usernamesWithDups)
)

func deduplicate(s []string) []string {
	found := make(map[string]bool)
	r := make([]string, 0, len(s))
	for _, item := range s {
		if _, ok := found[item]; !ok {
			r = append(r, item)
			found[item] = true
		}
	}
	return r
}

func TestScheduleAddRemove(t *testing.T) {
	t.Parallel()
	var (
		cycleSize  uint = 10
		estimation uint = uint(len(usernames) + 10)
	)

	bs := New(BalancedScheduleOpts{
		CycleSize:        cycleSize,
		EstimatedObjects: estimation,
		BalanceStrategy:  StrategyMurmur(uint32(cycleSize)),
	})
	for _, username := range usernames {
		bs.Add(username)
	}

	targets := []string{
		"silenthillscreamsx",
		"dragonheartedbeastx",
		"sparksfly",
	}
	min := Minute(3)

	for _, usr := range targets {
		if idx := utils.Find(bs.internal.schedule[min], usr); idx == -1 {
			t.Fatalf("user %s expected and not found", usr)
		}
	}

	t.Logf("before remove:\n%s", spew.Sdump(bs.internal.schedule[min]))
	for _, usr := range targets {
		bs.Remove(usr)
	}
	t.Logf("after remove:\n%s", spew.Sdump(bs.internal.schedule[min]))

	for _, usr := range targets {
		if idx := utils.Find(bs.internal.schedule[min], usr); idx != -1 {
			t.Fatalf("user %s NOT expected and found", usr)
		}
	}
}

func TestBalancedMinDistribution(t *testing.T) {
	t.Parallel()

	var (
		cycleSize  uint = 10
		estimation uint = uint(len(usernames) + 10)
	)

	bs := New(BalancedScheduleOpts{
		CycleSize:        cycleSize,
		EstimatedObjects: estimation,
	})
	for _, username := range usernames {
		bs.Add(username)
	}
	totalLen := len(usernames)
	threshold := 5

	dim := make([]int, 0, cycleSize)
	for _, streamers := range bs.internal.schedule {
		l := len(streamers)
		dim = append(dim, len(streamers))

		idealDistribution := 100 / int(cycleSize)
		currentPercentage := l * 100 / totalLen

		if currentPercentage > idealDistribution+threshold || currentPercentage < idealDistribution-threshold {
			t.Logf("total:%d threshold:%d ", totalLen, threshold)
			t.Logf("current len:%d current percentage:%d, ideal distribution:%d", l, currentPercentage, idealDistribution)
			t.Fatal("unbalanced distribution")
		}
	}

	data := stats.LoadRawData(dim)
	std, err := data.StandardDeviation()
	if err != nil {
		t.Fatal(err)
	}
	if std > 5 {
		t.Log(dim)
		t.Log(bs.internal.schedule)
		t.Logf("std: %f", std)
		t.Fatal("unbalanced distribution")
	}
}

func TestBalancedLowEstimation(t *testing.T) {
	t.Parallel()

	var (
		cycleSize  uint = 1200
		estimation uint = uint(len(usernames))
	)

	bs := New(BalancedScheduleOpts{
		CycleSize:        cycleSize,
		EstimatedObjects: estimation,
	})
	for _, username := range usernames {
		bs.Add(username)
	}

	if bs.opts.CycleSize != estimation {
		t.Fatal("expected cycle size to be equal to estimated streamers when estimation is lower than cycle size")
	}

	dim := make([]int, 0, bs.opts.CycleSize)
	for _, streamers := range bs.internal.schedule {
		l := len(streamers)
		if l != 1 {
			t.Log(dim)
			t.Log(bs.internal.schedule)
			t.Fatal("expected a relation 1:1 (streamer:min) in the cycle")
		}
		dim = append(dim, l)
	}
}

func TestBalancedScheduleRealTime1Pick(t *testing.T) {
	t.Parallel()
	pickInterval := time.Millisecond * 10
	after := time.Millisecond * 15
	r := make([]RealTimeMinute, 0, 5)

	bs := New(BalancedScheduleOpts{
		CycleSize:        10,
		EstimatedObjects: uint(len(usernames) + 10),
		Freq:             pickInterval,
	})
	for _, username := range usernames {
		bs.Add(username)
	}

	timer := time.NewTimer(after)
	bs.Start()
Cycle:
	for {
		select {
		case <-timer.C:
			timer.Stop()
			t.Log("cancel")
			bs.Stop()
			break Cycle
		case m := <-bs.RealTime():
			r = append(r, m)
		}
	}

	if len(r) != 1 {
		t.Fatalf("expected exactly 1 pick, picked: %d", len(r))
	}
	if r[0].Min != 0 {
		t.Fatal("expected scheduler to have picked minute 0")
	}
	if got := len(r[0].Objects); got != 37 {
		t.Logf("total users: %d", len(usernames))
		t.Logf("schedule\n%s", spew.Sdump(bs.internal.schedule))
		t.Logf("picked\n%s", spew.Sdump(r[0]))
		t.Fatalf("expected minute 0 to have 25 streamers, got %d", got)
	}
}

func TestBalancedScheduleRealTimeAfterStart(t *testing.T) {
	t.Parallel()
	pickInterval := time.Millisecond * 25

	bs := New(BalancedScheduleOpts{
		CycleSize:        10,
		EstimatedObjects: 20,
		Freq:             pickInterval,
	})
	bs.Start()
	time.Sleep(time.Second)

	key := "1"
	timeout := time.After(time.Second)
	addTimeout := time.After(100 * time.Millisecond)
	objsPerMin := make([][]string, 0, 20)
Cycle:
	for {
		select {
		case <-timeout:
			bs.Stop()
			break Cycle
		case <-addTimeout:
			bs.Add(key)
		case m := <-bs.RealTime():
			objsPerMin = append(objsPerMin, m.Objects)
		}
	}
	want := key
	notFound := true
	for _, objs := range objsPerMin {
		if len(objs) == 0 {
			continue
		}
		if objs[0] == want {
			notFound = false
		}
	}
	if notFound {
		spew.Dump(objsPerMin)
		t.Fatalf("expected '%s' to be Picked() by scheduler at least once", want)
	}
}

func init() {
	spew.Config = spew.ConfigState{
		SortKeys: true,
		SpewKeys: true,
	}
}
