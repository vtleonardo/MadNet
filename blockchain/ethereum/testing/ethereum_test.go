//go:build integration

package ethereum

import (
	"github.com/MadBase/MadNet/blockchain/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEthereum_AccountsFound(t *testing.T) {
	eth := testutils.GetEthereumNetwork(t, false)
	defer eth.Close()

	accountList := eth.GetKnownAccounts()
	for _, acct := range accountList {
		_, err := eth.GetAccountKeys(acct.Address)
		assert.Nilf(t, err, "Not able to get keys for account: %v", acct.Address)
	}
}

//func TestEthereum_NewEthereumEndpoint(t *testing.T) {
//
//	eth := setupEthereum(t, 4)
//	defer eth.Close()
//
//	type args struct {
//		endpoint                    string
//		pathKeystore                string
//		pathPasscodes               string
//		defaultAccount              string
//		timeout                     time.Duration
//		retryCount                  int
//		retryDelay                  time.Duration
//		finalityDelay               int
//		txFeePercentageToIncrease   int
//		getTxMaxGasFeeAllowedInGwei uint64
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    bool
//		wantErr assert.ErrorAssertionFunc
//	}{
//
//		{
//			name: "Create new ethereum endpoint failing with passcode file not found",
//			args: args{"", "", "", "", 0, 0, 0, 0, 0, 0},
//			want: false,
//			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
//				_, ok := err.(*fs.PathError)
//				if !ok {
//					t.Errorf("Failing test with an unexpected error")
//				}
//				return ok
//			},
//		},
//		{
//			name: "Create new ethereum endpoint failing with specified account not found",
//			args: args{"", "", "../assets/test/passcodes.txt", "", 0, 0, 0, 0, 0, 0},
//			want: false,
//			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
//				if !errors.Is(err, ethereum.ErrAccountNotFound) {
//					t.Errorf("Failing test with an unexpected error")
//				}
//				return true
//			},
//		},
//		{
//			name: "Create new ethereum endpoint failing on Dial Context",
//			args: args{
//				eth.GetEndpoint(),
//				"../assets/test/keys",
//				"../assets/test/passcodes.txt",
//				eth.GetDefaultAccount().Address.String(),
//				eth.Timeout(),
//				eth.RetryCount(),
//				eth.RetryDelay(),
//				int(eth.GetFinalityDelay()),
//				eth.GetTxFeePercentageToIncrease(),
//				eth.GetTxMaxGasFeeAllowedInGwei(),
//			},
//			want: false,
//			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
//				_, ok := err.(*net.OpError)
//				if !ok {
//					t.Errorf("Failing test with an unexpected error")
//				}
//				return ok
//			},
//		},
//		{
//			name: "Create new ethereum endpoint returning EthereumDetails",
//			args: args{
//				"http://localhost:8545",
//				"../assets/test/keys",
//				"../assets/test/passcodes.txt",
//				eth.GetDefaultAccount().Address.String(),
//				eth.Timeout(),
//				eth.RetryCount(),
//				eth.RetryDelay(),
//				int(eth.GetFinalityDelay()),
//				eth.GetTxFeePercentageToIncrease(),
//				eth.GetTxMaxGasFeeAllowedInGwei(),
//			},
//			want: true,
//			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
//				return true
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := ethereum.NewEndpoint(tt.args.endpoint, tt.args.pathKeystore, tt.args.pathPasscodes, tt.args.defaultAccount, tt.args.timeout, tt.args.retryCount, tt.args.retryDelay, tt.args.txFeePercentageToIncrease, tt.args.getTxMaxGasFeeAllowedInGwei)
//			if !tt.wantErr(t, err, fmt.Sprintf("NewEthereumEndpoint(%v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v)", tt.args.endpoint, tt.args.pathKeystore, tt.args.pathPasscodes, tt.args.defaultAccount, tt.args.timeout, tt.args.retryCount, tt.args.retryDelay, tt.args.txFeePercentageToIncrease, tt.args.getTxMaxGasFeeAllowedInGwei)) {
//				return
//			}
//			if tt.want {
//				assert.NotNilf(t, got, "Ethereum Details must not be nil")
//			}
//		})
//	}
//}
