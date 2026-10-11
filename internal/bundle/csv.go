package bundle

import (
	"bytes"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/yzoz/entropy/internal/domain"
)

// csvHeader is the column order used by CSV exports.
var csvHeader = []string{
	"id", "type", "from", "to", "amount", "currency",
	"issued_at", "expires_at", "status", "signers", "settles", "memo",
}

// EncodeCSV renders contracts as a human-readable CSV table. Unlike the JSON
// bundle this is a reporting format: it is not meant to be re-imported.
func EncodeCSV(contracts []*domain.Contract) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(csvHeader); err != nil {
		return nil, err
	}
	for _, c := range contracts {
		row := []string{
			c.ID,
			string(c.Type),
			c.From,
			c.To,
			domain.FormatAmount(c.Currency, c.Amount),
			string(c.Currency),
			time.Unix(c.IssuedAt, 0).UTC().Format(time.RFC3339),
			formatExpiry(c.ExpiresAt),
			string(c.Status),
			strconv.Itoa(len(c.Signatures)),
			c.Settles,
			c.Memo,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatExpiry(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}
