package testutils

import (
	"github.com/MadBase/MadNet/blockchain/ethereum"
	"reflect"
	"testing"
)

func TestConnectSimulatorEndpoint(t *testing.T) {

	tests := []struct {
		name       string
		cleanStart bool
		want       ethereum.Network
	}{
		{
			name:       "HardHat not running",
			cleanStart: true,
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetEthereumNetwork(t, tt.cleanStart, 8, ""); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEthereumNetwork() = %v, want %v", got, tt.want)
			}
		})
	}
}
