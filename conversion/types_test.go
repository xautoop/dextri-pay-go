package conversion

import "testing"

func TestUpdatePriceRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request UpdatePriceRequest
		wantErr bool
	}{
		{
			name:    "valid spread",
			request: UpdatePriceRequest{BuyBasePrice: "0.95", SellBasePrice: "1.05"},
		},
		{
			name:    "inverted spread",
			request: UpdatePriceRequest{BuyBasePrice: "1.05", SellBasePrice: "0.95"},
			wantErr: true,
		},
		{
			name:    "price scale exceeds API contract",
			request: UpdatePriceRequest{BuyBasePrice: "0.1234567890123456789", SellBasePrice: "1"},
			wantErr: true,
		},
		{
			name:    "zero price",
			request: UpdatePriceRequest{BuyBasePrice: "0", SellBasePrice: "1"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
