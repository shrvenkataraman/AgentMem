package e2e

import "testing"

func TestT49_QuotedSecretValues(t *testing.T) {
	got := redactedTexts(t,
		`PASSWORD="hunter two"`,
		`export API_KEY='k 1' && go`,
		`DB_TOKEN="a"`,
		`NAME="plain value"`,
	)
	want := []string{
		`PASSWORD=[REDACTED:secret]`,
		`export API_KEY=[REDACTED:secret] && go`,
		`DB_TOKEN=[REDACTED:secret]`,
		`NAME="plain value"`,
	}
	for i := range want {
		eq(t, "case "+string(rune('0'+i)), got[i], want[i])
	}
}
