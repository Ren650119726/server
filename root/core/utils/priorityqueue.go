package utils

/*
	leftchild = n*2
	rightchild = n*2+1
	parent = n/2
*/
type (
	IPriorityInterface interface {
		Priority() int64
	}
	Heap struct {
		arr []IPriorityInterface
		t   int // 1.小顶堆、 2.大顶堆
	}
)

func NewHeap(a []IPriorityInterface, t int) *Heap {
	h := &Heap{arr: a, t: t}
	if h.arr == nil {
		h.arr = make([]IPriorityInterface, 0, 10)
	} else {
		h.heapify()
	}
	return h
}

// 尾部插入一个元素
func (self *Heap) Push(val IPriorityInterface) {
	self.arr = append(self.arr, val)
	_up(self.arr, self.t, len(self.arr))
}

// 查看头部元素
func (self *Heap) Peek() IPriorityInterface {
	if len(self.arr) == 0 {
		return nil
	}
	return self.arr[0]
}

// 头部弹出一个元素
func (self *Heap) Pop() IPriorityInterface {
	n := len(self.arr)
	if n == 0 {
		return nil
	}
	self.arr[n-1], self.arr[0] = self.arr[0], self.arr[n-1]
	ret := self.arr[n-1]
	self.arr = self.arr[:n-1]
	_down(self.arr, self.t, 1)
	return ret
}

// 返回数量
func (self *Heap) Size() int {
	return len(self.arr)
}

/*
改变堆的性质
1.小顶堆
2.大顶堆
*/
func (self *Heap) Change(t int) {
	build := false
	if self.t != t {
		build = true
	}
	self.t = t
	if build {
		self.heapify()
	}
}

// 堆排序
func (self *Heap) Sort() {
	for i := len(self.arr) - 1; i >= 0; i-- {
		self.arr[0], self.arr[i] = self.arr[i], self.arr[0]
		_down(self.arr[:i], self.t, 1)
	}
}

// 建堆
func (self *Heap) heapify() {
	// 默认使用自底向上方式
	_buildHeap_bottom2top(self.arr, self.t)
	//_buildHeap_top2down(self.arr, self.t)
}

/*
向下渗透
arr: 使用的数组
t:   1.小顶堆  2.大顶堆
n:   当前数据的位置，1 表示根 n-1 表示数组下标
*/
func _down(arr []IPriorityInterface, t, n int) {
	l := n * 2
	r := l + 1
	var s int

	lenarr := len(arr)
	if l > lenarr {
		return
	}

	if r > len(arr) {
		if (t == 1 && arr[n-1].Priority() > arr[l-1].Priority()) || (t == 2 && arr[n-1].Priority() < arr[l-1].Priority()) {
			arr[n-1], arr[l-1] = arr[l-1], arr[n-1]
			_down(arr, t, l)
		}
	} else {
		if t == 1 {
			s = min(arr, l-1, r-1) + 1
			if arr[n-1].Priority() > arr[s-1].Priority() {
				arr[n-1], arr[s-1] = arr[s-1], arr[n-1]
				_down(arr, t, s)
			}
		} else if t == 2 {
			s = max(arr, l-1, r-1) + 1
			if arr[n-1].Priority() < arr[s-1].Priority() {
				arr[n-1], arr[s-1] = arr[s-1], arr[n-1]
				_down(arr, t, s)
			}
		}
	}
}

/*
向上排查
arr: 使用的数组
t:   1.小顶堆  2.大顶堆
n:   当前数据的位置，1 表示根 n-1 表示数组下标
*/
func _up(arr []IPriorityInterface, t, n int) {
	if n == 1 {
		return
	}
	p := n / 2

	if (t == 1 && arr[p-1].Priority() > arr[n-1].Priority()) || (t == 2 && arr[p-1].Priority() < arr[n-1].Priority()) {
		arr[p-1], arr[n-1] = arr[n-1], arr[p-1]
		_up(arr, t, p)
	}
}

// 自底向上法建堆
func _buildHeap_bottom2top(arr []IPriorityInterface, t int) {
	// 宏观上是自底向上， 对于每个节点，其实是自顶向下
	n := len(arr)
	for i := n / 2; i > 0; i-- {
		_down(arr, t, i)
	}
}

// 自顶向下法建堆
func _buildHeap_top2down(arr []IPriorityInterface, t int) {
	// 宏观上是自顶向下， 对于每个节点，其实是自底向上
	n := len(arr)
	for i := 1; i <= n; i++ {
		_up(arr, t, i)
	}
}

func min(arr []IPriorityInterface, i1, i2 int) int {
	if arr[i1].Priority() < arr[i2].Priority() {
		return i1
	}
	return i2
}

func max(arr []IPriorityInterface, i1, i2 int) int {
	if arr[i1].Priority() > arr[i2].Priority() {
		return i1
	}
	return i2
}
