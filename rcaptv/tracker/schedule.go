package tracker

import (
	"time"

	"github.com/spaolacci/murmur3"
	"pedro.to/rcaptv/utils"
)

type KeyBalancer interface {
	Key(k string) Minute
}

// CountBalance is a key balancer that simply counts the keys up to
// a maximum value. The load and keys will be 1:1, that is for 200 keys
// we will have 200 assignations where each key is assigned to a single
// container
type CountBalance struct {
	max, n uint
}

func (b *CountBalance) Key(k string) Minute {
	// note: int overflow will cause a slight imbalance for n=2^32;n+1 but that's
	// 8000 years of balancer running. If the unit becomes smaller than minutes
	// and the running time a long life span consider using if and max instead of
	// mod
	m := Minute(b.n % b.max)
	b.n++
	return m
}

func murmur(k string) uint32 {
	hasher := murmur3.New32()
	hasher.Write(utils.StringToByte(k))
	return hasher.Sum32()
}

// Murmur uses the murmur3 hash to generate a balanced key.
//
// Note: This was my first approach but it is overall much less effective than
// the deterministic CountBalance when it comes to load distribution. But, while
// the load distribution itself will be stochastic, the MurmurBalance provides
// a deterministic key assignment: a streamer with the same username is
// guaranteed to be assigned to the same key. In our use case this means that
// each streamer requests will be performed always in the same minute as long
// as the cycle size is the same. This could become handy in the future
type MurmurBalance struct {
	max uint32
}

func (b *MurmurBalance) Key(k string) Minute {
	return Minute(murmur(k) % b.max)
}

type Minute uint

const ResetMinute = Minute(0)

type BalancedScheduleOpts struct {
	// After a full cycle, every streamer will have been chosen by Pick()
	CycleSize uint
	// High estimation of the total number of streamers to be balanced.
	//
	// For the CountBalancer, If estimation is less than CycleSize, CycleSize
	// will be set to estimation and the streamer load will be distributed 1:1,
	// that is for 200 streamers the cycle will take 200 minutes and assign 1 min
	// to each streamer. Consecutively, the cycle will be more frequent than the
	// determined CycleSize and as more streamers are added, the cycle will take
	// longer to complete until the CycleSize is reached and the load is
	// balanced.
	EstimatedStreamers uint

	// Freq changes scheduler real-time pick interval. Useful for testing. Not
	// recommended for real use cases since minutes is how rate limiting is
	// calculated. BalancedSchedule specifically mentions minutes instead of a
	// generic duration unit for ease of use.
	Freq time.Duration

	// The load balancer strategy.
	//
	// - CountBalance: excelent distribution for any length size, deterministic
	// load distribution, stochastic key assignment. Streamers keys are determined
	// by the order they were added
	//
	// - MurmurBalance: good distribution especially in large numbers, stochastic
	// load distribution, deterministic key assignment. Streamers are guarantee
	// to have the same min assigned as long as the cycle size is the same
	BalanceStrategy KeyBalancer
}

type RealTimeMinute struct {
	Min       Minute
	Streamers []string
}

// BalancedSchedule balances the streamers in a given cycle size. Streamers can
// be hot-added while the scheduler is running.
//
// Start real-time scheduler with bs.Start().
//
// Balance is determined by opts.BalanceStrategy balancer. The default balancer
// is a deterministic count balancer.
type BalancedSchedule struct {
	schedule       map[Minute][]string
	realTime       chan RealTimeMinute
	cancelRealTime chan struct{}

	opts BalancedScheduleOpts
}

func (bs *BalancedSchedule) Add(streamer string) {
	min := bs.BalancedMin(streamer)
	bs.schedule[min] = append(bs.schedule[min], streamer)
}

func (bs *BalancedSchedule) Pick(min Minute) []string {
	return bs.schedule[min]
}

func (bs *BalancedSchedule) BalancedMin(streamer string) Minute {
	return bs.opts.BalanceStrategy.Key(streamer)
}

func (bs *BalancedSchedule) RealTime() <-chan RealTimeMinute {
	return bs.realTime
}

// Starts real-time scheduler.
//
// Every minute (or bs.opts.Freq), the bs.RealTime() channel will receive a
// RealTimeMinute object with the minute and the streamers corresponding to that
// minute.
//
// The scheduler must be stopped with bs.Cancel()
func (bs *BalancedSchedule) Start() {
	go func() {
		m := ResetMinute
		max := Minute(bs.opts.CycleSize - 1)
		d := bs.opts.Freq

		for {
			if bs == nil {
				return
			}
			select {
			case bs.realTime <- RealTimeMinute{Min: m, Streamers: bs.Pick(m)}:
				time.Sleep(d)
				if m >= max {
					m = ResetMinute
				} else {
					m++
				}
			case <-bs.cancelRealTime:
				return
			}
		}
	}()
}

func (bs *BalancedSchedule) Cancel() {
	bs.cancelRealTime <- struct{}{}
}

func newBalancedSchedule(opts BalancedScheduleOpts) *BalancedSchedule {
	if opts.EstimatedStreamers < opts.CycleSize {
		opts.CycleSize = opts.EstimatedStreamers
	}

	if opts.BalanceStrategy == nil {
		opts.BalanceStrategy = &CountBalance{
			max: opts.CycleSize,
		}
	}

	if opts.Freq == 0 {
		opts.Freq = time.Minute
	}

	bs := &BalancedSchedule{opts: opts}
	bs.schedule = make(map[Minute][]string, opts.CycleSize)
	// preallocate strings slices
	for min := range bs.schedule {
		bs.schedule[min] = make([]string, 0, opts.EstimatedStreamers/opts.CycleSize)
	}
	bs.realTime = make(chan RealTimeMinute)
	bs.cancelRealTime = make(chan struct{}, 1)

	return bs
}
