package consensus

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseAmount(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("empty amount")
	}
	if strings.HasPrefix(raw, "-") {
		return 0, fmt.Errorf("negative amount")
	}

	parts := strings.Split(raw, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount %q", raw)
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 8 {
		return 0, fmt.Errorf("amount has more than 8 decimals")
	}
	for len(frac) < 8 {
		frac += "0"
	}

	fraction, err := strconv.ParseInt(frac, 10, 64)
	if err != nil {
		return 0, err
	}

	return whole*Coin + fraction, nil
}
