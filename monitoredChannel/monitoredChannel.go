package monitoredChannel

import (
	"expvar"
	"fmt"
	"time"
)

type MonitoredChannel[T any] struct {
	name string
	ch   chan T
	vars *MonitoringVariables
}

type MonitoringVariables struct {
	length        *expvar.Int
	capacity      *expvar.Int
	sendsTotal    *expvar.Int
	receivesTotal *expvar.Int
	sendWaitNanos *expvar.Int
}

func NewMonitoredChannel[T any](name string, capacity int, monitoringInterval time.Duration) *MonitoredChannel[T] {
	mc := &MonitoredChannel[T]{
		name: name,
		ch:   make(chan T, capacity),
		vars: &MonitoringVariables{
			length:        expvar.NewInt(fmt.Sprintf("%s_channel_length", name)),
			capacity:      expvar.NewInt(fmt.Sprintf("%s_channel_capacity", name)),
			sendsTotal:    expvar.NewInt(fmt.Sprintf("%s_sends_total", name)),
			receivesTotal: expvar.NewInt(fmt.Sprintf("%s_receives_total", name)),
			sendWaitNanos: expvar.NewInt(fmt.Sprintf("%s_send_wait_ns_total", name)),
		},
	}

	mc.vars.capacity.Set(int64(capacity))

	go mc.monitorLength(monitoringInterval)

	return mc
}

func (mc *MonitoredChannel[T]) monitorLength(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		mc.vars.length.Set(int64(len(mc.ch)))
	}
}

func (mc *MonitoredChannel[T]) Send(item T) {
	startTime := time.Now()

	mc.ch <- item

	waitTime := time.Since(startTime)
	mc.vars.sendWaitNanos.Add(waitTime.Nanoseconds())
	mc.vars.sendsTotal.Add(1)
}

func (mc *MonitoredChannel[T]) Receive() T {
	item := <-mc.ch
	mc.vars.receivesTotal.Add(1)
	return item
}
