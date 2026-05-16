package utils

func IsLuhnValid(number string) bool {
	if len(number) == 0 {
		return false
	}
	var sum int
	parity := len(number) % 2
	for i, r := range number {
		if r < '0' || r > '9' {
			return false
		}
		d := int(r - '0')
		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}
