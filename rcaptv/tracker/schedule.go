package tracker

import (
	"sync"
	"time"

	"github.com/spaolacci/murmur3"

	"pedro.to/rcaptv/utils"
)

type KeyBalancer interface {
	Key(k string) Minute
}

// CountBalance is a key balancer that simply counts the keys up to a maximum
// value. The load and keys will be 1:1 if the load = number of keys, that is,
// for 200 keys we will have 200 assignations where each key is assigned to a
// single container. If the load > number of keys they will be distributed
// across the key pool in a first-in order.
//
// CountBalance has more effective load distribution, stochastic, does not
// support Remove(). Use this strategy balancer if you only Add() items
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

func StrategyCount(max uint) *CountBalance {
	return &CountBalance{
		max: max,
	}
}

func murmur(k string) uint32 {
	hasher := murmur3.New32()
	hasher.Write(utils.StringToByte(k))
	return hasher.Sum32()
}

// Murmur uses the murmur3 hash to generate a balanced key. Murmur is
// deterministic. Less effective load distribution. Supports Remove()
//
// Note: This was my first approach but it is overall much less effective than
// the deterministic CountBalance when it comes to load distribution. But, while
// the load distribution itself will be stochastic, the MurmurBalance provides
// a deterministic key assignment: a streamer with the same username is
// guaranteed to be assigned to the same key. In our use case this means that
// each streamer request will always be performed in the same minute as long
// as the cycle size is the same.
type MurmurBalance struct {
	max uint32
}

func (b *MurmurBalance) Key(k string) Minute {
	return Minute(murmur(k) % b.max)
}

func StrategyMurmur(max uint32) *MurmurBalance {
	return &MurmurBalance{
		max: max,
	}
}

type Minute uint

const ResetMinute = Minute(0)

type BalancedScheduleOpts struct {
	// After a full cycle, every streamer will have been chosen by Pick()
	CycleSize uint
	// High estimation of the total number of objects to be balanced.
	//
	// For the CountBalancer, If estimation is less than CycleSize, CycleSize
	// will be set to estimation and the streamer load will be distributed 1:1,
	// that is for 200 objects the cycle will take 200 minutes and assign 1 min
	// to each streamer. Consecutively, the cycle will be more frequent than the
	// determined CycleSize and as more objects are added, the cycle will take
	// longer to complete until the CycleSize is reached and the load is
	// balanced.
	EstimatedObjects uint

	// Freq changes scheduler real-time pick interval. Useful for testing. Not
	// recommended for real use cases since minutes is how rate limiting is
	// calculated. BalancedSchedule specifically mentions minutes instead of a
	// generic duration unit for ease of use.
	Freq time.Duration

	// The load balancer strategy.
	//
	// - CountBalance: excelent distribution for any length size, deterministic
	// load distribution, stochastic key assignment. Objects keys are determined
	// by the order they were added
	//
	// - MurmurBalance: good distribution especially in large numbers, stochastic
	// load distribution, deterministic key assignment. Objects are guarantee
	// to have the same min assigned as long as the cycle size is the same
	BalanceStrategy KeyBalancer
}

type RealTimeMinute struct {
	Min     Minute
	Objects []string
}

// Schedule is a guarded schedule map, safe for concurrent access
type Schedule struct {
	mu       sync.Mutex
	schedule map[Minute][]string
}

func (s *Schedule) Add(min Minute, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedule[min] = append(s.schedule[min], key)
}

func (s *Schedule) Remove(min Minute, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedule[min] = utils.RemoveKey(s.schedule[min], key)
}

func (s *Schedule) Pick(min Minute) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.schedule[min]
}

// BalancedSchedule balances the objects in a given cycle size. Objects can
// be hot-added while the scheduler is running.
//
// Start real-time scheduler with bs.Start().
//
// Balance is determined by opts.BalanceStrategy balancer. The default balancer
// is a deterministic count balancer.
type BalancedSchedule struct {
	internal       Schedule
	realTime       chan RealTimeMinute
	cancelRealTime chan struct{}

	opts BalancedScheduleOpts
}

func (bs *BalancedSchedule) Add(key string) {
	min := bs.BalancedMin(key)
	bs.internal.Add(min, key)
}

// Remove removes the key element from the schedule.
//
// IMPORTANT: For this feature to work a deterministic balancer is required.
// Use e.g. StrategyMurmur not StrategyCount. If you only add items you can
// use StrategyCount.
func (bs *BalancedSchedule) Remove(key string) {
	min := bs.BalancedMin(key)
	bs.internal.Remove(min, key)
}

func (bs *BalancedSchedule) Pick(min Minute) []string {
	return bs.internal.Pick(min)
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
// RealTimeMinute object with the minute and the objects corresponding to that
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
			case bs.realTime <- RealTimeMinute{Min: m, Objects: bs.Pick(m)}:
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

func NewBalancedSchedule(opts BalancedScheduleOpts) *BalancedSchedule {
	if opts.CycleSize == 0 {
		// prevent zero division in runtime
		panic("CycleSize must be greater than 0")
	}
	if opts.EstimatedObjects == 0 {
		opts.EstimatedObjects = 100
	}
	if opts.EstimatedObjects < opts.CycleSize {
		opts.CycleSize = opts.EstimatedObjects
	}
	if opts.BalanceStrategy == nil {
		opts.BalanceStrategy = StrategyCount(opts.CycleSize)
	}
	if opts.Freq == 0 {
		opts.Freq = time.Minute
	}

	pre := make(map[Minute][]string, opts.CycleSize)
	// preallocate strings slices
	for min := range pre {
		pre[min] = make([]string, 0, opts.EstimatedObjects/opts.CycleSize)
	}
	bs := &BalancedSchedule{
		opts: opts,
		internal: Schedule{
			schedule: pre,
		},
	}
	bs.realTime = make(chan RealTimeMinute)
	bs.cancelRealTime = make(chan struct{}, 1)

	return bs
}
