package draw

import (
	"errors"
	"math/rand"
)

var ErrNotEnoughMembers = errors.New("not_enough_members")

// Assign распределяет участников в один цикл.
// Каждый дарит следующему: perm[i] -> perm[(i+1) % n].
// Возвращает map[santaID]recipientID.
func Assign(participants []int64, rng *rand.Rand) (map[int64]int64, error) {
	n := len(participants)
	if n < 2 {
		return nil, ErrNotEnoughMembers
	}

	perm := make([]int64, n)
	copy(perm, participants)

	rng.Shuffle(n, func(i, j int) {
		perm[i], perm[j] = perm[j], perm[i]
	})

	result := make(map[int64]int64, n)
	for i := 0; i < n; i++ {
		result[perm[i]] = perm[(i+1)%n]
	}

	return result, nil
}
