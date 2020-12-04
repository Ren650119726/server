package utils

import (
	"errors"
	"fmt"
	"math"
)

/*
	这是一颗满四叉树，初始化层数，没有实现动态添加删除层数功能,暂时只支持正方形地图

	矩形区域的象限划分：		第二层为例 tree->[y][x]node
    UL(1)   |    UR(0)		layerMap[1][0][0] | layerMap[1][0][1]
  ----------|----------		------------------|--------------------
    LL(2)   |    LR(3)		layerMap[1][1][0] | layerMap[1][1][1]
    以下对应LayerID的值
*/
type grids [][]*tnode
type (
	// 四叉树对象
	QuadTree struct {
		root *tnode
		layerMap  []grids
		layerSize []float32
	}

	// 节点对象
	tnode struct {
		layerID string
		lu_pos Vec2f	// 左上坐标
		rd_pos Vec2f	// 右下坐标
		sets []IUnit // 当前节段存储的所有单位
		quadrant  []*tnode	// 0-3 四个象限
	}

	// 节点里最终需要存储的单位
	IUnit interface {
		Usize() float32
		Upos() Vec2f

	}
)

func NewQuadTree(layer int,lup,rdown Vec2f) *QuadTree {
	node := newNode(lup,rdown,"0")
	layerSize := make([]float32,0)
	layerMap := make([]grids,0)
	Size0 := _maxSize(lup,rdown)
	for i:=1;i<=layer;i++{
		count := int(math.Pow(2, float64(i-1)))
		layerSize = append(layerSize, Size0/float32(count))

		_grids := make([][]*tnode,count,count)
		for i:=0;i < len(_grids);i++{
			_grids[i] = make([]*tnode,count,count)
		}
		layerMap = append(layerMap,_grids)
	}
	layerMap[0][0][0] = node
	var split func(cur *tnode,splitcount,c int)
	split = func(cur *tnode,splitcount,c int) {
		if splitcount == 0{
			return
		}
		cur.quadrant = make([]*tnode,4,4)
		center := Div2d(Add2d(cur.lu_pos,cur.rd_pos),2)

		cur.quadrant[0] = newNode(Vec2f{X: center.X, Y: cur.lu_pos.Y},Vec2f{X: cur.rd_pos.X, Y: center.Y},cur.layerID+"0")
		ix,iy := _getMapPos(node,cur.quadrant[0].lu_pos,layerSize[c])
		layerMap[c][iy][ix] = cur.quadrant[0]
		split(cur.quadrant[0],splitcount-1,c+1)

		cur.quadrant[1] = newNode(cur.lu_pos, center,cur.layerID+"1")
		ix,iy = _getMapPos(node,cur.quadrant[1].lu_pos,layerSize[c])
		layerMap[c][iy][ix] = cur.quadrant[1]
		split(cur.quadrant[1],splitcount-1,c+1)

		cur.quadrant[2] = newNode(Vec2f{X: cur.lu_pos.X, Y: center.Y},Vec2f{X: center.X, Y: cur.rd_pos.Y},cur.layerID+"2")
		ix,iy = _getMapPos(node,cur.quadrant[2].lu_pos,layerSize[c])
		layerMap[c][iy][ix] = cur.quadrant[2]
		split(cur.quadrant[2],splitcount-1,c+1)

		cur.quadrant[3] = newNode(center,cur.rd_pos,cur.layerID+"3")
		ix,iy = _getMapPos(node,cur.quadrant[3].lu_pos,layerSize[c])
		layerMap[c][iy][ix] = cur.quadrant[3]
		split(cur.quadrant[3],splitcount-1,c+1)
	}
	split(node,layer-1,1)

	return &QuadTree{
		root:node,
		layerSize:layerSize,
		layerMap:layerMap,
	}
}

func newNode(lup,rdown Vec2f,ID string) *tnode {
	return &tnode{
		lu_pos:   lup,
		rd_pos:   rdown,
		sets:make([]IUnit,0),
		quadrant:nil,
		layerID:ID,
	}
}

func (this *QuadTree)Add(u IUnit) error {
	cur := this.root
	for {
		// 已经是最底层 或者 下一层无法完全包围住单位，就放在当前层
		if layer := len(cur.layerID);layer == len(this.layerSize) || u.Usize() >= this.layerSize[layer+1] {
			cur.sets = append(cur.sets, u)
			return nil
		}else {
			// 算出u属于第几象限
			q := _quadrant(cur.lu_pos,cur.rd_pos,u.Upos())
			if q == -1{
				return errors.New(fmt.Sprintf("error lu:%v rd:%v upos:%v", cur.lu_pos.String(),cur.rd_pos.String(),u.Upos().String()))
			}
			cur = cur.quadrant[q]
		}
	}
}


/*
lu左上坐标
dr右下坐标

传入一个矩形区域，获得附近可能发生碰撞的单位
分为两种情况:
1.如果下一层能完全包围住这个矩形，则只需要获得当前层所在节点，以及周围8个节点所管理的单位，然后进去下一层空间继续查找
2.如果下一层不能完全包围住这个矩形，则无需再进入下一层判断，把当前层所在节点以及周围8个节点的所有单位全部获取即可
*/
func (this *QuadTree)GetNearby(lu,rd Vec2f) []IUnit{
	l := _maxSize(lu,rd)
	center := Div2d(Add2d(lu,rd),2)
	sets := make([]IUnit,0)
	cur := this.root
	for i:= 0;i < len(this.layerSize);i++{
		if i < len(this.layerSize) - 1 && l <= this.layerSize[i+1]{	// 情况1
			// 获得周围8个节点 添加其当前层所有单位
			sets = append(sets,cur.sets...)
			x,y := _getMapPos(this.root,cur.lu_pos,this.layerSize[i])
			aroud := _getAroud(x,y)
			for _,pos := range aroud{
				x := pos[0]
				y := pos[1]
				if x < 0 || x >= len(this.layerMap[i])||
					y < 0 || y >= len(this.layerMap[i]){
					continue
				}
				sets = append(sets,this.layerMap[i][y][x].sets...)
			}
			next := _quadrant(cur.lu_pos,cur.rd_pos,center)
			cur = cur.quadrant[next]
		}else{ // 情况2
			// 获得周围8个节点 添加其当前层所有单位以及子节点所有单位
			sets =  _getSetsOfNode_all(cur)
			x,y := _getMapPos(this.root,cur.lu_pos,this.layerSize[i])
			aroud := _getAroud(x,y)
			for _,pos := range aroud{
				x,y := pos[0],pos[1]
				if x < 0 || x >= len(this.layerMap[i])||
					y < 0 || y >= len(this.layerMap[i]){
					continue
				}
				sets = append(sets,_getSetsOfNode_all(this.layerMap[i][y][x])...)
			}
			break
		}
	}
	return sets
}

// 左上角坐标，转换成数组map坐标
func _getMapPos(node *tnode,v Vec2f,layerSize float32)(x,y int){
	tY := int((node.lu_pos.Y - node.rd_pos.Y)/layerSize)
	x = int(math.Floor(float64(v.X / layerSize)))
	y = tY-int(math.Floor(float64(v.Y / layerSize)))
	return
}

// 以x,y为中心，获得周边8个坐标点[x,y]
func _getAroud(x,y int) [][]int{
	ret := make([][]int,8)
	ret = append(ret,[]int{x+1,y+1})	// 右上
	ret = append(ret,[]int{x,y+1})		// 上
	ret = append(ret,[]int{x-1,y+1})	// 左上
	ret = append(ret,[]int{x-1,y})		// 左
	ret = append(ret,[]int{x-1,y-1})	// 左下
	ret = append(ret,[]int{x,y-1})		// 下
	ret = append(ret,[]int{x+1,y-1})	// 右下
	ret = append(ret,[]int{x+1,y})		// 右
	return ret
}
//
func _maxSize(lup,rdown Vec2f) float32{
	dw := rdown.X - lup.X
	dl := lup.Y - rdown.Y
	if dw > dl{
		return dw
	}else{
		return dl
	}
}
//
func _quadrant(lup,rdown,target Vec2f) int {
	centerPos := Div2d(Add2d(lup,rdown),2)
	if target.X >= centerPos.X && target.Y > centerPos.Y{
		return 0
	}else if target.X < centerPos.X && target.Y >= centerPos.Y{
		return 1
	}else if target.X <= centerPos.X && target.Y < centerPos.Y{
		return 2
	}else if target.X > centerPos.X && target.Y <= centerPos.Y{
		return 3
	}
	return -1
}

// 递归获得当前节点下以及子节点的所有单位
func _getSetsOfNode_all(node *tnode) []IUnit{
	ret := make([]IUnit,0)
	ret = append(ret,node.sets...)
	for _,qnode := range node.quadrant{
		ret = append(ret,_getSetsOfNode_all(qnode)...)
	}
	return ret
}

