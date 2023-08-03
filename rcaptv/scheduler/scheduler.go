package scheduler

import (
	"context"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
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

func (m Minute) String() string {
	return strconv.FormatUint(uint64(m), 10)
}

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

	// Salt to be appended to keys.
	Salt string

	// Hook to be run after processing each op. Useful for testing
	AfterOp func(op *Op)
}

type RealTimeMinute struct {
	Min     Minute
	Objects []string
}

type schedulerOp int

const (
	OpAdd schedulerOp = iota
	OpRemove
)

type Op struct {
	Typ schedulerOp
	Key string
	Min Minute
}

type (
	ScheduleMap    map[Minute][]string
	KeyToMinuteMap map[string]Minute
	opsChan        chan *Op
	pickChan       chan []string
)

// Schedule is a schedule map.
//
// IMPORTANT: Schedule is not safe for concurrent access, ensure that only 1
// goroutine access the contents at the same time.
type Schedule struct {
	schedule ScheduleMap
	// denormalize. More memory required, O(1) Add operations
	keyToMin KeyToMinuteMap
	ops      opsChan
	picks    pickChan

	realTime chan RealTimeMinute

	readyCh chan struct{}
	afterOp func(op *Op)
}

func (s *Schedule) add(min Minute, key string) {
	// O(1)
	if _, found := s.keyToMin[key]; !found {
		s.schedule[min] = append(s.schedule[min], key)
		s.keyToMin[key] = min
	}
}

func (s *Schedule) remove(min Minute, key string) {
	// O(n); n = len(s.schedule[min])
	s.schedule[min] = utils.RemoveKey(s.schedule[min], key)
	delete(s.keyToMin, key)
}

func (s *Schedule) pick(min Minute) []string {
	orig := s.schedule[min]
	clone := make([]string, len(orig))
	copy(clone, orig)
	return clone
}

func (s *Schedule) RealTime() <-chan RealTimeMinute {
	return s.realTime
}

// Worker of schedule.
//
// IMPORTANT: only this goroutine should access contents at the same time
func (s *Schedule) Worker(ctx context.Context, freq time.Duration, cycleSize uint) {
	l := log.With().Str("ctx", "scheduler").Logger()
	ticker := time.NewTicker(freq)
	defer ticker.Stop()
	m := ResetMinute
	max := Minute(cycleSize - 1)

	once := make(chan struct{}, 1)
	for {
		select {
		case <-ctx.Done():
			return
		case op := <-s.ops:
			switch op.Typ {
			case OpAdd:
				s.add(op.Min, op.Key)
			case OpRemove:
				s.remove(op.Min, op.Key)
			}
			s.afterOp(op)
		case <-ticker.C:
			select {
			case s.realTime <- RealTimeMinute{Min: m, Objects: s.pick(m)}:
			default:
				l.Warn().Msgf("WARN: discarding minute (min:%d) because 'realTime' channel is blocked.", m)
			}
			if m >= max {
				m = ResetMinute
			} else {
				m++
			}
		case once <- struct{}{}:
			// notify we're ready. readyCh is a channel that no goroutine will ever
			// write, so it is safe to check if it is closed by trying to receive
			// from it
			select {
			case <-s.readyCh:
				// if closed this would never block and therefore never enter default
			default:
				// if not closed <-v.readyCh will block and enter here
				close(s.readyCh)
			}
		}
	}
}

// return a clone of current contents of schedule.
//
// Use only for testing and after stopping scheduler, not safe for concurrent
// access.
func (s *Schedule) UnsafeSchedule() ScheduleMap {
	clone := make(ScheduleMap, len(s.schedule))
	for k, v := range s.schedule {
		clone[k] = v
	}
	return clone
}

// return a clone of current contents of keyToMinute.
//
// Use only for testing and after stopping scheduler, not safe for concurrent
// access.
func (s *Schedule) UnsafeKeyToMinute() KeyToMinuteMap {
	clone := make(KeyToMinuteMap, len(s.keyToMin))
	for k, v := range s.keyToMin {
		clone[k] = v
	}
	return clone
}

type SchedulerCtx struct {
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	stopping chan struct{}
}

// BalancedSchedule balances the objects in a given cycle size. Objects can
// be hot-added while the scheduler is running.
//
// Start real-time scheduler with bs.Start(). Read live minutes with bs.RealTime()
//
// Balance is determined by opts.BalanceStrategy balancer. The default balancer
// is a deterministic count balancer.
//
// BalanceSchedule is safe for concurrent access
type BalancedSchedule struct {
	internal *Schedule
	ctx      *SchedulerCtx
	salt     string

	opts BalancedScheduleOpts
}

func (bs *BalancedSchedule) send(op *Op) {
	// execute in a different goroutine to prevent the current goroutine from
	// blocking. It also holds only the needed values to keep memory footprint as
	// small as possible if blocking go-routines start to pile up. We could have
	// the same effect with a buffered channel, but then we would have to keep an
	// eye on the buffer size
	go func(op *Op) {
		bs.internal.ops <- op
	}(op)
}

// Add adds the key element to the schedule if it is not already in it. Add is
// not a blocking op. Add is safe for concurrent access
func (bs *BalancedSchedule) Add(key string) {
	bs.send(&Op{
		Typ: OpAdd,
		Key: key,
		Min: bs.BalancedMin(key),
	})
}

// Remove removes the key element from the schedule. Remove is not a blocking
// op. Remove is safe for concurrent access.
//
// IMPORTANT: For this feature to work a deterministic balancer is required.
// Use e.g. StrategyMurmur not StrategyCount. If you only add items you can
// use StrategyCount.
func (bs *BalancedSchedule) Remove(key string) {
	bs.send(&Op{
		Typ: OpRemove,
		Key: key,
		Min: bs.BalancedMin(key),
	})
}

func (bs *BalancedSchedule) BalancedMin(key string) Minute {
	if bs.salt != "" {
		return bs.opts.BalanceStrategy.Key(key + bs.salt)
	}
	return bs.opts.BalanceStrategy.Key(key)
}

func (bs *BalancedSchedule) RealTime() <-chan RealTimeMinute {
	return bs.internal.RealTime()
}

func (bs *BalancedSchedule) CycleSize() uint {
	return bs.opts.CycleSize
}

func (bs *BalancedSchedule) EstimatedObjects() uint {
	return bs.opts.EstimatedObjects
}

// return a clone of current contents of schedule.
//
// Use only for testing and after stopping scheduler, not safe for concurrent
// access.
func (bs *BalancedSchedule) UnsafeSchedule() ScheduleMap {
	return bs.internal.UnsafeSchedule()
}

// return a clone of current contents of keyToMinute.
//
// Use only for testing and after stopping scheduler, not safe for concurrent
// access.
func (bs *BalancedSchedule) UnsafeKeyToMinute() KeyToMinuteMap {
	return bs.internal.UnsafeKeyToMinute()
}

// Starts real-time scheduler.
//
// Every minute (or bs.opts.Freq), the bs.RealTime() channel will receive a
// RealTimeMinute object with the minute and the objects corresponding to that
// minute.
//
// The scheduler must be stopped with bs.Stop()
func (bs *BalancedSchedule) Start() {
	bs.resetContext(false)
	ctx := bs.context()
	go func() {
		bs.internal.Worker(ctx, bs.opts.Freq, bs.opts.CycleSize)
		bs.resetContext(true)
		close(bs.ctx.stopping)
	}()
	<-bs.internal.readyCh
}

func (bs *BalancedSchedule) context() context.Context {
	bs.ctx.mu.Lock()
	defer bs.ctx.mu.Unlock()
	return bs.ctx.ctx
}

func (bs *BalancedSchedule) resetContext(empty bool) {
	bs.ctx.mu.Lock()
	defer bs.ctx.mu.Unlock()
	if empty {
		bs.ctx.ctx, bs.ctx.cancel = nil, nil
	} else {
		bs.ctx.ctx, bs.ctx.cancel = context.WithCancel(context.Background())
		bs.ctx.stopping = make(chan struct{})
	}
}

// Stop the scheduler. Stop is idempotent
func (bs *BalancedSchedule) Stop() {
	bs.ctx.mu.Lock()
	if bs.ctx.ctx != nil && bs.ctx.cancel != nil {
		bs.ctx.cancel()
	}
	bs.ctx.mu.Unlock()
	<-bs.ctx.stopping
}

func New(opts BalancedScheduleOpts) *BalancedSchedule {
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

	pre := make(ScheduleMap, opts.CycleSize)
	// preallocate strings slices
	estSize, cycleSize := float64(opts.EstimatedObjects), float64(opts.CycleSize)
	minSize := int64(math.Round(estSize / cycleSize))
	for min := range pre {
		pre[min] = make([]string, 0, minSize)
	}
	bs := &BalancedSchedule{
		opts: opts,
		internal: &Schedule{
			schedule: pre,
			keyToMin: make(KeyToMinuteMap),
			ops:      make(opsChan),
			picks:    make(pickChan),
			realTime: make(chan RealTimeMinute),
			readyCh:  make(chan struct{}),
			afterOp:  func(op *Op) {},
		},
		salt: opts.Salt,
		ctx:  new(SchedulerCtx),
	}

	if opts.AfterOp != nil {
		bs.internal.afterOp = opts.AfterOp
	}
	return bs
}
