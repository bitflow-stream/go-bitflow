package clustering

import (
	"fmt"
	"math"
)

// TODO clean this up or make more usable

const doDebugCoresets = false

type coresetDebugData struct {
	history    []*Coreset
	historyStr []string
	added      [][]float64
	removed    [][]float64
}

func (c *Coreset) debugString() string {
	// if !doDebugCoresets {
	// 	return ""
	// }
	return fmt.Sprintf(",cf1:%v,cf2:%v (cf1/w)^2: %v, cf2/w: %v", c.cf1, c.cf2, VectorSquared(c.cf1, 1/c.w), c.cf2/c.w)
}

func (c *Coreset) storeHistory(message string, params ...interface{}) {
	if !doDebugCoresets {
		return
	}
	c.history = append(c.history, c.clone())
	c.historyStr = append(c.historyStr, fmt.Sprintf(message, params...))
}

func (c *Coreset) debugCreate() {
	c.storeHistory("created")
}

func (c *Coreset) debugClone() {
	c.storeHistory("cloned")
}

func (c *Coreset) debugDecay(decayFactor float64) {
	c.storeHistory("Decayed %v", decayFactor)
}

func (c *Coreset) copyDebugDataFrom(other *Coreset) {
	if !doDebugCoresets {
		return
	}
	c.history = make([]*Coreset, len(other.history))
	c.historyStr = make([]string, len(other.historyStr))
	c.added = make([][]float64, len(other.added))
	c.removed = make([][]float64, len(other.removed))
	copy(c.added, other.added)
	copy(c.removed, other.removed)
	copy(c.historyStr, other.historyStr)
	copy(c.history, other.history)
}

func (c *Coreset) debugCopyFrom(other *Coreset) {
	if !doDebugCoresets {
		return
	}
	c.copyDebugDataFrom(other)
	c.storeHistory("copied from %v", other)
}

func (c *Coreset) debugMerge(p []float64) {
	if !doDebugCoresets {
		return
	}
	c.added = append(c.added, _copyPoint(p))
	c.storeHistory("Merged point %v", p)
}

func (c *Coreset) debugSubtract(p []float64) {
	if !doDebugCoresets {
		return
	}
	c.removed = append(c.removed, _copyPoint(p))
	c.storeHistory("Subtracted Point %v", p)
}

func (c *Coreset) debugMergeCoreset(other *Coreset) {
	if !doDebugCoresets {
		return
	}
	c.added = append(c.added, other.added...)
	c.removed = append(c.removed, other.removed...)
	c.storeHistory("Merged coreset: %v", other)
}

func (c *Coreset) debugSubtractCoreset(other *Coreset) {
	if !doDebugCoresets {
		return
	}
	c.added = append(c.added, other.removed...)
	c.removed = append(c.removed, other.added...)
	c.storeHistory("Subtracted coreset: %v", other)
}

func (c *Coreset) silentMerge(p []float64) {
	for i, v := range p {
		c.cf1[i] += v
	}
	c.cf2 += VectorSquared(p, 1)
	c.w++
	c.added = append(c.added, _copyPoint(p))
}

func (c *Coreset) silentSubtractPoint(p []float64) {
	for i, v := range p {
		c.cf1[i] -= v
	}
	c.cf2 -= VectorSquared(p, 1)
	c.w--
	c.removed = append(c.removed, _copyPoint(p))
}

func (c *Coreset) validate() {
	dimensions := -1
	if len(c.added) > 0 {
		dimensions = len(c.added[0])
	}
	if len(c.removed) > 0 {
		dimensions = len(c.removed[0])
	}
	if dimensions == -1 {
		return
	}

	cf1 := make([]float64, dimensions)
	cf2 := 0.0
	w := float64(len(c.added) - len(c.removed))
	for _, add := range c.added {
		for i, v := range add {
			cf1[i] += v
		}
		cf2 += VectorSquared(add, 1)
	}
	for _, rem := range c.removed {
		for i, v := range rem {
			cf1[i] -= v
		}
		cf2 -= VectorSquared(rem, 1)
	}

	small := 1e-10
	isClose := func(a, b float64) bool {
		return math.Abs(a-b) < small
	}

	ok := isClose(w, c.w)
	if ok {
		ok = isClose(cf2, c.cf2)
	}
	if ok {
		for i, c := range c.cf1 {
			if !isClose(c, cf1[i]) {
				ok = false
				break
			}
		}
	}
	if !ok {
		c.doPanic(fmt.Sprintf("Invalid cluster (added %v, removed %v) (w: %v vs %v, cf2: %v vs %v, cf1: %v vs %v)",
			len(c.added), len(c.removed), c.w, w, c.cf2, cf2, c.cf1, cf1))
	}
}

func (c *Coreset) debugWrongRadiusComputation() {
	if !doDebugCoresets {
		return
	}
	c.validate()

	x := NewCoreset(len(c.added[0]))
	for _, add := range c.added {
		x.silentMerge(add)
	}
	for _, p := range c.removed {
		x.silentSubtractPoint(p)
	}
	fmt.Printf("Added: %#v", x.added)
	fmt.Printf("Removed: %#v", x.removed)
	x.validate()

	cf2 := x.cf2 / x.w
	cf1 := VectorSquared(x.cf1, 1/x.w)
	fmt.Printf("THE NEW cf2: %v cf1: %v, w: %v\n", cf2, cf1, x.w)
}

func (c *Coreset) doPanic(header string) {
	str := ""
	for i, histStr := range c.historyStr {
		str += fmt.Sprintf("\n\n %v: %v \n %v", i, histStr, c.history[i])
	}
	fmt.Println(header)
	fmt.Println(str)
	panic(header)
}

func _copyPoint(f []float64) []float64 {
	res := make([]float64, len(f))
	copy(res, f)
	return res
}
