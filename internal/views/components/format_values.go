package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/akrck02/go-renderer/draw2d"
)

// Colour is a theme colour, named here so profiles need not import draw2d for it.
type Colour = draw2d.Color

// spanishMonths are the short month names, January first.
var spanishMonths = [...]string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

// FormatMM writes millimetres with two decimals and a decimal comma: 61,23.
func FormatMM(value float64) string {
	return strings.Replace(fmt.Sprintf("%.2f", value), ".", ",", 1)
}

// FormatDate writes a date as "el 1 nov, 22:10", or "en fecha desconocida" when it is unknown.
func FormatDate(date time.Time) string {
	if date.IsZero() {
		return "en fecha desconocida"
	}
	return fmt.Sprintf("el %d %s, %02d:%02d", date.Day(), spanishMonths[date.Month()-1], date.Hour(), date.Minute())
}
