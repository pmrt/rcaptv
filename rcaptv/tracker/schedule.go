package tracker

import (
	"time"

	"github.com/spaolacci/murmur3"
	"pedro.to/rcaptv/logger"
	"pedro.to/rcaptv/utils"
)

func murmur(k string) uint32 {
	hasher := murmur3.New32()
	hasher.Write(utils.StringToByte(k))
	return hasher.Sum32()
}

type Minute int

const ResetMinute = Minute(0)

type BalancedScheduleOpts struct {
	// After a full cycle, every streamer will have been chosen by Pick()
	CycleSize int
	// High estimation of the total number of streamers to be balanced.
	// Estimation must be greater than CycleSize
	EstimatedStreamers int

	// Freq changes scheduler real-time pick interval. Useful for testing. Not
	// recommended for real use cases since minutes is how rate limiting is
	// calculated. BalancedSchedule specifically mentions minutes instead of a
	// generic duration unit for ease of use.
	Freq time.Duration
}

type RealTimeMinute struct {
	Min       Minute
	Streamers []string
}

// BalancedSchedule balances the streamers in a given cycle size. Streamers can
// be hot-added while the scheduler is running.
//
// Start real-time scheduler with bs.Start().
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
	return Minute(murmur(streamer) % uint32(bs.opts.CycleSize))
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
	bs.realTime = make(chan RealTimeMinute)
	bs.cancelRealTime = make(chan struct{})
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
	l := logger.New("tracker", "scheduler")

	if opts.EstimatedStreamers < opts.CycleSize {
		l.Fatal().Msg("Estimated number of streamers must be greater than the cycle size")
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

	return bs
}
