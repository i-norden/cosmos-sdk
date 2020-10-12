# ADR 030: Msg Authorization Module

## Changelog

- 2019-11-06: Initial Draft
- 2020-10-12: Updated Draft

## Status

Accepted

## Abstract

## Context


## Decision

We will create a module named `msg_authorization` which provides support for
granting arbitrary capabilities from one account (the granter) to another account (the grantee). Authorizations
must be granted for a particular type of `Msg` one by one using an implementation
of `Authorization`.

### Types

Authorizations determine exactly what action is granted. They are extensible
and can be defined for any `Msg` service method even outside of the module where
the `Msg` method is defined. `Authorization`s use the new `ServiceMsg` type from
ADR 031.

#### Authorization

```go
type Authorization interface {
	// MethodName returns the fully-qualified method name for the Msg as described in ADR 031.
	MethodName() string

	// Accept determines whether this grant permits the provided sdk.ServiceMsg to be performed, and if
	// so provides an upgraded authorization grant.
	Accept(msg sdk.ServiceMsg, block abci.Header) (allow bool, updated Authorization, delete bool)
}
```

For example a `SendAuthorization` like this is defined for `MsgSend` that takes
a `SpendLimit` and updates it down to zero:

```go
type SendAuthorization struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this authorization and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

func (cap SendAuthorization) MethodName() string {
	return "/cosmos.bank.v1beta1.Msg/Send"
}

func (cap SendAuthorization) Accept(msg sdk.ServiceMsg, block abci.Header) (allow bool, updated Authorization, delete bool) {
	switch req := msg.Request.(type) {
	case bank.MsgSend:
		left, invalid := cap.SpendLimit.SafeSub(req.Amount)
		if invalid {
			return false, nil, false
		}
		if left.IsZero() {
			return true, nil, true
		}
		return true, SendAuthorization{SpendLimit: left}, false
	}
	return false, nil, false
}
```

A different type of capability for `MsgSend` could be implemented
using the `Authorization` interface with no need to change the underlying
`bank` module.

### `Msg` Service

```proto
service Msg {
  // GrantAuthorization grants the provided authorization to the grantee on the granter's
  // account with the provided expiration time.
  rpc GrantAuthorization(MsgGrantAuthorization) returns (MsgGrantAuthorizationResponse);

  // ExecAuthorized attempts to execute the provided messages using
  // authorizations granted to the grantee. Each message should have only
  // one signer corresponding to the granter of the authorization.
  rpc ExecAuthorized(MsgExecAuthorized) returns (MsgExecAuthorizedResponse)


  // RevokeAuthorization revokes any authorization corresponding to the provided method name on the
  // granter's account with that has been granted to the grantee.
  rpc RevokeAuthorization(MsgRevokeAuthorization) returns (MsgRevokeAuthorizationResponse);
}

message MsgGrantAuthorization{
  string granter = 1;
  string grantee = 2;
  google.protobuf.Any authorization = 3 [(cosmos_proto.accepts_interface) = "Authorization"];
  google.protobuf.Timestamp expiration = 4;
}

message MsgExecAuthorized {
    string grantee = 1;
    repeated google.protobuf.Any msgs = 2;
}

message MsgRevokeAuthorization{
  string granter = 1;
  string grantee = 2;
  string method_name = 3;
}
```

### Router Middleware

The `msg_authorization` `Keeper` will expose a `DispatchActions` method which allows other modules to send `ServiceMsg`s
to the router based on `Authorization` grants:

```go
type Keeper interface {
  DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.ServiceMsg) sdk.Result`
}
```

### CLI

#### `--send-as` Flag

When a CLI user wants to run a transaction as another user using `MsgExecAuthorized`, they
can use the `--send-as` flag. For instance `gaiacli tx gov vote 1 yes --from mykey --send-as cosmos3thsdgh983egh823`
would send a transaction like this:

```go
MsgExecAuthorized {
  Grantee: mykey,
  Msgs: []sdk.SericeMsg{
    ServiceMsg {
      MethodName:"/cosmos.gov.v1beta1.Msg/Vote"
      Request: MsgVote {
	    ProposalID: 1,
	    Voter: cosmos3thsdgh983egh823
	    Option: Yes
      }
    }
  }
}
```

#### `tx grant <grantee> <authorization> --from <granter>`

This CLI command will send a `MsgGrantAuthorization` transaction. `authorization` should be encoded as
JSON on the CLI.

#### `tx revoke <grantee> <method-name> --from <granter>`

This CLI command will send a `MsgRevokeAuthorization` transaction.

### Built-in Authorizations

#### `SendAuthorization`

```proto
// SendAuthorization allows the grantee to spend up to spend_limit coins from
// the granter's account.
message SendAuthorization {
  repeated cosmos.base.v1beta1.Coin spend_limit = 1;
}
```

#### `GenericAuthorization`

```proto
// GenericAuthorization gives the grantee unrestricted permissions to execute
// the provide method on behalf of the granter's account.
message GenericAuthorization {
  string method_name = 1;
}
```

## Consequences

### Positive

- Users will be able to authorize arbitrary permissions on their accounts to other
users, simplifying key management for some use cases
- The solution is more generic than previously considered approaches and the
`Authorization` interface approach can be extended to cover other use cases by 
SDK users

### Negative

### Neutral

## References

- Initial Hackatom implementation: https://github.com/cosmos-gaians/cosmos-sdk/tree/hackatom/x/delegation
- Post-Hackatom spec: https://gist.github.com/aaronc/b60628017352df5983791cad30babe56#delegation-module
- B-Harvest subkeys spec: https://github.com/cosmos/cosmos-sdk/issues/4480