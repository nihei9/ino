package main

// in Ino:
//     data Pen = BallpointPen string float32
//              | FountainPen

type Pen interface {
	Match() *matcher_Pen
}

type cons_Pen_BallpointPen struct {
	p1 string
	p2 float32
}

func BallpointPen(p1 string, p2 float32) *cons_Pen_BallpointPen {
	return &cons_Pen_BallpointPen{
		p1: p1,
		p2: p2,
	}
}

func (p *cons_Pen_BallpointPen) Match() *matcher_Pen {
	return &matcher_Pen{
		x: p,
	}
}

type cons_Pen_FountainPen struct {
}

func FountainPen() *cons_Pen_FountainPen {
	return &cons_Pen_FountainPen{}
}

func (p *cons_Pen_FountainPen) Match() *matcher_Pen {
	return &matcher_Pen{
		x: p,
	}
}

type matcher_Pen struct {
	x Pen
}

func (m *matcher_Pen) AsBallpointPen() *maybe_BallpointPen {
	if x, ok := m.x.(*cons_Pen_BallpointPen); ok {
		return &maybe_BallpointPen{
			x: x,
		}
	}
	return &maybe_BallpointPen{}
}

func (m *matcher_Pen) AsFountainPen() *maybe_FountainPen {
	if x, ok := m.x.(*cons_Pen_FountainPen); ok {
		return &maybe_FountainPen{
			x: x,
		}
	}
	return &maybe_FountainPen{}
}

type maybe_BallpointPen struct {
	x *cons_Pen_BallpointPen
}

func (p *maybe_BallpointPen) OK() bool {
	return p.x != nil
}

func (p *maybe_BallpointPen) Parameters() (string, float32, bool) {
	if p.OK() {
		return p.x.p1, p.x.p2, true
	}
	return "", 0, false
}

type maybe_FountainPen struct {
	x *cons_Pen_FountainPen
}

func (p *maybe_FountainPen) OK() bool {
	return p.x != nil
}
