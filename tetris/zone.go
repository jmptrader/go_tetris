// mainZone of game
// a list of lines
package tetris

var (
	constClearLine []Color
)

type zone struct {
	h, w     int
	data     [][]Color
	wrapZone [][]Color
}

func newZone(height, width int) *zone {
	constClearLine = make([]Color, width)
	for i := range constClearLine {
		constClearLine[i] = constColorNothing
	}
	z := make([][]Color, height)
	for i := range z {
		z[i] = make([]Color, width)
	}
	w := make([][]Color, height)
	for i := range w {
		w[i] = make([]Color, width)
	}
	return &zone{
		h:        height,
		w:        width,
		data:     z,
		wrapZone: w,
	}
}

func (z zone) width() int                    { return z.w }
func (z zone) height() int                   { return z.h }
func (z zone) getDotByCoor(y, x int) Color   { return z.data[y][x] }
func (z zone) getLineByHeight(y int) []Color { return z.data[y] }
func (z *zone) setDot(y, x int, val Color)   { z.data[y][x] = val }
func (z *zone) setLine(y int, val []Color) {
	for x, v := range val {
		z.data[y][x] = v
	}
}

// check whether the line is a stone line
func (z zone) isStoneLine(y int) bool {
	for _, c := range z.getLineByHeight(y) {
		if c.isStone() || c.isBomb() {
			return true
		}
	}
	return false
}

// check whether the line contains nothing
func (z zone) isNothingLine(y int) bool {
	for _, c := range z.getLineByHeight(y) {
		if !c.isNothing() {
			return false
		}
	}
	return true
}

// get the x-coordinate of the bomb on a stone line
func (z zone) getBombXCoor(y int) (x int) {
	x = -1
	for i, c := range z.getLineByHeight(y) {
		if c.isBomb() {
			x = i
			break
		}
	}
	return
}

// the function should be called after clearLinesByIndex
// check whether the game zone is clear
func (z zone) isZoneClear() bool {
	for y := 0; y < z.height(); y++ {
		for x := 0; x < z.width(); x++ {
			if !z.getDotByCoor(y, x).isNothing() {
				return false
			}
		}
	}
	return true
}

// the function should be called after setting the active block on zone
// calculate which lines should be cleared
// return the index of lines, num of lines to clear, num of bombs hit
func (z zone) calculateLinesToClear(b block) ([]int, int, int) {
	ltc := make([]int, 0)
	lines := 0
	bombs := 0
loopY:
	for y := 0; y < z.height(); y++ {
		if z.isNothingLine(y) {
			continue
		}
		// if the line is a stone line
		// check whether the block hits the bomb
		// if hit, check how many bombs it hits
		// after all checking, the loop should be exit
		// because stone line is the last to check
		if z.isStoneLine(y) {
			xOfBomb := z.getBombXCoor(y)
			for _, d := range b {
				if d.y != y-1 {
					continue
				}
				if d.x == xOfBomb {
					bombs++
					ltc = append(ltc, y)
					for y = y + 1; y < z.height(); y++ {
						xOfBomb = z.getBombXCoor(y)
						if xOfBomb != d.x {
							break loopY
						}
						bombs++
						ltc = append(ltc, y)
					}
				}
			}
			break loopY
		}
		// if the line only contains active color
		// it should be clear
		for x := 0; x < z.width(); x++ {
			c := z.getDotByCoor(y, x)
			if !c.isActiveColor() {
				continue loopY
			}
		}
		ltc = append(ltc, y)
		lines++
	}
	return ltc, lines, bombs
}

// the indice should be sorted asc
func (z *zone) clearLinesByIndex(indice []int) {
	for _, y := range indice {
		if y > 0 {
			for i := y; i > 0; i-- {
				z.setLine(i, z.getLineByHeight(i-1))
			}
		}
		z.setLine(0, constClearLine)
	}
}

// remove the stone lines after being ko by the opponent
func (z *zone) removeStoneLines() {
	indice := make([]int, 0)
	for y := 0; y < z.height(); y++ {
		if z.isStoneLine(y) {
			indice = append(indice, y)
		}
	}
	z.clearLinesByIndex(indice)
}

// the function should be called after canHoldStoneLines
// add n stone lines to the zone
func (z *zone) addStoneLinesToZone(n int) {
	var l = z.height()
	for n > 0 {
		n--
		var stoneLine = make([]Color, z.width())
		xOfBomb := randSeed.Intn(z.width())
		for i := 0; i < z.width(); i++ {
			if i == xOfBomb {
				stoneLine[i] = constColorBomb
			} else {
				stoneLine[i] = constColorStone
			}
		}
		for i := 0; i < l-1; i++ {
			z.setLine(i, z.getLineByHeight(i+1))
		}
		z.setLine(l-1, stoneLine)
	}
}

// check if the zone can hold num stone lines or not
func (z zone) canHoldStoneLines(num int) bool {
	n := 0
	for y := 0; y < z.height(); y++ {
		if z.isNothingLine(y) {
			n++
			if n >= num {
				return true
			}
		}
	}
	return false
}

// check if being ko
func (z zone) beingKO() bool {
	return !z.canHoldStoneLines(1)
}

// the function should be called after canPutBlockOnZone
// put the block on zone
func (z *zone) putBlockOnZone(b block) {
	for _, d := range b {
		if z.getDotByCoor(d.y, d.x).isNothing() {
			z.setDot(d.y, d.x, d.Color)
		}
	}
}

// drop a block on main zone, return the last location of the block
func (z zone) dropBlockOnZone(b block) block {
	for z.canBlockMoveDown(b) {
		b = b.moveDown()
	}
	return b
}

// remember to call this function after being attacked
// to check if the block can be put on the zone or not
//
// check if the block can be put on the zone
func (z zone) canPutBlockOnZone(b block) bool {
	for _, d := range b {
		if !z.getDotByCoor(d.y, d.x).isNothing() {
			return false
		}
	}
	return true
}

// check if the block can move down
func (z zone) canBlockMoveDown(b block) bool {
	for _, d := range b {
		if d.y >= z.height()-1 || !z.getDotByCoor(d.y+1, d.x).isNothing() {
			return false
		}
	}
	return true
}

// check if the block can move left
func (z zone) canBlockMoveLeft(b block) bool {
	for _, d := range b {
		if d.x <= 0 || !z.getDotByCoor(d.y, d.x-1).isNothing() {
			return false
		}
	}
	return true
}

// check if the block can move right
func (z zone) canBlockMoveRight(b block) bool {
	for _, d := range b {
		if d.x >= z.width()-1 || !z.getDotByCoor(d.y, d.x+1).isNothing() {
			return false
		}
	}
	return true
}

// check if the block can rotate
func (z zone) canBlockRotate(b block) (block, bool) {
	b = b.rotate()
	// better op
	for b.outBoundTop(0) {
		b = b.moveDown()
	}
	for b.outBoundButtom(z.height() - 1) {
		b = b.moveUp()
	}
	for b.outBoundLeft(0) {
		b = b.moveRight()
	}
	for b.outBoundRight(z.width() - 1) {
		b = b.moveLeft()
	}
	for _, v := range b {
		if !z.getDotByCoor(v.y, v.x).isNothing() {
			return b, false
		}
	}
	return b, true
}

// render zone for AS client
func (z *zone) render(b block) [][]Color {
	// render projection of the block
	projB := b
	for z.canBlockMoveDown(projB) {
		projB = projB.moveDown()
	}
	(&projB).transparentBlock()
	// z.putBlockOnZone(projB)
	// render active block
	// z.putBlockOnZone(b)
	for y := 0; y < z.height(); y++ {
		for x := 0; x < z.width(); x++ {
			z.wrapZone[y][x] = z.getDotByCoor(y, x)
		}
	}
	for _, d := range projB {
		z.wrapZone[d.y][d.x] = d.Color
	}
	for _, d := range b {
		z.wrapZone[d.y][d.x] = d.Color
	}
	// return z.data
	return z.wrapZone
}
