package permcalc

import "github.com/nsf/termbox-go"

type gamedata struct {
	running  bool
	x, y     int
	perm     int
	readonly bool
}

var data gamedata

func handleKey(key termbox.Key, char rune) {
	switch key {
	case termbox.KeyArrowUp:
		moveY(true)
	case termbox.KeyArrowLeft:
		moveX(false)
	case termbox.KeyArrowRight:
		moveX(true)
	case termbox.KeyArrowDown:
		moveY(false)

	case termbox.KeySpace:
		if data.readonly {
			break
		}
		data.perm = data.perm ^ PermOrder[pos2perm(data.x, data.y)]
	case termbox.KeyEsc:
		data.running = false
	}

	switch char {
	case 'q':
		data.running = false
	}
}

func moveX(right bool) {
	if right && data.y == optionY1+optionH1-1 {
		data.y--
	}

	if right {
		if data.x == optionX1+1 {
			data.x = optionX2 + 1
		}
	} else {
		if data.x == optionX2+1 {
			data.x = optionX1 + 1
		}
	}
}
func moveY(up bool) {
	y := data.y
	switch {
	case up && y == optionY1:
	case !up && y == optionY3+optionH3-1:

	case y == optionY1+optionH1-1 && data.x == optionX1+1 && !up:
		fallthrough
	case !up && y == optionY1+optionH1-2 && data.x == optionX2+1:
		data.y = optionY2
	case up && y == optionY2:
		data.y = optionY1 + optionH1 - 1
		if data.x == optionX2+1 {
			data.y--
		}

	case !up && y == optionY2+optionH2-1:
		data.y = optionY3
	case up && y == optionY3:
		data.y = optionY2 + optionH2 - 1

	default:
		if up {
			data.y--
		} else {
			data.y++
		}
	}
}

const optionX1 = 1
const optionY1 = 2

const optionH1 = 6
const optionH2 = 5
const optionH3 = 3

const offset = 2

const optionX2 = 30
const optionY2 = optionY1 + optionH1 + offset
const optionY3 = optionY2 + optionH2 + offset
const optionY4 = optionY3 + optionH3 + offset

func drawScreen() error {
	err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if err != nil {
		return err
	}

	drawString(optionX1-1, optionY1-1, "General Permissions")
	total := 0
	drawOptions(optionX1, optionY1, optionH1, &total)
	drawOptions(optionX2, optionY1, optionH1-1, &total)

	drawString(optionX1-1, optionY2-1, "Text Permissions")
	drawOptions(optionX1, optionY2, optionH2, &total)
	drawOptions(optionX2, optionY2, optionH2, &total)

	drawString(optionX1-1, optionY3-1, "Voice Permissions")
	drawOptions(optionX1, optionY3, optionH3, &total)
	drawOptions(optionX2, optionY3, optionH3, &total)

	y := optionY4
	if !data.readonly {
		drawString(optionX1-1, y, "Press space to toggle permissions.")
		y++
	}
	drawString(optionX1-1, y, "Press Esc or Q to exit.")

	return termbox.Flush()
}

func pos2perm(x, y int) int {
	index := 0
	section2 := x == optionX2+1

	if y >= optionY3 {
		index = optionH1*2 - 1 + optionH2*2 + y - optionY3
		if section2 {
			index += optionH3
		}
	} else if y >= optionY2 {
		index = optionH1*2 - 1 + y - optionY2
		if section2 {
			index += optionH2
		}
	} else {
		index = y - optionY1
		if section2 {
			index += optionH1
		}
	}

	return index
}

func drawOptions(x int, y int, amount int, total *int) {
	for i := 0; i < amount; i++ {
		drawOption(x, y+i, *total+i)
	}
	*total += amount
}
func drawOption(x int, y int, index int) {
	perm := PermOrder[index]

	char := " "
	if data.perm|perm == data.perm {
		char = "*"
	}
	drawString(x, y, "["+char+"] "+PermStrings[perm])
}
func drawString(x int, y int, str string) {
	for i, c := range str {
		drawCell(x+i, y, c, termbox.ColorDefault)
	}
}

func drawCell(x int, y int, c rune, fg termbox.Attribute) {
	bg := termbox.ColorDefault
	if x == data.x && y == data.y {
		bg = termbox.ColorWhite
	}
	termbox.SetCell(x, y, c, fg, bg)
}
