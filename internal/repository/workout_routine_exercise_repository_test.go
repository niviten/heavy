package repository

import "testing"

func TestExerciseOrderRoundTrip(t *testing.T) {
	tests := []struct {
		text  string
		units int64
	}{
		{text: "0", units: 0},
		{text: "1.000", units: 1000},
		{text: "12.345", units: 12345},
		{text: "9999999.999", units: maxExerciseOrder},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			units, err := parseExerciseOrder(test.text)
			if err != nil {
				t.Fatalf("parseExerciseOrder() error = %v", err)
			}
			if units != test.units {
				t.Fatalf("parseExerciseOrder() = %d, want %d", units, test.units)
			}
			if units != 0 && formatExerciseOrder(units) != test.text {
				t.Errorf("formatExerciseOrder() = %q, want %q", formatExerciseOrder(units), test.text)
			}
		})
	}
}

func TestCalculateWorkoutRoutineExerciseOrder(t *testing.T) {
	orders := []workoutRoutineExerciseOrder{
		{Order: 1000},
		{Order: 2000},
		{Order: 3000},
	}
	tests := []struct {
		name      string
		index     int
		want      int64
		available bool
	}{
		{name: "beginning", index: 0, want: 500, available: true},
		{name: "middle", index: 1, want: 1500, available: true},
		{name: "end", index: 3, want: 4000, available: true},
		{name: "invalid", index: 4, available: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, available := calculateWorkoutRoutineExerciseOrder(orders, test.index)
			if got != test.want || available != test.available {
				t.Errorf("calculateWorkoutRoutineExerciseOrder() = (%d, %v), want (%d, %v)",
					got, available, test.want, test.available)
			}
		})
	}
}

func TestCalculateWorkoutRoutineExerciseOrderRequiresNormalization(t *testing.T) {
	tests := []struct {
		name   string
		orders []workoutRoutineExerciseOrder
		index  int
	}{
		{
			name:   "no room before first",
			orders: []workoutRoutineExerciseOrder{{Order: 1}},
			index:  0,
		},
		{
			name: "no room between neighbors",
			orders: []workoutRoutineExerciseOrder{
				{Order: 1000},
				{Order: 1001},
			},
			index: 1,
		},
		{
			name:   "no room after maximum",
			orders: []workoutRoutineExerciseOrder{{Order: maxExerciseOrder}},
			index:  1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, available := calculateWorkoutRoutineExerciseOrder(test.orders, test.index); available {
				t.Error("calculateWorkoutRoutineExerciseOrder() available = true, want false")
			}
		})
	}
}
