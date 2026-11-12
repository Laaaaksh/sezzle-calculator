package calculator

import (
	"github.com/shopspring/decimal"

	"github.com/Laaaaksh/sezzle-calculator/backend/internal/modules/calculator/entities"
)

// normalise puts a result in canonical form. decimal.String already drops
// trailing zeros and exponent notation; zero is returned as decimal.Zero so
// the text is "0" however the zero was produced (for example -5 * 0).
func normalise(value decimal.Decimal) entities.Result {
	if value.IsZero() {
		value = decimal.Zero
	}
	return entities.Result{Text: value.String()}
}
