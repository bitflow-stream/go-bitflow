package denstream

import (
	"github.com/bitflow-stream/go-bitflow-pipeline/clustering"
	// log "github.com/sirupsen/logrus"
	"time"
)

var _ ClusterSpace = new(BirchTreeClusterSpace)

const _maxChildren = 7

var TotalSparseness float64
var TotalOverlap float64
var subClusterCount int

type BirchTreeClusterSpace struct {
	root          *BirchTreeNode
	nextClusterId int
	totalClusters int
	numDimensions int
}

type BirchTreeNode struct {
	clustering.Coreset

	numChildren int
	children    []*BirchTreeNode
	parent      *BirchTreeNode
	isverified  bool
}

func (n *BirchTreeNode) isLeaf() bool {
	return n.numChildren == 0
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
	s.root.children = make([]*BirchTreeNode, _maxChildren)
	s.root.Coreset = clustering.NewCoreset(numDimensions)
	s.root.parent = nil
}

func (s *BirchTreeClusterSpace) NumClusters() int {
	return s.totalClusters
}

func (s *BirchTreeClusterSpace) TotalWeight() float64 {
	return s.root.Coreset.W()
}

func (s *BirchTreeClusterSpace) NearestCluster(point []float64) (nearestCluster clustering.SphericalCluster) {
	//check root, if root is empty add a child, insert and update root. If not empty, then increment root and all the nodes along the traversal
	var closestDistance float64
	var nearestNode *BirchTreeNode
	curNode := s.root
	if curNode.numChildren != 0 {
		//do this until you reach a leaf
		for !curNode.isLeaf() {
			for idx := 0; idx < curNode.numChildren; idx++ {
				childClust := curNode.children[idx]
				dist := clustering.EuclideanDistance(point, childClust.Center()) - childClust.Radius()
				if nearestNode == nil || dist < closestDistance {
					nearestNode = curNode.children[idx]
					closestDistance = dist
				}
			}
			curNode = nearestNode
			nearestNode = nil
		}
		nearestCluster = curNode
	}

	return
}

func (s *BirchTreeClusterSpace) ClustersDo(do func(cluster clustering.SphericalCluster)) {
	s.traverseTree(s.root, do)
}

func (s *BirchTreeClusterSpace) NewCluster(point []float64, creationTime time.Time) clustering.SphericalCluster {
	clust := &BirchTreeNode{
		Coreset:  clustering.NewCoreset(s.numDimensions),
		children: make([]*BirchTreeNode, _maxChildren),
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *BirchTreeClusterSpace) Insert(cluster clustering.SphericalCluster) {
	//create a new node and embed the cluster
	newChildNode := cluster.(*BirchTreeNode)
	newChildNode.SetId(s.nextClusterId)
	s.nextClusterId++
	if !newChildNode.isLeaf() {
		panic("Can only insert a leaf node without children")
	}

	nearestParentNode := s.root
	if nearestParentNode.numChildren == 0 {
		s.addChild(nearestParentNode, newChildNode)
		s.addCFtoParentNode(nearestParentNode, newChildNode)
	} else {
		nearestNode := nearestParentNode
		//identify the node where newnode can be insterted as a leaf, if the num of children is max, then split the node
		for nearestNode.isLeaf() == false {
			nearestParentNode = nearestNode
			nearestchildIdx := s.findNearestChildNode(nearestNode, newChildNode)
			nearestNode = nearestNode.children[nearestchildIdx]
		}

		if nearestParentNode.numChildren < _maxChildren {
			s.addChild(nearestParentNode, newChildNode)
			s.addCFtoParentNodes(nearestParentNode, newChildNode)
		} else {
			s.addNode(nearestParentNode, newChildNode)
			s.updateCFofParentNodes(newChildNode)
		}
	}
	s.totalClusters++
}

func (s *BirchTreeClusterSpace) Delete(cluster clustering.SphericalCluster, reason string) {

	node := cluster.(*BirchTreeNode)
	parentNode := node.parent

	if node == s.root {
		return
	}
	if parentNode == s.root {
		s.delCFfromParentNode(parentNode, *node)
		s.deleteChild(parentNode, node)
		s.totalClusters--
		return
	}
	if parentNode.numChildren == 1 {
		// Last child -> immediately delete the parent instead
		s.Delete(parentNode, reason)
	} else {
		s.delCFfromParentNodes(parentNode, node)
		s.deleteChild(parentNode, node)
		s.totalClusters--

	}

}

func (s *BirchTreeClusterSpace) TransferCluster(cluster clustering.SphericalCluster, otherSpace ClusterSpace) {
	s.Delete(cluster, "transfering")
	otherSpace.Insert(cluster)
}

func (s *BirchTreeClusterSpace) UpdateCluster(cluster clustering.SphericalCluster, do func() (reinsertCluster bool)) {

	s.Delete(cluster, "Update")
	if do() {
		s.Insert(cluster)
	}
}

func (s *BirchTreeClusterSpace) checkClusterForOpt(epsilon float64) float64 {

	s.traverseSubClusters(s.root, epsilon)

	clustersapartby := (TotalSparseness + TotalOverlap) / float64(2*subClusterCount)

	TotalSparseness = 0.0
	TotalOverlap = 0.0
	subClusterCount = 0
	s.clearParentFlag(s.root)
	return clustersapartby
}

// ========================================================================================================
// ==== Internal ====
// ========================================================================================================

func (s *BirchTreeClusterSpace) createNewNode() *BirchTreeNode {
	node := new(BirchTreeNode)
	node.SetId(s.nextClusterId)
	s.nextClusterId++
	node.children = make([]*BirchTreeNode, _maxChildren)
	node.Coreset = clustering.NewCoreset(s.numDimensions)
	return node
}

func (s *BirchTreeClusterSpace) traverseTree(node *BirchTreeNode, do func(cluster clustering.SphericalCluster)) {
	if node.isLeaf() && node != s.root {
		do(node)
		return
	}
	for i := 0; i < node.numChildren; i++ {
		s.traverseTree(node.children[i], do)
	}
}

func (s *BirchTreeClusterSpace) addNode(node *BirchTreeNode, childNode *BirchTreeNode) {
	if node.numChildren < _maxChildren {
		s.addChild(node, childNode)
	} else {

		parentNode := node.parent
		if parentNode == nil {
			//var newRoot *BirchTreeNode
			parentNode = s.createNewNode()
			s.addChild(parentNode, node)
			s.addCFtoParentNode(parentNode, node)
			s.root = parentNode
		}
		s.addNode(parentNode, s.splitNode(node, childNode))
	}
}

func (s *BirchTreeClusterSpace) splitNode(parentNode *BirchTreeNode, newNode *BirchTreeNode) (newBrother *BirchTreeNode) {

	numChildren := parentNode.numChildren
	children := parentNode.children

	if numChildren == _maxChildren {
		//Identify 2 farthest child as seeds of new node
		farthest := 0.0
		c1 := 0
		c2 := 0
		var dist [_maxChildren + 1][_maxChildren + 1]float64
		for i := 0; i < numChildren; i++ {
			for j := i + 1; j < numChildren; j++ {
				dist[i][j] = clustering.EuclideanDistance(children[i].Center(), children[j].Center())
				dist[j][i] = dist[i][j]
				if farthest < dist[i][j] {
					c1 = i
					c2 = j
					farthest = dist[i][j]
				}
			}

			dist[i][numChildren] = clustering.EuclideanDistance(children[i].Center(), newNode.Center())
			dist[numChildren][i] = dist[i][numChildren]
			if farthest < dist[i][numChildren] {
				c1 = i
				c2 = numChildren
				farthest = dist[i][numChildren]
			}
		}
		s.resetNode(parentNode, s.numDimensions)
		newBrother = s.createNewNode()

		for i := 0; i < numChildren; i++ {
			if dist[i][c1] < dist[i][c2] {
				s.addChild(parentNode, children[i])
			} else {
				s.addChild(newBrother, children[i])

			}
		}

		if dist[numChildren][c1] < dist[numChildren][c2] {
			s.addChild(parentNode, newNode)
		} else {
			s.addChild(newBrother, newNode)
		}

		for i := 0; i < parentNode.numChildren; i++ {
			s.addCFtoParentNode(parentNode, parentNode.children[i])
		}

		for i := 0; i < newBrother.numChildren; i++ {
			s.addCFtoParentNode(newBrother, newBrother.children[i])
		}

	}

	return newBrother
}
func (s *BirchTreeClusterSpace) findNearestChildNode(nearestNode *BirchTreeNode, cluster clustering.SphericalCluster) (nearestChildIdx int) {
	nearestChildIdx = -1

	var closestDistance float64
	for idx := 0; idx < nearestNode.numChildren; idx++ {
		childClust := nearestNode.children[idx]

		dist := clustering.EuclideanDistance(cluster.Center(), childClust.Center()) - childClust.Radius()
		if nearestChildIdx == -1 || dist < closestDistance {
			nearestChildIdx = idx
			closestDistance = dist
		}
	}
	return
}

func (s *BirchTreeClusterSpace) resetNode(node *BirchTreeNode, numDimensions int) {
	node.children = nil
	node.numChildren = 0
	node.children = make([]*BirchTreeNode, _maxChildren)
	node.Coreset = clustering.NewCoreset(numDimensions)
}

func (s *BirchTreeClusterSpace) addChild(node *BirchTreeNode, childNode *BirchTreeNode) {
	if node.numChildren < _maxChildren {
		childNode.parent = node
		node.children[node.numChildren] = childNode
		node.numChildren++
	} else {
		panic("Cannot add child to a node that is already full")
	}
}

func (s *BirchTreeClusterSpace) deleteChild(parentNode *BirchTreeNode, child *BirchTreeNode) {

	for i := 0; i < parentNode.numChildren; i++ {
		if parentNode.children[i] == child {
			parentNode.children[i] = nil
			parentNode.children = append(parentNode.children[:i], parentNode.children[i+1:]...)
			parentNode.children = append(parentNode.children, nil)
			parentNode.numChildren--
		}
	}

}

func (s *BirchTreeClusterSpace) addCFtoParentNodes(parentNode *BirchTreeNode, newNode *BirchTreeNode) {
	for {
		curNode := parentNode
		s.addCFtoParentNode(parentNode, newNode)
		if parentNode.parent == nil {
			break
		}
		parentNode = curNode.parent
	}
}

func (s *BirchTreeClusterSpace) addCFtoParentNode(parentNode *BirchTreeNode, newNode *BirchTreeNode) {
	parentNode.MergeCoreset(&newNode.Coreset)

}

func (s *BirchTreeClusterSpace) delCFfromParentNode(parentNode *BirchTreeNode, delNode BirchTreeNode) {
	parentNode.SubtractCoreset(&delNode.Coreset)
}

func (s *BirchTreeClusterSpace) delCFfromParentNodes(parentNode *BirchTreeNode, delNode *BirchTreeNode) {
	for {
		curNode := parentNode
		s.delCFfromParentNode(curNode, *delNode)
		if curNode.parent == nil {
			break
		}
		parentNode = curNode.parent
	}
}

func (s *BirchTreeClusterSpace) updateCFofParentNodes(node *BirchTreeNode) {

	curNode := node
	for {
		parentNode := curNode.parent
		if parentNode == nil {
			break
		}
		parentNode.Reset()
		for i := 0; i < parentNode.numChildren; i++ {
			s.addCFtoParentNode(parentNode, parentNode.children[i])
		}
		curNode = parentNode
	}
}

/////////////////////////Hyperparameter optimisations///////////////////
func (s *BirchTreeClusterSpace) findCentralCluster(node *BirchTreeNode) (centralCluster *BirchTreeNode) {
	if node.isLeaf() {
		return
	}
	subClustCenter := node.Coreset.Center()
	var closestDistance float64
	// var closestCluster *BirchTreeNode
	for i := 0; i < node.numChildren; i++ {
		childClust := node.children[i]
		dist := clustering.EuclideanDistance(subClustCenter, childClust.Center())
		if closestDistance == float64(0) || dist < closestDistance {
			closestDistance = dist
			centralCluster = childClust
		}
	}
	return
}

func (s *BirchTreeClusterSpace) isCompactSubCluster(node *BirchTreeNode, epsilon float64) {
	centralCluster := s.findCentralCluster(node)
	subClusterCount++
	totalDistApart := 0.0
	countSubclustersApart := 0
	totalDistOverlap := 0.0
	countSubclustersOverlap := 0

	for i := 0; i < node.numChildren; i++ {
		childClust := node.children[i]
		dist := clustering.EuclideanDistance(centralCluster.Center(), childClust.Center())
		if dist == 0 {
			continue
		}

		gap := dist - 2*epsilon
		if gap > 0 {
			totalDistApart += gap
			countSubclustersApart++
		} else {
			totalDistOverlap += gap
			countSubclustersOverlap++
		}
	}
	//avergae sparseness or overlap of each subcluster is summed up
	// to determine if the radius of the cluster has to increase or decrease
	if countSubclustersApart > 0 {
		TotalSparseness += (totalDistApart / (2 * float64(countSubclustersApart)))
	}
	if countSubclustersOverlap > 0 {
		TotalOverlap += (totalDistOverlap / (2 * float64(countSubclustersOverlap)))
	}

}

func (s *BirchTreeClusterSpace) traverseSubClusters(node *BirchTreeNode, epsilon float64) {
	if node.isLeaf() && node.parent.isverified == false {
		s.isCompactSubCluster(node.parent, epsilon)
		node.parent.isverified = true
		return
	}
	for i := 0; i < node.numChildren; i++ {
		s.traverseSubClusters(node.children[i], epsilon)
	}
}

func (s *BirchTreeClusterSpace) clearParentFlag(node *BirchTreeNode) {
	if node.isLeaf() && node.parent.isverified == true {
		node.parent.isverified = false

	}
	for i := 0; i < node.numChildren; i++ {
		s.clearParentFlag(node.children[i])
	}
}
