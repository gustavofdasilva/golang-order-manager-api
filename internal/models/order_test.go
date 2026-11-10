package models

import "testing"

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from OrderStatus
		to   OrderStatus
		want bool
	}{
		{"pending to paid is allowed", OrderStatusPending, OrderStatusPaid, true},
		{"pending to cancelled is allowed", OrderStatusPending, OrderStatusCancelled, true},
		{"paid to completed is allowed", OrderStatusPaid, OrderStatusCompleted, true},
		{"paid to cancelled is allowed", OrderStatusPaid, OrderStatusCancelled, true},
		{"pending to completed is not allowed", OrderStatusPending, OrderStatusCompleted, false},
		{"paid to pending is not allowed", OrderStatusPaid, OrderStatusPending, false},
		{"cancelled to paid is not allowed", OrderStatusCancelled, OrderStatusPaid, false},
		{"completed to cancelled is not allowed", OrderStatusCompleted, OrderStatusCancelled, false},
		{"completed to paid is not allowed", OrderStatusCompleted, OrderStatusPaid, false},
		{"cancelled to completed is not allowed", OrderStatusCancelled, OrderStatusCompleted, false},
		{"cancelled to pending is not allowed", OrderStatusCancelled, OrderStatusPending, false},
		{"completed to pending is not allowed", OrderStatusCompleted, OrderStatusPending, false},
		{"invalid to paid is not allowed", OrderStatus("invalid"), OrderStatusPaid, false},
		{"pending to invalid is not allowed", OrderStatusPending, OrderStatus("invalid"), false},
		{"paid to invalid is not allowed", OrderStatusPaid, OrderStatus("invalid"), false},
		{"pending to pending is not allowed", OrderStatusPending, OrderStatusPending, false},
		{"paid to paid is not allowed", OrderStatusPaid, OrderStatusPaid, false},
		{"cancelled to cancelled is not allowed", OrderStatusCancelled, OrderStatusCancelled, false},
		{"completed to completed is not allowed", OrderStatusCompleted, OrderStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.want {
				t.Errorf("%s -> %s: got: %v, want: %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
