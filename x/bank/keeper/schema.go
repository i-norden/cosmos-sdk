package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type tableDecoder struct {
	cdc         codec.BinaryCodec
	legacyAmino *codec.LegacyAmino
}

// NewTableDecoder returns an implementation of the TableDecoder interface
func NewTableDecoder(cdc codec.BinaryCodec, legacyAmino *codec.LegacyAmino) sdk.TableDecoder {
	return &tableDecoder{
		cdc:         cdc,
		legacyAmino: legacyAmino,
	}
}

var _ sdk.TableDecoder = tableDecoder{}

var (
	BalanceTableInfo = sdk.TableInfo{
		Name:             "Balance",
		Type:             &types.BalanceTable{},
		PrimaryKeyFields: []string{"Address", "Demom"},
	}
	SupplyTableInfo = sdk.TableInfo{
		Name:             "Supply",
		Type:             &types.SupplyTable{},
		PrimaryKeyFields: []string{"Denom"},
	}
	MetaDataTableInfo = sdk.TableInfo{
		Name:             "Metadata",
		Type:             &types.DenomMetadataTable{},
		PrimaryKeyFields: []string{"Denom"},
	}
	EnabledTableInfo = sdk.TableInfo{
		Name:             "Enabled",
		Type:             &types.DenomEnabledTable{},
		PrimaryKeyFields: []string{"Denom"},
	}
)

// Schema satisfies the TableDecoder interface
// it returns the underlying TableSchema
func (td tableDecoder) Schema() sdk.TableSchema {
	return sdk.TableSchema{
		Tables: []sdk.TableInfo{
			BalanceTableInfo,
			SupplyTableInfo,
			MetaDataTableInfo,
			EnabledTableInfo,
		},
	}
}

// Decode satisfies the TableDecoder interface
// it decodes a key-value pair into a list of TableUpdates
/// If value is set to nil this indicates that the key-value pair was deleted from storage.
func (td tableDecoder) Decode(key, value []byte) ([]sdk.TableUpdate, error) {
	switch {
	case bytes.HasPrefix(key, types.SupplyKey):
		return td.decodeSupplyUpdate(key, value)
	case bytes.HasPrefix(key, types.BalancesPrefix):
		return td.decodeBalanceUpdate(key, value)
	case bytes.HasPrefix(key, types.DenomMetadataPrefix):
		return td.decodeMetadataUpdate(key, value)
	case bytes.Equal(key, types.KeySendEnabled):
		return td.decodeSendEnabledUpdate(key, value)
	case bytes.Equal(key, types.KeyDefaultSendEnabled):
		return td.decodeDefaultSendEnabledUpdate(key, value)
	}
	return nil, nil
}

func (td tableDecoder) decodeSupplyUpdate(key, value []byte) ([]sdk.TableUpdate, error) {
	denomBytes := bytes.TrimPrefix(key, types.SupplyKey)
	denom := string(denomBytes)

	if value == nil {
		return []sdk.TableUpdate{{
			Table:           "Supply",
			UpdateOrReplace: true,
			Updated: &types.SupplyTable{
				Amount: "",
				Denom:  denom,
			},
			ClearedFields: []string{"Amount"},
		}}, nil
	}

	var amount sdk.Int
	err := amount.Unmarshal(value)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal supply value %v", err)
	}

	return []sdk.TableUpdate{{
		Table:           "Supply",
		UpdateOrReplace: true,
		Updated: &types.SupplyTable{
			Amount: amount.String(),
			Denom:  denom,
		},
	}}, nil
}

func (td tableDecoder) decodeBalanceUpdate(key, value []byte) ([]sdk.TableUpdate, error) {
	addrDenomKey := bytes.TrimPrefix(key, types.BalancesPrefix)
	addrLenByte := addrDenomKey[0]
	addrLen := int(addrLenByte)
	addr := string(addrDenomKey[1 : addrLen+1])
	denom := string(addrDenomKey[addrLen+2:])

	if value == nil {
		return []sdk.TableUpdate{{
			Table:           "Balance",
			UpdateOrReplace: true,
			Updated: &types.BalanceTable{
				Address: addr,
				Denom:   denom,
				Balance: nil,
			},
			ClearedFields: []string{"Balance"},
		}}, nil
	}

	var balance sdk.Coin
	td.cdc.MustUnmarshal(value, &balance)

	return []sdk.TableUpdate{{
		Table:           "Balance",
		UpdateOrReplace: true,
		Updated: &types.BalanceTable{
			Address: addr,
			Denom:   denom,
			Balance: &balance,
		},
	}}, nil
}

func (td tableDecoder) decodeMetadataUpdate(key, value []byte) ([]sdk.TableUpdate, error) {
	denomBytes := bytes.TrimPrefix(key, types.DenomMetadataPrefix)
	denom := string(denomBytes)

	if value == nil {
		return []sdk.TableUpdate{{
			Table:           "Metadata",
			UpdateOrReplace: true,
			Updated: &types.DenomMetadataTable{
				Denom:    denom,
				Metadata: nil,
			},
			ClearedFields: []string{"Metadata"},
		}}, nil
	}

	var metadata types.Metadata
	td.cdc.MustUnmarshal(value, &metadata)

	return []sdk.TableUpdate{{
		Table:           "Metadata",
		UpdateOrReplace: true,
		Updated: &types.DenomMetadataTable{
			Denom: denom,
			Metadata: &types.Metadata{
				Name:        metadata.Name,
				Base:        metadata.Base,
				DenomUnits:  metadata.DenomUnits,
				Description: metadata.Description,
				Display:     metadata.Display,
				Symbol:      metadata.Symbol,
			},
		},
	}}, nil
}

func (td tableDecoder) decodeSendEnabledUpdate(key, value []byte) ([]sdk.TableUpdate, error) {
	if value == nil {
		return []sdk.TableUpdate{{
			Table:           "Enabled",
			UpdateOrReplace: true,
			Updated:         nil,
			ClearedFields:   []string{"Enabled"},
		},
		}, nil
	}

	var sendEnables []*types.SendEnabled
	if err := td.legacyAmino.UnmarshalJSON(value, &sendEnables); err != nil {
		return nil, fmt.Errorf("unable to unmarshal send enabled value")
	}
	updates := make([]sdk.TableUpdate, len(sendEnables))
	for i, se := range sendEnables {
		updates[i] = sdk.TableUpdate{
			Table:           "Enabled",
			UpdateOrReplace: true,
			Updated: &types.DenomEnabledTable{
				Denom:   se.Denom,
				Enabled: se.Enabled,
			},
		}
	}
	return updates, nil
}

func (td tableDecoder) decodeDefaultSendEnabledUpdate(key, value []byte) ([]sdk.TableUpdate, error) {
	if value == nil {
		return []sdk.TableUpdate{{
			Table:           "Enabled",
			UpdateOrReplace: true,
			Updated:         nil,
			ClearedFields:   []string{"Enabled"},
		},
		}, nil
	}

	var defaultSendEnabled bool
	if err := td.legacyAmino.UnmarshalJSON(value, &defaultSendEnabled); err != nil {
		return nil, fmt.Errorf("unable to unmarshal default send enabled value")
	}
	return []sdk.TableUpdate{{
		Table:           "Enabled",
		UpdateOrReplace: true,
		Updated: &types.DenomEnabledTable{
			Denom:   "",
			Enabled: defaultSendEnabled,
		},
	}}, nil
}
