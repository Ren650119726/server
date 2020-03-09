package logic

var RobotMgr = NewRobotMgr()

type (
	robotMgr struct {
	}
)

func NewRobotMgr() *robotMgr {
	return &robotMgr{}
}

func (self *robotMgr) update() {

}
