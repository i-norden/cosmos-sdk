package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/schema"
)

var StakingSchema = []schema.KeyDescriptor{
	{
		Name:   "LastValidatorPowerKey",
		Prefix: LastValidatorPowerKey,
		Parts: []*schema.KeyDescriptor_Part{
			{
				Sum: &schema.KeyDescriptor_Part_Bytes_{
					Bytes: &schema.KeyDescriptor_Part_Bytes{
						Name:        "Operator",
						Description: "Validator operator address",
						FixedWidth:  sdk.AddrLen,
						GoType:   	 &sdk.ValAddress{},
						Relations:   []string{"ValidatorsByConsAddrKey"},
					},
				},
			},
		},
		ValueProtoType: &sdk.IntProto{},
		ValueGoType: &sdk.Int{},
	}, {
		Name:           "LastTotalPowerKey",
		Description:    "",
		Prefix:         LastTotalPowerKey,
		Parts:      	nil,
		ValueProtoType: &sdk.IntProto{},
		ValueGoType: 	&sdk.Int{},
	}, {
		Name:        "ValidatorsKey",
		Description: "",
		Prefix:      ValidatorsKey,
		Parts: []*schema.KeyDescriptor_Part{
			{
				Sum: &schema.KeyDescriptor_Part_Bytes_{
					Bytes: &schema.KeyDescriptor_Part_Bytes{
						Name:        "Operator",
						Description: "Validator operator address",
						FixedWidth:  sdk.AddrLen,
						GoType:   	 &sdk.ValAddress{},
						Relations:   []string{"ValidatorsByConsAddrKey"},
					},
				},
			},
		},
		ValueGoType: &Validator{},
	}, {
		Name:        "ValidatorsByConsAddrKey",
		Description: "",
		Prefix:      ValidatorsByConsAddrKey,
		Parts: []*schema.KeyDescriptor_Part{
			{
				Sum: &schema.KeyDescriptor_Part_Bytes_{
					Bytes: &schema.KeyDescriptor_Part_Bytes{
						Name:        "Operator",
						Description: "Validator consensus address",
						FixedWidth:  sdk.AddrLen,
						GoType:      sdk.ConsAddress{},
					},
				},
			},
		},
		ValueProtoType: &types.{},
		ValueGoType:    sdk.ValAddress{},
	},
}