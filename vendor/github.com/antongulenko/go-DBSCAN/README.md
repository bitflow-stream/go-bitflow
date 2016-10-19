# DBSCAN (go package) [![GoDoc](https://godoc.org/bitbucket.org/sjbog/go-dbscan?status.svg)](https://godoc.org/bitbucket.org/sjbog/go-dbscan)
DBSCAN (Density-based spatial clustering) clustering optimized for multicore processing.

### Idea
If the distance of two points in any dimension is more than <i>eps</i>, than the total distance is more than <i>eps</i>

1. Calculate variance for each dimension (in parallel), find & store dimension with max variance

2. Sort by the dimension with max variance

3. Build a neighborhood map in parallel (Euclidean distance)

  * Slide through sorted data, from lowest to highest. Sliding performed until neighbor_value <= curr_value + eps

  * Use array / indices to store neighbors lists; ConcurrentLinkedQueue holds density reachable points.

4. Use DFS (Depth First Search) to find clusters

##### Note: Unless go1.5 (or newer), need to set GOMAXPROCS=XX to utilize all cores

#### Example explaining how it works

Consider the following points:
```
Name => { X, Y }
"0" => { 2, 4 }
"1" => { 7, 3 }
"2" => { 3, 5 }
"3" => { 5, 3 }
"4" => { 7, 4 }
"5" => { 6, 8 }
"6" => { 6, 5 }
"7" => { 8, 4 }
"8" => { 2, 5 }
"9" => { 3, 7 }
```

Let eps = 2.0, minPts = 2:
```
  Y|
   |
  8|                       x <- Noise
   |
  7|           x
   |           |
  6|           | <- Points are density connected (from [3,5])
   |           |
  5|       x---x           x
   |         /               \
  4|       x                   x---x
   |                           |
  3|                   x-------x
   |                        /\
  2|                   Points are density reachable
   |
  1|
___|___________________________________
  0|   1   2   3   4   5   6   7   8  X
```

If we'll consider only X dimension, points sorted by this dimension will look like:
```
X  0  1  2  3  4  5  6  7  8  9
P        o  o     o  o  o  o

"0": 2.0, "8": 2.0, "2": 3.0, "9": 3.0, "3": 5.0, "5": 6.0, "6": 6.0, "1": 7.0, "4": 7.0, "7": 8.0
or
"0": [2.0, 4.0], "8": [2.0, 5.0], "2": [3.0, 5.0], "9": [3.0, 7.0], "3": [5.0, 3.0], "5": [6.0, 8.0], "6": [6.0, 5.0], "1": [7.0, 3.0], "4": [7.0, 4.0], "7": [8.0, 4.0]
```

We build neighborhood map (array containing concurrent non-blocking queues of density connected points' indices), by considering forward points, which are reachable by eps:
```
                       0    1    2    3    4    5    6    7    8    9
Sorted data points: [ "0", "8", "2", "9", "3", "5", "6", "1", "4", "7" ]

Point "0": 2.0, sees 3 points in range dimensionValue + eps - "8": 2.0, "2": 3.0, "9": 3.0
(e.g. 2.0 + 2.0 =4 - every point with X dimension value above 4 is definitely not reachable)
Then we check if they are reachable with Euclidean distance and add to neighbors.
Point "9": [3.0, 7.0] isn't density reachable to point "0": [2.0, 4.0]

Neighborhood map:
0: [ 1, 2 ]
1: [ 0 ]
2: [ 0 ]

Point "8": 2.0, sees "2": 3.0, "9": 3.0; but point "9": [3.0, 7.0] isn't density reachable
Neighborhood map:
0: [ 1, 2 ]
1: [ 0, 2 ]
2: [ 0, 1 ]

Point "2": 3.0, sees "9": 3.0, "3": 5.0; but point "3": [5.0, 3.0] isn't density reachable
Neighborhood map:
0: [ 1, 2 ]
1: [ 0, 2 ]
2: [ 0, 1, 3 ]
3: [ 2 ]
```
Note that we slide only forward, thus also add back-references to self index.
Concurrent non-blocking queue allows to add in parallel (order doesn't matter).

Finally neighborhood map:
```
0: [0, 1, 2]
1: [0, 1, 2]
2: [0, 2, 3, 1]
3: [3, 2]
4: [4, 7]
5: [5]
6: [6, 8]
7: [7, 8, 9, 4]
8: [6, 8, 9, 7]
9: [8, 9, 7]
```

Now its easy to traverse and find clusters with Depth First Search (or BFS)
```
Cluster size: 4
 Point: 0, [2.0, 4.0]
 Point: 2, [3.0, 5.0]
 Point: 8, [2.0, 5.0]
 Point: 9, [3.0, 7.0]
Cluster size: 5
 Point: 3, [5.0, 3.0]
 Point: 1, [7.0, 3.0]
 Point: 7, [8.0, 4.0]
 Point: 4, [7.0, 4.0]
 Point: 6, [6.0, 5.0]
```
