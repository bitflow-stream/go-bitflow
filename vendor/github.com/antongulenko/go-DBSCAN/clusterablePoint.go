package dbscan

import (
	"fmt"
	"sort"
)

type ClusterablePoint interface {
	GetPoint() []float64
	String() string
}

type NamedPoint struct {
	Name  string
	Point []float64
}

func NewNamedPoint(name string, point []float64) *NamedPoint {
	return &NamedPoint{
		Name:  name,
		Point: point,
	}
}

func (self *NamedPoint) String() string {
	return fmt.Sprintf("\"%s\": %v", self.Name, self.Point)
}

func (self *NamedPoint) GetPoint() []float64 {
	return self.Point
}

func (self *NamedPoint) Copy() *NamedPoint {
	var p = new(NamedPoint)
	p.Name = self.Name
	copy(p.Point, self.Point)
	return p
}

// Slice attaches the methods of Interface to []float64, sorting in increasing order.
type ClusterablePointSlice struct {
	Data          []ClusterablePoint
	SortDimension int
}

func (self ClusterablePointSlice) Len() int { return len(self.Data) }
func (self ClusterablePointSlice) Less(i, j int) bool {
	return self.Data[i].GetPoint()[self.SortDimension] < self.Data[j].GetPoint()[self.SortDimension]
}
func (self ClusterablePointSlice) Swap(i, j int) {
	self.Data[i], self.Data[j] = self.Data[j], self.Data[i]
}

// Sort is a convenience method.
func (self ClusterablePointSlice) Sort() { sort.Sort(self) }

func NamedPointToClusterablePoint(in []*NamedPoint) (out []ClusterablePoint) {
	out = make([]ClusterablePoint, len(in))
	for i, v := range in {
		out[i] = v
	}
	return
}
