// Color is not "Color", include other stuffs
// like it may represent "nothing", "stone", "bomb", "transparent-Color"
package tetris

import "fmt"

type Color int

const (
	constColorNothing     = Color(0)
	constColorBomb        = Color(-98)
	constColorStone       = Color(-99)
	constColorTransparent = Color(-1)
	maxColor              = 7
)

func randomColor() Color {
	return Color(randSeed.Intn(maxColor) + 1)
}

func (c Color) String() string {
	return fmt.Sprintf("%v", int(c))
}

func (c Color) isActiveColor() bool {
	return int(c) <= 7 && int(c) >= 1
}

func (c Color) isStone() bool {
	// return int(c) == Color_stone
	return c == constColorStone
}

func (c Color) isNothing() bool {
	// return int(c) == Color_nothing
	return c == constColorNothing
}

func (c Color) isTransparent() bool {
	// return int(c) == -1
	// return int(c) == -int(oc)
	return c == constColorTransparent
}

func (c Color) isBomb() bool {
	// return int(c) == Color_bomb
	return c == constColorBomb
}
