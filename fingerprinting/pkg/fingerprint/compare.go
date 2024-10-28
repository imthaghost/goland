package fingerprint

// Option 1: Increase maxDistance significantly
const maxDistance = 16 // Increase from 5 to 16 to catch the lower range of distances

// OR Option 2: Use a percentage-based approach
func compareHashes(hash1, hash2 string, maxDistance int) bool {
	distance := hammingDistance(hash1, hash2)
	maxAllowedDistance := len(hash1) * 30 / 100 // Allow up to 30% difference
	return distance <= maxAllowedDistance
}

func hammingDistance(s1, s2 string) int {
	if len(s1) != len(s2) {
		return -1
	}
	distance := 0
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			distance++
		}
	}
	return distance
}
