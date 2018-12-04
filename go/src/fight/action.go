package fight

const (
	Action_move_  = 1
	Action_magic_ = 2
)

type ActionMove struct {
	Radio     int // _all_=1  _half_=2  _quarter_=3
	Direction int /* _top_, _right_, _bottom_, _left_ */
	Loc       Point
}

type ActionMagic struct{}

type ActionEvent struct {
	Token string
	Typ   int // 1->Move  2-> Query
	Ac    interface{}
}
