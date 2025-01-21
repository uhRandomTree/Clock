package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strconv"
)

var (
	SquareSegments bool
	SegmentType string
	SixNineHaveTail bool
	Nine uint8 = 0b11100110
	Six uint8 = 0b00111110
	Colon [9][5]bool
	DigitColour string
	BackgroundColour string
	DigitRGB = [3]uint8{255, 255, 255}
	BackgroundRGB [3]uint8
	ShowHexError error
)

func DigitMatrix(DigitWanted uint8) (DrawTable [9][5]bool) {
	//Since a 7-segment digit happens to have less states than a byte, we can easily fit all the segment data for a number into a uint8.
	//I'll use a lookup table
	//Fix descriptions later

	//Segments will be clockwise from the top, with the middle segment coming last, MSB order.
	//Last bit is unused. For now.
	var(
		LookupTable = [10]uint8{
		0b11111100,//0
		0b01100000,//1
		0b11011010,//2
		0b11110010,//3
		0b01100110,//4
		0b10110110,//5
		Six,//6
		0b11100000,//7
		0b11111110,//8
		Nine}//9
		RequestedDigitData = LookupTable[DigitWanted]

		//The action of the segments will be also controlled by binary data.
		//On a grid 5 wide and 9 tall, you need 3 and 4 bits, respectively to represent cartesian coordinates.
		//We can use those for location data, and the extra bit for directional. (Do we write the segment vertically or horizontally?)
		//Let's do that: 3 bits for X, 4 for Y, and 1 for direction, 0 being horizontal.

		SegmentsTable = [7]uint8{//Following segment order.
		//As indexes, not regular numbers. (0-4, not 1-5)
		0b001_0000_0,//1, 0, across
		0b100_0001_1,//4, 1, down
		0b100_0101_1,//4, 5, down
		0b001_1000_0,//1, 8, across
		0b000_0101_1,//0, 5, down
		0b000_0001_1,//0, 1, down
		0b001_0100_0}//1, 4, across
	)
	
	for BitCounter := uint8(0) ; BitCounter < 8 ; BitCounter++ {
		
		if RequestedDigitData & 0b10000000 != 0 {//Equivalent to  RequestedDigitData & 0b10000000 == 0b10000000
			var(
				CurrentSegment = SegmentsTable[BitCounter]
			//Declaring in here so it's not pulling it if it doesn't hit.
				X = (CurrentSegment & 0b11100000) >> 5
				Y = (CurrentSegment & 0b00011110) >> 1
				Direction bool = (CurrentSegment & 0b00000001) == 0b00000001
			)

			if Direction {
				DrawTable[Y][X] = true
				DrawTable[Y+1][X] = true
				DrawTable[Y+2][X] = true
				//Absolutely gotta be a better way to do this, will deal with later.
			} else {
				DrawTable[Y][X] = true
				DrawTable[Y][X+1] = true
				DrawTable[Y][X+2] = true
			}
		}
		RequestedDigitData <<= 1
	}
	return
}

func RenderDigits(DrawBoard [9][]bool) {
	for _, Window := range DrawBoard {
		for _, Cell := range Window {
				if Cell {
					if SquareSegments {fmt.Print(SegmentType, SegmentType)} else {fmt.Print(SegmentType)}
				} else {
					if SquareSegments {fmt.Print("  ")} else {fmt.Print(" ")}
				}
			}
			fmt.Println()
		}
}

func CombineDrawTables(DrawTable ...[9][5]bool) (DrawBoard [9][]bool) {
	for _, Window := range DrawTable {
		for Index, Row := range Window {
			DrawBoard[Index] = append(DrawBoard[Index], Row[:]...)
			DrawBoard[Index] = append(DrawBoard[Index], false)
		}
	}
	 return
}

//Braille render
//3 tall
//24 wide
//Character is 0b0010_1000_
// a blank character is _0000_0000

// ⠁ is _0000_0001
// ⠂ is _0000_0010
// ⠃ is _0000_0011
// ⠄ is _0000_0100

// ⠞ is _0001_1110
// ⠟ is _0001_1111

// ⠿ is _0011_1111
// ⡀ is _0100_0000

//We've got, let's say, f, t, t, t, f, f, f, t, t...
//iterate through the three vertical, then the next column, etc.
//0,0         0,1        0,2.          1,0          1,1          1,2
//0:2,0:4   0:2,4:8   0:2,8:12   2:4,0:4    2:4,4:8     2:4,8:12

// func BrailleRender (DrawBoard [9][]bool) {
// 	var BrailleTable [3][24]rune
	
// }

func DigitMatrixMini(DigitWanted uint8) (DrawTable [9][5]bool) {
	var RequestedDigitData=[10]uint8{252,96,218,242,102,182,Six,224,254,Nine}[DigitWanted];for BitCounter:=uint8(0);BitCounter<8;BitCounter++{if RequestedDigitData&128==128{var(CurrentSegment=[7]uint8{32,131,139,48,11,3,40}[BitCounter];X=(CurrentSegment&224)>>5;Y=(CurrentSegment&30)>>1;Direction bool=CurrentSegment&1==1);if Direction{DrawTable[Y][X],DrawTable[Y+1][X],DrawTable[Y+2][X]=true,true,true}else{DrawTable[Y][X],DrawTable[Y][X+1],DrawTable[Y][X+2]=true,true,true}};RequestedDigitData<<=1};return}

func ColourHexString(Colour string, Container *[3]uint8) {
	if Colour != "none" {
		if Colour[0:2] == "0x" || Colour[0:2] == "0X" {
			Colour = Colour[2:]
		} else if Colour[0] == 'h' || Colour[0] == '$' || Colour[0] == '#' {
			Colour = Colour[1:]
		}
		for i := range 3 {
			TempInt64, Error := strconv.ParseInt(Colour[i*2 : i*2+2], 16, 16)
			if Error != nil {
				ShowHexError = Error
				break
			} else {
				Container[i] = uint8(TempInt64)
			}
		}
	}
}

func main() {
	var GracefulKill chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(GracefulKill, 
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	go func() {
		<-GracefulKill
		fmt.Print("\x1b[10B\x1b[0m\x1b[?25h")
		if ShowHexError != nil {
			fmt.Print(ShowHexError)
		}
		os.Exit(0)
	}()

	flag.BoolVar(&SquareSegments, "ss", false, "Disables square aspect ratio for the segments.")
	flag.StringVar(&SegmentType, "st", "█", "The character the segments are made of.")
	flag.BoolVar(&SixNineHaveTail, "t", false, "Removes tail from six and nine.")
	flag.StringVar(&DigitColour, "dc", "none", "The colour of the digit as a hex code.")
	flag.StringVar(&BackgroundColour, "bc", "none", "The colour of the background as a hex code.")
	flag.Parse()
	if SquareSegments { SquareSegments = false } else {SquareSegments = true}
	if SixNineHaveTail { SixNineHaveTail = false } else {SixNineHaveTail = true}
	
	ColourHexString(DigitColour, &DigitRGB)
	ColourHexString(BackgroundColour, &BackgroundRGB)
	
	if SixNineHaveTail {
		Nine = 0b11110110
		Six = 0b10111110
	}

	Colon[2][2] = true
	Colon[6][2] = true

	fmt.Print("\x1b[?25l")
	if DigitColour != "none" {fmt.Printf("\x1b[38;2;%d;%d;%dm", DigitRGB[0], DigitRGB[1], DigitRGB[2])}
	if BackgroundColour != "none" {fmt.Printf("\x1b[48;2;%d;%d;%dm", BackgroundRGB[0], BackgroundRGB[1], BackgroundRGB[2])}

	var ClockTicker = time.NewTicker(time.Second)
	for {
		select {
		case DisplayTime := <-ClockTicker.C:
			var iHour, iMinute, iSecond int = DisplayTime.Clock()
			var Hour, Minute, Second uint8 = uint8(iHour), uint8(iMinute), uint8(iSecond)
			
			RenderDigits(
				CombineDrawTables(
				DigitMatrix(Hour/10),
				DigitMatrix(Hour%10),
				Colon,
				DigitMatrix(Minute/10),
				DigitMatrix(Minute%10),
				Colon,
				DigitMatrix(Second/10),
				DigitMatrix(Second%10)))

			fmt.Print(DisplayTime.Date())
			fmt.Print("\x1b[9A\r")
		}
	}

	// fmt.Print("\x1b[?25h")


}