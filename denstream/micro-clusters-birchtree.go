package denstream

import (
	// log "github.com/sirupsen/logrus"
	"time"
)

var _ ClusterSpace = new(BirchTreeClusterSpace)
var _maxChildren = 3

type BirchTreeClusterSpace struct {
	root          *BirchTreeNode
	nextClusterId int
	totalClusters int
	numDimensions int
}

type BirchTreeNode struct {
	isLeaf      bool
	numChildren int
	cluster     *BasicMicroCluster
	children    []*BirchTreeNode
	parent      *BirchTreeNode
}

func NewBirchTreeClusterSpace(numDimensions int) *BirchTreeClusterSpace {
	res := new(BirchTreeClusterSpace)
	res.Init(numDimensions)
	return res
}

func (s *BirchTreeClusterSpace) Init(numDimensions int) {
	s.nextClusterId = 0
	s.numDimensions = numDimensions
	s.root = new(BirchTreeNode)
	s.root.children = make([]*BirchTreeNode, 3)
	s.root.cluster = &BasicMicroCluster{
		cf1:          make([]float64, numDimensions),
		cf2:          make([]float64, numDimensions),
		creationTime: time.Time{},
		center:       make([]float64, numDimensions),
	}
	s.root.isLeaf = false
	s.root.parent = nil
}

func (s *BirchTreeClusterSpace) NumClusters() int {
	return s.totalClusters
}

func (s *BirchTreeClusterSpace) NearestCluster(point []float64) (nearestCluster MicroCluster) {
	//check root, if root is empty add a child, insert and update root. If not empty, then increment root and all the nodes along the traversal
	var closestDistance float64
	var nearestNode *BirchTreeNode
	curNode := s.root

	// var node BirchTreeNode
	if curNode.numChildren != 0 {
		//do this until you reach a leaf
		for curNode.isLeaf == false {
			for idx := 0; idx < curNode.numChildren; idx++ {
				childClust := curNode.children[idx].cluster
				dist := euclideanDistance(point, childClust.Center()) - childClust.Radius()
				if nearestNode == nil || dist < closestDistance {
					nearestNode = curNode.children[idx]
					closestDistance = dist
				}
			}
			curNode = nearestNode
			nearestNode = nil

		}
		nearestCluster = curNode.cluster
	}

	return
}

func (s *BirchTreeClusterSpace) ClustersDo(do func(cluster MicroCluster)) {

	traverseTree(s.root, do)
}

func (s *BirchTreeClusterSpace) NewCluster(point []float64, creationTime time.Time) MicroCluster {
	clust := &BasicMicroCluster{
		cf1:          make([]float64, s.numDimensions),
		cf2:          make([]float64, s.numDimensions),
		creationTime: creationTime,
		center:       make([]float64, s.numDimensions),
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *BirchTreeClusterSpace) Insert(cluster MicroCluster) {
	//create a new node and embed the cluster
	basic := cluster.(*BasicMicroCluster)
	basic.id = s.nextClusterId
	newNode := &BirchTreeNode{
		true,
		0,
		basic,
		nil,
		nil,
	}
	s.nextClusterId++
	newNode.isLeaf = true

	parentNode := s.root
	if parentNode.numChildren < _maxChildren {
		addCFtoParentNode(parentNode, newNode)
		addChild(parentNode, newNode)
	} else {
		nearestNode := parentNode
		//identify the node where newnode can be insterted as a leaf, if the num of children is max, then split the node
		for nearestNode.isLeaf == false {
			parentNode = nearestNode
			// addCFtoParentNode(parentNode, newNode)
			nearestchildIdx := findNearestChildNode(nearestNode, cluster)
			nearestNode = nearestNode.children[nearestchildIdx]
		}
		if parentNode.numChildren < _maxChildren {
			addChild(parentNode, newNode)
			addCFtoParentNodes(parentNode, newNode)
		} else {
			splitNode(parentNode, newNode)
		}
	}
	s.totalClusters++
	return

}

func (s *BirchTreeClusterSpace) Delete(cluster MicroCluster, reason string) {

	parentNode := s.root
	nearestNode := parentNode
	for nearestNode.isLeaf == false {
		parentNode = nearestNode
		nearestchildIdx := findNearestChildNode(nearestNode, cluster)
		nearestNode = nearestNode.children[nearestchildIdx]
	}

	for i := 0; i < parentNode.numChildren; i++ {
		if parentNode.children[i].cluster.Id() == cluster.Id() {
			delCFfromParentNodes(parentNode, parentNode.children[i])
			parentNode.children[i] = nil
			parentNode.children = append(parentNode.children[:i], parentNode.children[i+1:]...)
			parentNode.numChildren--
			s.totalClusters--
			return
		}
	}
	panic("Cluster not in cluster space during: " + reason)
}

func (s *BirchTreeClusterSpace) TransferCluster(cluster MicroCluster, otherSpace ClusterSpace) {
	s.Delete(cluster, "transfering")
	otherSpace.Insert(cluster)
}

func (s *BirchTreeClusterSpace) UpdateCluster(cluster MicroCluster, do func() (reinsertCluster bool)) {

	parentNode := s.root
	nearestNode := parentNode
	for nearestNode.isLeaf == false {
		parentNode = nearestNode
		nearestchildIdx := findNearestChildNode(nearestNode, cluster)
		nearestNode = nearestNode.children[nearestchildIdx]

	}
	for i := 0; i < parentNode.numChildren; i++ {

		if parentNode.children[i].cluster.Id() == cluster.Id() {
			delCFfromParentNodes(parentNode, parentNode.children[i])
			if do() {
				parentNode.children[i].cluster = cluster.(*BasicMicroCluster)
				parentNode.children[i].cluster.Update()
				addCFtoParentNodes(parentNode, parentNode.children[i])
			}
		}
	}
}

// ========================================================================================================
// ==== Internal ====
// ========================================================================================================

func traverseTree(node *BirchTreeNode, do func(cluster MicroCluster)) {
	if node.isLeaf == true {
		do(node.cluster)
		return
	}
	for _, child := range node.children {
		traverseTree(child, do)
	}
}
func findNearestChildNode(nearestNode *BirchTreeNode, cluster MicroCluster) (nearestChildIdx int) {
	var closestDistance float64
	for idx := 0; idx < nearestNode.numChildren; idx++ {
		childClust := nearestNode.children[idx].cluster
		dist := euclideanDistance(cluster.Center(), childClust.Center())
		if nearestChildIdx == -1 || dist < closestDistance {
			nearestChildIdx = idx
			closestDistance = dist
		}
	}
	return
}
func splitNode(parentNode *BirchTreeNode, newNode *BirchTreeNode) {
	numChildren := parentNode.numChildren
	children := parentNode.children
	numDimensions := len(parentNode.cluster.Center())
	if numChildren == _maxChildren {
		//Identify 2 farthest child as seeds of new node
		farthest := 0.0
		c1 := 0
		c2 := 0
		var dist [3 + 1][3 + 1]float64
		for i := 0; i < numChildren; i++ {
			for j := i + 1; j < numChildren; j++ {
				dist[i][j] = euclideanDistance(children[i].cluster.Center(), children[j].cluster.Center())
				dist[j][i] = dist[i][j]
				if farthest < dist[i][j] {
					c1 = i
					c2 = j
					farthest = dist[i][j]
				}
			}

			dist[i][numChildren] = euclideanDistance(children[i].cluster.Center(), newNode.cluster.Center())
			dist[numChildren][i] = dist[i][numChildren]
			if farthest < dist[i][numChildren] {
				c1 = i
				c2 = numChildren
				farthest = dist[i][numChildren]
			}
		}

		var node1 *BirchTreeNode
		var node2 *BirchTreeNode
		node1 = new(BirchTreeNode)
		node1.children = make([]*BirchTreeNode, _maxChildren)
		node1.cluster = &BasicMicroCluster{
			cf1:          make([]float64, numDimensions),
			cf2:          make([]float64, numDimensions),
			creationTime: time.Time{},
		}
		node2 = new(BirchTreeNode)
		node2.children = make([]*BirchTreeNode, _maxChildren)
		node2.cluster = &BasicMicroCluster{
			cf1:          make([]float64, numDimensions),
			cf2:          make([]float64, numDimensions),
			creationTime: time.Time{},
		}
		for i := 0; i < numChildren; i++ {
			if dist[i][c1] < dist[i][c2] {
				addCFtoParentNode(node1, children[i])
				addChild(node1, children[i])
			} else {
				addCFtoParentNode(node2, children[i])
				addChild(node2, children[i])
			}
		}

		if dist[numChildren][c1] < dist[numChildren][c2] {
			addCFtoParentNode(node1, newNode)
			addChild(node1, newNode)
		} else {
			addCFtoParentNode(node2, newNode)
			addChild(node2, newNode)
		}
		//clear parent
		clearBirchTreeNode(parentNode)
		parentNode.children = make([]*BirchTreeNode, _maxChildren)
		parentNode.cluster = &BasicMicroCluster{
			cf1:          make([]float64, numDimensions),
			cf2:          make([]float64, numDimensions),
			creationTime: time.Time{},
		}
		//update the new sibling nodes to the parent node
		addCFtoParentNode(parentNode, node1)
		addChild(parentNode, node1)

		addCFtoParentNode(parentNode, node2)
		addChild(parentNode, node2)
	}
	return
}

func clearBirchTreeNode(node *BirchTreeNode) {
	node.children = nil
	node.numChildren = 0
	node.cluster.reset()
	node.isLeaf = false
}

func addChild(node *BirchTreeNode, childNode *BirchTreeNode) {
	if node.numChildren < _maxChildren {
		childNode.parent = node
		node.children[node.numChildren] = childNode
		node.numChildren++
	}
}

func addCFtoParentNodes(parentNode *BirchTreeNode, newNode *BirchTreeNode) {
	for {
		curNode := parentNode
		addCFtoParentNode(parentNode, newNode)
		if parentNode.parent == nil {
			break
		}
		parentNode = curNode.parent
	}
}

func addCFtoParentNode(parentNode *BirchTreeNode, newNode *BirchTreeNode) {

	parentClust := parentNode.cluster
	newClust := newNode.cluster

	for i := range newClust.cf1 {
		parentClust.cf1[i] += newClust.cf1[i]
		parentClust.cf2[i] += newClust.cf2[i]
	}
	parentClust.w += newClust.w
	if parentClust.w != 0 {
		parentClust.Update()
	}
}

func delCFfromParentNodes(parentNode *BirchTreeNode, delNode *BirchTreeNode) {
	for {
		curNode := parentNode
		delCFfromParentNode(curNode, *delNode)
		if curNode.parent == nil {
			break
		}
		parentNode = curNode.parent
	}
}
func delCFfromParentNode(parentNode *BirchTreeNode, delNode BirchTreeNode) {
	parentClust := parentNode.cluster
	delClust := delNode.cluster
	for i := range parentClust.cf1 {
		parentClust.cf1[i] -= delClust.cf1[i]
		parentClust.cf2[i] -= delClust.cf2[i]
	}
	parentClust.w -= delClust.w
	if parentClust.w != 0 {
		parentClust.Update()
	}
}
