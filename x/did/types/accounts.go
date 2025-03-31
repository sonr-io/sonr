package types

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/sonr-io/snrd/internal/accounts"
	"github.com/sonr-io/snrd/internal/transaction"
)

var (
	accountsModuleAddress = address.Module("accounts")
	ErrInvalidType        = errors.New("invalid type")
)

// AccountsInterface is the exported interface of an Account.
type AccountsInterface = accounts.Account

// AccountsExecuteBuilder is the exported type of AccountsExecuteBuilder.
type AccountsExecuteBuilder = accounts.ExecuteBuilder

// AccountsQueryBuilder is the exported type of AccountsQueryBuilder.
type AccountsQueryBuilder = accounts.QueryBuilder

// AccountsInitBuilder is the exported type of AccountsInitBuilder.
type AccountsInitBuilder = accounts.InitBuilder

// AccountCreatorFunc is the exported type of AccountCreatorFunc.
type AccountCreatorFunc = accounts.AccountCreatorFunc

func DIAccount[A AccountsInterface](name string, constructor func(deps Dependencies) (A, error)) DepinjectAccount {
	return DepinjectAccount{MakeAccount: AddAccount(name, constructor)}
}

type DepinjectAccount struct {
	MakeAccount AccountCreatorFunc
}

func (DepinjectAccount) IsManyPerContainerType() {}

// Dependencies is the exported type of Dependencies.
type Dependencies = accounts.Dependencies

func RegisterAccountsExecuteHandler[
	Req any, ProtoReq accounts.ProtoMsgG[Req], Resp any, ProtoResp accounts.ProtoMsgG[Resp],
](router *AccountsExecuteBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	accounts.RegisterExecuteHandler(router, handler)
}

// RegisterAccountsQueryHandler registers a query handler for a smart account that uses protobuf.
func RegisterAccountsQueryHandler[
	Req any, ProtoReq accounts.ProtoMsgG[Req], Resp any, ProtoResp accounts.ProtoMsgG[Resp],
](router *AccountsQueryBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	accounts.RegisterQueryHandler(router, handler)
}

// RegisterAccountsInitHandler registers an initialisation handler for a smart account that uses protobuf.
func RegisterAccountsInitHandler[
	Req any, ProtoReq accounts.ProtoMsgG[Req], Resp any, ProtoResp accounts.ProtoMsgG[Resp],
](router *AccountsInitBuilder, handler func(ctx context.Context, req ProtoReq) (ProtoResp, error),
) {
	accounts.RegisterInitHandler(router, handler)
}

// AddAccount is a helper function to add a smart account to the list of smart accounts.
func AddAccount[A AccountsInterface](name string, constructor func(deps Dependencies) (A, error)) AccountCreatorFunc {
	return func(deps accounts.Dependencies) (string, accounts.Account, error) {
		acc, err := constructor(deps)
		return name, acc, err
	}
}

// Whoami returns the address of the account being invoked.
func Whoami(ctx context.Context) []byte {
	return accounts.Whoami(ctx)
}

// Sender returns the sender of the execution request.
func Sender(ctx context.Context) []byte {
	return accounts.Sender(ctx)
}

// HasSender checks if the execution context was sent from the provided sender
func HasSender(ctx context.Context, wantSender []byte) bool {
	return bytes.Equal(Sender(ctx), wantSender)
}

// SenderIsSelf checks if the sender of the request is the account itself.
func SenderIsSelf(ctx context.Context) bool { return HasSender(ctx, Whoami(ctx)) }

// SenderIsAccountsModule returns true if the sender of the execution request is the accounts module.
func SenderIsAccountsModule(ctx context.Context) bool {
	return bytes.Equal(Sender(ctx), accountsModuleAddress)
}

// Funds returns if any funds were sent during the execute or init request. In queries this
// returns nil.
func Funds(ctx context.Context) sdk.Coins { return accounts.Funds(ctx) }

func ExecModule[MsgResp, Msg transaction.Msg](ctx context.Context, msg Msg) (resp MsgResp, err error) {
	untyped, err := accounts.ExecModule(ctx, msg)
	if err != nil {
		return resp, err
	}
	return assertOrErr[MsgResp](untyped)
}

// QueryModule can be used by an account to execute a module query.
func QueryModule[Resp, Req transaction.Msg](ctx context.Context, req Req) (resp Resp, err error) {
	untyped, err := accounts.QueryModule(ctx, req)
	if err != nil {
		return resp, err
	}
	return assertOrErr[Resp](untyped)
}

// UnpackAny unpacks a protobuf Any message generically.
func UnpackAny[Msg any, ProtoMsg accounts.ProtoMsgG[Msg]](any *accounts.Any) (*Msg, error) {
	return accounts.UnpackAny[Msg, ProtoMsg](any)
}

// PackAny packs a protobuf Any message generically.
func PackAny(msg transaction.Msg) (*accounts.Any, error) {
	return accounts.PackAny(msg)
}

// ExecModuleAnys can be used to execute a list of messages towards a module
// when those messages are packed in Any messages. The function returns a list
// of responses packed in Any messages.
func ExecModuleAnys(ctx context.Context, msgs []*accounts.Any) ([]*accounts.Any, error) {
	responses := make([]*accounts.Any, len(msgs))
	for i, msg := range msgs {
		concreteMessage, err := accounts.UnpackAnyRaw(msg)
		if err != nil {
			return nil, fmt.Errorf("error unpacking message %d: %w", i, err)
		}
		resp, err := accounts.ExecModule(ctx, concreteMessage)
		if err != nil {
			return nil, fmt.Errorf("error executing message %d: %w", i, err)
		}
		// pack again
		respAnyPB, err := accounts.PackAny(resp)
		if err != nil {
			return nil, fmt.Errorf("error packing response %d: %w", i, err)
		}
		responses[i] = respAnyPB
	}
	return responses, nil
}

// asserts the given any to the provided generic, returns ErrInvalidType if it can't.
func assertOrErr[T any](r any) (concrete T, err error) {
	concrete, ok := r.(T)
	if !ok {
		return concrete, ErrInvalidType
	}
	return concrete, nil
}
