package welcomer

import (
	"time"
)

type Timing struct {
	lastTime time.Time
	entries  []TimingEntry
}

type TimingEntry struct {
	Name  string
	Value int64
}

func NewTiming() *Timing {
	st := &Timing{
		lastTime: time.Now(),
		entries:  []TimingEntry{},
	}

	return st
}

func (st *Timing) Track(name string) {
	now := time.Now()

	st.entries = append(st.entries, TimingEntry{
		Name:  name,
		Value: time.Since(st.lastTime).Milliseconds(),
	})

	st.lastTime = now
}

func (st *Timing) String() string {
	res := ""

	for i, entry := range st.entries {
		res += entry.Name + ";dur=" + Itoa(entry.Value)
		if i+1 < len(st.entries) {
			res += ","
		}
	}

	return res
}
