package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// TODO: Revisit this once we have propoer gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054, https://github.com/cosmos/cosmos-sdk/discussions/9072
const gasCostPerIteration = uint64(10)

var _ authz.Authorization = &SendAuthorization{}

// NewSendAuthorization creates a new SendAuthorization object.
func NewSendAuthorization(allowed []sdk.AccAddress, spendLimit sdk.Coins) (*SendAuthorization, error) {
	allowedAddrs, err := validateAllowedAddresses(allowed)
	if err != nil {
		return nil, err
	}

	a := SendAuthorization{}
	if allowedAddrs != nil {
		a.AllowList = allowedAddrs
	}

	if spendLimit != nil {
		a.SpendLimit = spendLimit
	}

	return &a, nil
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a SendAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgSend{})
}

// Accept implements Authorization.Accept.
func (a SendAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	var toAddr string
	mSend, ok := msg.(*MsgSend)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}
	toAddr = mSend.ToAddress
	limitLeft, isNegative := a.SpendLimit.SafeSub(mSend.Amount...)
	if isNegative {
		return authz.AcceptResponse{}, sdkerrors.ErrInsufficientFunds.Wrapf("requested amount is more than spend limit")
	}
	if limitLeft.IsZero() {
		return authz.AcceptResponse{Accept: true, Delete: true}, nil
	}

	isAddrExists := false
	allowedList := a.GetAllowList()

	for _, addr := range allowedList {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "send authorization")
		if addr == toAddr {
			isAddrExists = true
			break
		}
	}

	if len(allowedList) > 0 && !isAddrExists {
		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("cannot send to %s address", toAddr)
	}
	return authz.AcceptResponse{Accept: true, Delete: false, Updated: &SendAuthorization{SpendLimit: limitLeft, AllowList: allowedList}}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a SendAuthorization) ValidateBasic() error {
	if a.SpendLimit == nil {
		return sdkerrors.ErrInvalidCoins.Wrap("spend limit cannot be nil")
	}
	if !a.SpendLimit.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrapf("spend limit must be positive")
	}
	return nil
}

func validateAllowedAddresses(allowed []sdk.AccAddress) ([]string, error) {
	if len(allowed) == 0 {
		return nil, nil
	}

	allowedAddrs := make([]string, len(allowed))
	for i, addr := range allowed {
		allowedAddrs[i] = addr.String()
	}
	return allowedAddrs, nil
}
