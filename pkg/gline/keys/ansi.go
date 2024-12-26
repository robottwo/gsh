package keys

// Control keys. See:
// https://en.wikipedia.org/wiki/C0_and_C1_control_codes
const (
	ansiNUL = 0  // null, \0
	ansiSOH = 1  // start of heading
	ansiSTX = 2  // start of text
	ansiETX = 3  // break, ctrl+c
	ansiEOT = 4  // end of transmission
	ansiENQ = 5  // enquiry
	ansiACK = 6  // acknowledge
	ansiBEL = 7  // bell, \a
	ansiBS  = 8  // backspace
	ansiHT  = 9  // horizontal tabulation, \t
	ansiLF  = 10 // line feed, \n
	ansiVT  = 11 // vertical tabulation \v
	ansiFF  = 12 // form feed \f
	ansiCR  = 13 // carriage return, \r
	ansiSO  = 14 // shift out
	ansiSI  = 15 // shift in
	ansiDLE = 16 // data link escape
	ansiDC1 = 17 // device control one
	ansiDC2 = 18 // device control two
	ansiDC3 = 19 // device control three
	ansiDC4 = 20 // device control four
	ansiNAK = 21 // negative acknowledge
	ansiSYN = 22 // synchronous idle
	ansiETB = 23 // end of transmission block
	ansiCAN = 24 // cancel
	ansiEM  = 25 // end of medium
	ansiSUB = 26 // substitution
	ansiESC = 27 // escape, \e
	ansiFS  = 28 // file separator
	ansiGS  = 29 // group separator
	ansiRS  = 30 // record separator
	ansiUS  = 31 // unit separator
)
