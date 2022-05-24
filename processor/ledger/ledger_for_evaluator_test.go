package ledger_test

import (
	"crypto/rand"
	"log"
	"testing"

	"github.com/algorand/go-algorand/agreement"
	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/ledger"
	"github.com/algorand/go-algorand/ledger/ledgercore"
	"github.com/algorand/go-algorand/logging"
	"github.com/algorand/go-algorand/rpcs"
	block_processor "github.com/algorand/indexer/processor/blockprocessor"
	indxLeder "github.com/algorand/indexer/processor/ledger"
	"github.com/algorand/indexer/util"
	"github.com/algorand/indexer/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLedgerForEvaluatorLatestBlockHdr(t *testing.T) {

	l := makeTestLedger(t, "ledger")
	defer l.Close()
	pr, _ := block_processor.MakeProcessor(l, nil)
	txn := test.MakePaymentTxn(0, 100, 0, 1, 1, 0, test.AccountA, test.AccountA, basics.Address{}, basics.Address{})
	block, err := test.MakeBlockForTxns(test.MakeGenesisBlock().BlockHeader, &txn)
	assert.Nil(t, err)
	rawBlock := rpcs.EncodedBlockCert{Block: block, Certificate: agreement.Certificate{}}
	err = pr.Process(&rawBlock)
	assert.Nil(t, err)

	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	ret, err := ld.LatestBlockHdr()
	require.NoError(t, err)

	assert.Equal(t, block.BlockHeader, ret)
}

func TestLedgerForEvaluatorAccountDataBasic(t *testing.T) {
	l := makeTestLedger(t, "ledger")
	defer l.Close()
	block_processor.MakeProcessor(l, nil)
	accountData, _, err := l.LookupWithoutRewards(0, test.AccountB)
	require.NoError(t, err)

	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	ret, err :=
		ld.LookupWithoutRewards(map[basics.Address]struct{}{test.AccountB: {}})
	require.NoError(t, err)

	accountDataRet := ret[test.AccountB]
	require.NotNil(t, accountDataRet)
	assert.Equal(t, accountData, *accountDataRet)
}

func TestLedgerForEvaluatorAccountDataMissingAccount(t *testing.T) {
	l := makeTestLedger(t, "ledger")
	defer l.Close()
	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	var addr basics.Address
	_, err = rand.Read(addr[:])
	ret, err :=
		ld.LookupWithoutRewards(map[basics.Address]struct{}{addr: {}})
	require.NoError(t, err)

	accountDataRet := ret[addr]
	assert.Nil(t, accountDataRet)
}

func TestLedgerForEvaluatorAsset(t *testing.T) {
	l := makeTestLedger(t, "ledger")
	defer l.Close()
	pr, _ := block_processor.MakeProcessor(l, nil)

	txn0 := test.MakeAssetConfigTxn(0, 2, 0, false, "", "", "", test.AccountA)
	txn1 := test.MakeAssetConfigTxn(0, 4, 0, false, "", "", "", test.AccountA)
	txn2 := test.MakeAssetConfigTxn(0, 6, 0, false, "", "", "", test.AccountA)
	txn3 := test.MakeAssetConfigTxn(0, 8, 0, false, "", "", "", test.AccountB)
	txn4 := test.MakeAssetDestroyTxn(1, test.AccountA)

	block, err := test.MakeBlockForTxns(test.MakeGenesisBlock().BlockHeader, &txn0, &txn1, &txn2, &txn3, &txn4)
	assert.Nil(t, err)
	rawBlock := rpcs.EncodedBlockCert{Block: block, Certificate: agreement.Certificate{}}
	err = pr.Process(&rawBlock)
	assert.Nil(t, err)

	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	ret, err :=
		ld.LookupResources(map[basics.Address]map[ledger.Creatable]struct{}{
			test.AccountA: {
				{Index: 1, Type: basics.AssetCreatable}: {},
				{Index: 2, Type: basics.AssetCreatable}: {},
				{Index: 3, Type: basics.AssetCreatable}: {},
			},
			test.AccountB: {
				{Index: 4, Type: basics.AssetCreatable}: {},
			},
		})
	require.NoError(t, err)

	expected := map[basics.Address]map[ledger.Creatable]ledgercore.AccountResource{
		test.AccountA: {
			ledger.Creatable{Index: 1, Type: basics.AssetCreatable}: {},
			ledger.Creatable{Index: 2, Type: basics.AssetCreatable}: {
				AssetHolding: &basics.AssetHolding{
					Amount: 4,
					Frozen: false,
				},
				AssetParams: &basics.AssetParams{
					Total:         4,
					Decimals:      0,
					DefaultFrozen: false,
					UnitName:      "",
					AssetName:     "",
					URL:           "",
					MetadataHash:  [32]byte{},
					Manager:       test.AccountA,
					Reserve:       test.AccountA,
					Freeze:        test.AccountA,
					Clawback:      test.AccountA,
				},
			},
			ledger.Creatable{Index: 3, Type: basics.AssetCreatable}: {
				AssetHolding: &basics.AssetHolding{
					Amount: 6,
					Frozen: false,
				},
				AssetParams: &basics.AssetParams{
					Total:         6,
					Decimals:      0,
					DefaultFrozen: false,
					UnitName:      "",
					AssetName:     "",
					URL:           "",
					MetadataHash:  [32]byte{},
					Manager:       test.AccountA,
					Reserve:       test.AccountA,
					Freeze:        test.AccountA,
					Clawback:      test.AccountA,
				},
			},
		},
		test.AccountB: {
			ledger.Creatable{Index: 4, Type: basics.AssetCreatable}: {
				AssetHolding: &basics.AssetHolding{
					Amount: 8,
					Frozen: false,
				},
				AssetParams: &basics.AssetParams{
					Total:         8,
					Decimals:      0,
					DefaultFrozen: false,
					UnitName:      "",
					AssetName:     "",
					URL:           "",
					MetadataHash:  [32]byte{},
					Manager:       test.AccountB,
					Reserve:       test.AccountB,
					Freeze:        test.AccountB,
					Clawback:      test.AccountB,
				},
			},
		},
	}
	assert.Equal(t, expected, ret)
}

func TestLedgerForEvaluatorApp(t *testing.T) {
	l := makeTestLedger(t, "ledger")
	defer l.Close()
	pr, _ := block_processor.MakeProcessor(l, nil)

	txn0 := test.MakeAppCallTxn(0, test.AccountA)
	txn1 := test.MakeAppCallTxnWithLogs(0, test.AccountA, []string{"testing"})
	txn2 := test.MakeAppCallWithInnerTxn(test.AccountA, test.AccountA, test.AccountB, basics.Address{}, basics.Address{})
	txn3 := test.MakeAppCallWithMultiLogs(test.AccountA)
	txn4 := test.MakeAppDestroyTxn(1, test.AccountA)
	txn5 := test.MakeAppCallTxn(0, test.AccountB)

	block, err := test.MakeBlockForTxns(test.MakeGenesisBlock().BlockHeader, &txn0, &txn1, &txn2, &txn3, &txn4, &txn5)
	assert.Nil(t, err)
	rawBlock := rpcs.EncodedBlockCert{Block: block, Certificate: agreement.Certificate{}}
	err = pr.Process(&rawBlock)
	assert.Nil(t, err)

	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	ret, err :=
		ld.LookupResources(map[basics.Address]map[ledger.Creatable]struct{}{
			test.AccountA: {
				{Index: 1, Type: basics.AppCreatable}: {},
				{Index: 2, Type: basics.AppCreatable}: {},
				{Index: 3, Type: basics.AppCreatable}: {},
				{Index: 4, Type: basics.AppCreatable}: {},
			},
			test.AccountB: {
				{Index: 6, Type: basics.AppCreatable}: {},
			},
		})
	require.NoError(t, err)

	expected := map[basics.Address]map[ledger.Creatable]ledgercore.AccountResource{
		test.AccountA: {
			ledger.Creatable{Index: 1, Type: basics.AppCreatable}: {},
			ledger.Creatable{Index: 2, Type: basics.AppCreatable}: {
				AppParams: &basics.AppParams{
					ApprovalProgram:   []byte{0x06, 0x81, 0x01},
					ClearStateProgram: []byte{0x06, 0x81, 0x01},
					GlobalState:       nil,
					StateSchemas:      basics.StateSchemas{},
					ExtraProgramPages: 0,
				},
			},
			ledger.Creatable{Index: 3, Type: basics.AppCreatable}: {
				AppParams: &basics.AppParams{
					ApprovalProgram:   []byte{0x06, 0x81, 0x01},
					ClearStateProgram: []byte{0x06, 0x81, 0x01},
					GlobalState:       nil,
					StateSchemas:      basics.StateSchemas{},
					ExtraProgramPages: 0,
				},
			},
			ledger.Creatable{Index: 4, Type: basics.AppCreatable}: {
				AppParams: &basics.AppParams{
					ApprovalProgram:   []byte{0x06, 0x81, 0x01},
					ClearStateProgram: []byte{0x06, 0x81, 0x01},
					GlobalState:       nil,
					StateSchemas:      basics.StateSchemas{},
					ExtraProgramPages: 0,
				},
			},
		},
		test.AccountB: {
			ledger.Creatable{Index: 6, Type: basics.AppCreatable}: {
				AppParams: &basics.AppParams{
					ApprovalProgram:   []byte{0x06, 0x81, 0x01},
					ClearStateProgram: []byte{0x06, 0x81, 0x01},
					GlobalState:       nil,
					StateSchemas:      basics.StateSchemas{},
					ExtraProgramPages: 0,
				},
			},
		},
	}
	assert.Equal(t, expected, ret)
}

func TestLedgerForEvaluatorFetchAllResourceTypes(t *testing.T) {
	l := makeTestLedger(t, "ledger")
	defer l.Close()
	pr, _ := block_processor.MakeProcessor(l, nil)

	txn0 := test.MakeAppCallTxn(0, test.AccountA)
	txn1 := test.MakeAssetConfigTxn(0, 2, 0, false, "", "", "", test.AccountA)

	block, err := test.MakeBlockForTxns(test.MakeGenesisBlock().BlockHeader, &txn0, &txn1)
	assert.Nil(t, err)
	rawBlock := rpcs.EncodedBlockCert{Block: block, Certificate: agreement.Certificate{}}
	err = pr.Process(&rawBlock)
	assert.Nil(t, err)

	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	ret, err :=
		ld.LookupResources(map[basics.Address]map[ledger.Creatable]struct{}{
			test.AccountA: {
				{Index: 1, Type: basics.AppCreatable}:   {},
				{Index: 2, Type: basics.AssetCreatable}: {},
			},
		})
	require.NoError(t, err)

	expected := map[basics.Address]map[ledger.Creatable]ledgercore.AccountResource{
		test.AccountA: {
			ledger.Creatable{Index: 1, Type: basics.AppCreatable}: {
				AppParams: &basics.AppParams{
					ApprovalProgram:   []byte{0x06, 0x81, 0x01},
					ClearStateProgram: []byte{0x06, 0x81, 0x01},
					GlobalState:       nil,
					StateSchemas:      basics.StateSchemas{},
					ExtraProgramPages: 0,
				},
			},
			ledger.Creatable{Index: 2, Type: basics.AssetCreatable}: {
				AssetHolding: &basics.AssetHolding{
					Amount: 2,
					Frozen: false,
				},
				AssetParams: &basics.AssetParams{
					Total:         2,
					Decimals:      0,
					DefaultFrozen: false,
					UnitName:      "",
					AssetName:     "",
					URL:           "",
					MetadataHash:  [32]byte{},
					Manager:       test.AccountA,
					Reserve:       test.AccountA,
					Freeze:        test.AccountA,
					Clawback:      test.AccountA,
				},
			},
		},
	}
	assert.Equal(t, expected, ret)
}

func TestLedgerForEvaluatorLookupMultipleAccounts(t *testing.T) {
	l := makeTestLedger(t, "ledger")
	defer l.Close()
	block_processor.MakeProcessor(l, nil)

	addresses := []basics.Address{
		test.AccountA, test.AccountB, test.AccountC, test.AccountD}

	addressesMap := make(map[basics.Address]struct{})
	for _, address := range addresses {
		addressesMap[address] = struct{}{}
	}
	addressesMap[test.FeeAddr] = struct{}{}
	addressesMap[test.RewardAddr] = struct{}{}

	ld, err := indxLeder.MakeLedgerForEvaluator(l)
	require.NoError(t, err)
	defer ld.Close()

	ret, err :=
		ld.LookupWithoutRewards(addressesMap)
	require.NoError(t, err)

	for _, address := range addresses {
		accountData, _ := ret[address]
		require.NotNil(t, accountData)
	}
}

func TestLedgerForEvaluatorAssetCreatorBasic(t *testing.T) {

}

func TestLedgerForEvaluatorAssetCreatorMultiple(t *testing.T) {

}

func TestLedgerForEvaluatorAppCreatorBasic(t *testing.T) {

}

func TestLedgerForEvaluatorAppCreatorMultiple(t *testing.T) {

}

func TestLedgerForEvaluatorAccountTotals(t *testing.T) {

}

func makeTestLedger(t *testing.T, prefix string) *ledger.Ledger {
	// initialize local ledger
	genesis := test.MakeGenesis()
	genesisBlock := test.MakeGenesisBlock()
	initState, err := util.CreateInitState(&genesis, &genesisBlock)
	if err != nil {
		log.Panicf("test init err: %v", err)
	}
	logger := logging.NewLogger()
	l, err := ledger.OpenLedger(logger, prefix, true, initState, config.GetDefaultLocal())
	if err != nil {
		log.Panicf("test init err: %v", err)
	}
	return l
}
