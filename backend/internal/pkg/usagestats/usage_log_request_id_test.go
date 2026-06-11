package usagestats

import "testing"

func TestResolveUsageLogRequestID(t *testing.T) {
	tests := []struct {
		name              string
		requestID         string
		clientRequestID   string
		want              string
		wantErr           bool
	}{
		{
			name:            "request_id takes precedence",
			requestID:         "client:abc",
			clientRequestID:   "def",
			want:              "client:abc",
		},
		{
			name:      "bare uuid request_id adds client prefix",
			requestID: "550e8400-e29b-41d4-a716-446655440000",
			want:      "client:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "non uuid request_id stays unchanged",
			requestID: "req-upstream-abc",
			want:      "req-upstream-abc",
		},
		{
			name:            "client_request_id adds prefix",
			clientRequestID: "550e8400-e29b-41d4-a716-446655440000",
			want:            "client:550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:            "already prefixed client",
			clientRequestID: "client:abc",
			want:            "client:abc",
		},
		{
			name:            "already prefixed local",
			clientRequestID: "local:demo-1",
			want:            "local:demo-1",
		},
		{
			name:    "missing both",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolveUsageLogRequestID(tc.requestID, tc.clientRequestID)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("ResolveUsageLogRequestID()=%q want %q", got, tc.want)
			}
		})
	}
}
