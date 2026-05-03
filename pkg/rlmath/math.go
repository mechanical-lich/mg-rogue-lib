package rlmath

import "math/rand"

func GetRandom(low int, high int) int {
	if high < low {
		return low
	}
	if low == high {
		return low
	}
	return (rand.Intn((high - low))) + low
}

// Distance returns an estimated Chebyshev distance between two points.
func Distance(x1 int, y1 int, x2 int, y2 int) int {
	var dy int
	if y1 > y2 {
		dy = y1 - y2
	} else {
		dy = (y2 - y1)
	}

	var dx int
	if x1 > x2 {
		dx = x1 - x2
	} else {
		dx = x2 - x1
	}

	var d int
	if dy > dx {
		d = dy + (dx >> 1)
	} else {
		d = dx + (dy >> 1)
	}

	return d
}

func Shuffle(s []string) {
	rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

func Sgn(a int) int {
	switch {
	case a < 0:
		return -1
	case a > 0:
		return +1
	}
	return 0
}

// BresenhamLine returns the tiles from (x1,y1) exclusive to (x2,y2) inclusive
// using Bresenham's line algorithm. Returns nil if (x1,y1)==(x2,y2).
func BresenhamLine(x1, y1, x2, y2 int) [][2]int {
	dx := x2 - x1
	dy := y2 - y1
	adx := dx
	if adx < 0 {
		adx = -adx
	}
	ady := dy
	if ady < 0 {
		ady = -ady
	}
	if adx == 0 && ady == 0 {
		return nil
	}
	sx := 1
	if dx < 0 {
		sx = -1
	}
	sy := 1
	if dy < 0 {
		sy = -1
	}

	tiles := make([][2]int, 0, adx+ady)
	x, y := x1, y1
	if adx >= ady {
		t := ady*2 - adx
		for i := 0; i < adx; i++ {
			if t >= 0 {
				y += sy
				t -= adx * 2
			}
			x += sx
			t += ady * 2
			tiles = append(tiles, [2]int{x, y})
		}
	} else {
		t := adx*2 - ady
		for i := 0; i < ady; i++ {
			if t >= 0 {
				x += sx
				t -= ady * 2
			}
			y += sy
			t += adx * 2
			tiles = append(tiles, [2]int{x, y})
		}
	}
	return tiles
}

// BestFacingDirection returns the closest 4-direction value (0=right,1=down,2=up,3=left)
// for the vector from (fx,fy) toward (tx,ty).
func BestFacingDirection(fx, fy, tx, ty int) int {
	dx := tx - fx
	dy := ty - fy
	adx := dx
	if adx < 0 {
		adx = -adx
	}
	ady := dy
	if ady < 0 {
		ady = -ady
	}
	if adx >= ady {
		if dx >= 0 {
			return 0 // right
		}
		return 3 // left
	}
	if dy >= 0 {
		return 1 // down
	}
	return 2 // up
}
