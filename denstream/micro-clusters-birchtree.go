package denstream

import (
	"time"
)

var _ ClusterSpace = new(BirchTreeClusterSpace)

type BirchTreeClusterSpace struct {
	root          BirchTreeNode
	nextClusterId int
}

type BirchTreeNode struct {
	cluster     *BasicMicroCluster
	numChildren int
	children    []BirchTreeNode
	isLeaf      bool
}

func NewBirchTreeClusterSpace(numDimensions int) *BirchTreeClusterSpace {
	res := new(BirchTreeClusterSpace)
	res.Init()
	return res
}

func (s *BirchTreeClusterSpace) Init() {
	s.nextClusterId = 0
}

func (s *BirchTreeClusterSpace) NumClusters() int {
	//return len(s.clusters)
	return 0
}

func (s *BirchTreeClusterSpace) NearestCluster(point []float64) (nearestCluster MicroCluster) {

	//check root, if root is empty add a child, insert and update root. If not empty, then increment root and all the nodes along the traversal
	var closestDistance float64
	curNode := s.root
	var node BirchTreeNode
	if curNode.numChildren == 0 {
		//create a new leaf node and insert point
		createBirchTreeNode(&node)
		node.isLeaf = true
		node.cluster.id = s.nextClusterId
		s.nextClusterId++
		addChildToNode(&curNode, &node)
		nearestCluster = node.cluster
	} else {
		//while instance of leaf
		for _, child := range curNode.children {
			childClust := child.cluster
			// dist can be negative, if the point is inside a cluster
			dist := euclideanDistance(point, childClust.Center()) - childClust.Radius()
			if nearestCluster == nil || dist < closestDistance {
				nearestCluster = childClust
				closestDistance = dist
			}
		}
	}

	return
}

func (s *BirchTreeClusterSpace) ClustersDo(do func(cluster MicroCluster)) {
	// for cluster := range s.clusters {
	// 	do(cluster)
	// }
}

func (s *BirchTreeClusterSpace) NewCluster(point []float64, creationTime time.Time) MicroCluster {
	clust := &BasicMicroCluster{
		cf1:          make([]float64, len(point)),
		cf2:          make([]float64, len(point)),
		creationTime: creationTime,
	}
	clust.Merge(point)
	s.Insert(clust)
	return clust
}

func (s *BirchTreeClusterSpace) Insert(cluster MicroCluster) {

	basic := cluster.(*BasicMicroCluster)
	basic.id = s.nextClusterId
	s.nextClusterId++

	var newNode BirchTreeNode
	createBirchTreeNode(&newNode)
	newNode.isLeaf = true

	var closestDistance float64
	var nearestCluster MicroCluster
	parentNode := s.root
	if parentNode.numChildren == 0 {
		addCFtoParentNode(&parentNode, &newNode)
		addChildToNode(&parentNode, &newNode)
	} else {
		nearestNode := parentNode
		//identify the node where newnode can be insterted as a leaf, if the num of children is max, then split the node
		for nearestNode.isLeaf == false {
			parentNode := nearestNode
			addCFtoParentNode(&parentNode, &newNode)
			for _, child := range parentNode.children {
				childClust := child.cluster
				//	dist can be negative, if the point is inside a cluster
				dist := euclideanDistance(cluster.Center(), childClust.Center()) - childClust.Radius()
				if nearestCluster == nil || dist < closestDistance {
					nearestCluster = childClust
					closestDistance = dist
				}
				nearestNode = child
			}

		}
		if parentNode.numChildren < 3 {
			addChildToNode(&parentNode, &newNode)
		} else {
			splitNode(&parentNode, newNode)
		}
	}
	return

}

func (s *BirchTreeClusterSpace) Delete(cluster MicroCluster, reason string) {
	//b := cluster.(*BasicMicroCluster)
	//if _, ok := s.clusters[b]; !ok {
	//	panic("Cluster not in cluster space during: " + reason)
	//}
	//delete(s.clusters, b)
}

func (s *BirchTreeClusterSpace) TransferCluster(cluster MicroCluster, otherSpace ClusterSpace) {
	//	s.Delete(cluster, "transfering")
	//	otherSpace.Insert(cluster)
}

func (s *BirchTreeClusterSpace) UpdateCluster(cluster MicroCluster, do func() (reinsertCluster bool)) {
	//basic := cluster.(*BasicMicroCluster)
	//s.Delete(basic, "update")
	//if do() {
	//	s.clusters[basic] = true
	//}
}

// ========================================================================================================
// ==== Internal ====
// ========================================================================================================

func splitNode(parentNode *BirchTreeNode, newNode BirchTreeNode) {
	numChildren := parentNode.numChildren
	children := parentNode.children
	if numChildren == 3 {
		//Identify 2 farthest child as seeds of new node
		farest := 0.0
		c1 := 0
		c2 := 0
		var dist [3 + 1][3 + 1]float64
		for i := 0; i < numChildren; i++ {
			for j := i + 1; j < numChildren; j++ {
				dist[i][j] = euclideanDistance(children[i].cluster.Center(), children[j].cluster.Center())
				dist[j][i] = dist[i][j]
				if farest < dist[i][j] {
					c1 = i
					c2 = j
					farest = dist[i][j]
				}
			}

			dist[i][numChildren] = euclideanDistance(children[i].cluster.Center(), newNode.cluster.Center())
			dist[numChildren][i] = dist[i][numChildren]
			if farest < dist[i][numChildren] {
				c1 = i
				c2 = numChildren
				farest = dist[i][numChildren]
			}
		}

		var node1 BirchTreeNode
		var node2 BirchTreeNode

		createBirchTreeNode(&node1)
		createBirchTreeNode(&node1)

		for i := 0; i < numChildren; i++ {
			if dist[i][c1] < dist[i][c2] {
				addChildToNode(&node1, &children[i])
			} else {
				addChildToNode(&node2, &children[i])
			}
		}
		//clear parent
		clearBirchTreeNode(parentNode)

		//update the new sibling nodes to the parent node
		addCFtoParentNode(parentNode, &node1)
		addChildToNode(parentNode, &node1)
		addCFtoParentNode(parentNode, &node2)
		addChildToNode(parentNode, &node2)
		return
	}
}

func createBirchTreeNode(node *BirchTreeNode) {
	node = &BirchTreeNode{
		&BasicMicroCluster{0, []float64{}, []float64{}, 0, 0, []float64{}, time.Time{}},
		0,
		nil,
		false,
	}
}
func clearBirchTreeNode(node *BirchTreeNode) {
	node.children = nil
	node.numChildren = 0
	node.cluster.reset()
	node.isLeaf = false
}

func addChildToNode(node *BirchTreeNode, childNode *BirchTreeNode) (status bool) {
	if node.numChildren < 3 {
		node.children[node.numChildren] = *childNode
		node.numChildren++
		return true
	}
	return false
}

func delChildFromNode(node *BirchTreeNode, childNode *BirchTreeNode) (status bool) {
	if node.numChildren > 0 {
		node.children[node.numChildren] = *childNode
		node.numChildren--
		return true
	}
	return false
}

func addCFtoParentNode(parentNode *BirchTreeNode, addNode *BirchTreeNode) {
	clust := parentNode.cluster
	addClust := addNode.cluster
	for i := range clust.cf1 {
		clust.cf1[i] += addClust.cf1[i]
		clust.cf2[i] += addClust.cf2[i]
	}
	clust.w += addClust.w
	clust.Update()
}

func delCFfromParentNode(parentNode *BirchTreeNode, delNode BirchTreeNode) {
	clust := parentNode.cluster
	delClust := delNode.cluster
	for i := range clust.cf1 {
		clust.cf1[i] -= delClust.cf1[i]
		clust.cf2[i] -= delClust.cf2[i]
	}
	clust.w -= delClust.w
	clust.Update()
}

func findNearestChildNode(parentNode BirchTreeNode, clust BasicMicroCluster) (childIndex int) {
	var closestDistance float64
	index := -1
	for i, child := range parentNode.children {
		dist := euclideanDistance(clust.Center(), child.cluster.Center())
		if index == -1 || dist < closestDistance {
			index = i
			closestDistance = dist
		}
	}
	childIndex = index
	return
}
