package draw_test

import (
	"math/rand"
	"testing"

	"github.com/andreypisarev/secret-santa/internal/draw"
)

func TestAssign_NotEnoughMembers(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	_, err := draw.Assign([]int64{}, rng)
	if err != draw.ErrNotEnoughMembers {
		t.Errorf("err = %v, want ErrNotEnoughMembers", err)
	}

	_, err = draw.Assign([]int64{1}, rng)
	if err != draw.ErrNotEnoughMembers {
		t.Errorf("err = %v, want ErrNotEnoughMembers", err)
	}
}

func TestAssign_TwoMembers(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	result, err := draw.Assign([]int64{1, 2}, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result[1] == 1 || result[2] == 2 {
		t.Error("someone is assigned to themselves")
	}
	if len(result) != 2 {
		t.Errorf("result size = %d, want 2", len(result))
	}
}

func TestAssign_Invariants(t *testing.T) {
	participants := []int64{10, 20, 30, 40, 50, 60}
	rng := rand.New(rand.NewSource(123))

	result, err := draw.Assign(participants, rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for santa, recipient := range result {
		if santa == recipient {
			t.Errorf("santa %d assigned to self", santa)
		}
	}

	if len(result) != len(participants) {
		t.Errorf("result size = %d, want %d", len(result), len(participants))
	}

	recipients := make(map[int64]bool)
	for _, r := range result {
		if recipients[r] {
			t.Errorf("recipient %d assigned twice", r)
		}
		recipients[r] = true
	}

	visited := make(map[int64]bool)
	current := participants[0]
	for i := 0; i < len(participants); i++ {
		if visited[current] {
			t.Fatalf("cycle broken at step %d, node %d", i, current)
		}
		visited[current] = true
		current = result[current]
	}
	if current != participants[0] {
		t.Error("did not return to start — not a single cycle")
	}
	if len(visited) != len(participants) {
		t.Errorf("visited %d nodes, want %d", len(visited), len(participants))
	}
}

func TestAssign_Deterministic(t *testing.T) {
	participants := []int64{1, 2, 3, 4, 5}

	r1, _ := draw.Assign(participants, rand.New(rand.NewSource(999)))
	r2, _ := draw.Assign(participants, rand.New(rand.NewSource(999)))

	for k, v := range r1 {
		if r2[k] != v {
			t.Errorf("not deterministic: key %d, got %d vs %d", k, v, r2[k])
		}
	}
}

func TestAssign_DoesNotMutateInput(t *testing.T) {
	original := []int64{1, 2, 3, 4, 5}
	input := make([]int64, len(original))
	copy(input, original)

	draw.Assign(input, rand.New(rand.NewSource(42)))

	for i, v := range input {
		if v != original[i] {
			t.Errorf("input[%d] = %d, was %d — input was mutated", i, v, original[i])
		}
	}
}
