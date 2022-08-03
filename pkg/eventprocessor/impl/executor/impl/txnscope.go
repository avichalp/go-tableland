package impl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/textileio/go-tableland/internal/tableland"
	"github.com/textileio/go-tableland/pkg/eventprocessor/eventfeed"
	"github.com/textileio/go-tableland/pkg/eventprocessor/impl/executor"
	"github.com/textileio/go-tableland/pkg/parsing"
	"github.com/textileio/go-tableland/pkg/tables/impl/ethereum"
)

var tableIDIsEmpty = "table id is empty"

type txnScope struct {
	log zerolog.Logger

	parser    parsing.SQLValidator
	acl       tableland.ACL
	scopeVars scopeVars

	txn *sql.Tx
}

func (ts *txnScope) executeTxnEvents(ctx context.Context, tx *sql.Tx, evmTxn eventfeed.TxnEvents) (executor.TxnExecutionResult, error) {
	var res executor.TxnExecutionResult
	var err error

	for _, event := range evmTxn.Events {
		switch event := event.(type) {
		case *ethereum.ContractRunSQL:
			ts.log.Debug().Str("statement", event.Statement).Msgf("executing run-sql event")
			res, err = ts.executeRunSQLEvent(ctx, event)
			if err != nil {
				return executor.TxnExecutionResult{}, fmt.Errorf("executing runsql event: %s", err)
			}
		case *ethereum.ContractCreateTable:
			ts.log.Debug().
				Str("owner", event.Owner.Hex()).
				Str("tokenId", event.TableId.String()).
				Str("statement", event.Statement).
				Msgf("executing create-table event")
			res, err = ts.executeCreateTableEvent(ctx, event)
			if err != nil {
				return executor.TxnExecutionResult{}, fmt.Errorf("executing create-table event: %s", err)
			}
		case *ethereum.ContractSetController:
			ts.log.Debug().
				Str("controller", event.Controller.Hex()).
				Str("tokenId", event.TableId.String()).
				Msgf("executing set-controller event")
			res, err = ts.executeSetControllerEvent(ctx, event)
			if err != nil {
				return executor.TxnExecutionResult{}, fmt.Errorf("executing set-controller event: %s", err)
			}
		case *ethereum.ContractTransferTable:
			ts.log.Debug().
				Str("from", event.From.Hex()).
				Str("to", event.To.Hex()).
				Str("tableId", event.TableId.String()).
				Msgf("executing table transfer event")

			res, err = ts.executeTransferEvent(ctx, event)
			if err != nil {
				return executor.TxnExecutionResult{}, fmt.Errorf("executing transfer event: %s", err)
			}
		default:
			return executor.TxnExecutionResult{}, fmt.Errorf("unknown event type %t", event)
		}

		// If the current event fail, we stop processing further events in this transaction and already
		// return the failed receipt. This receipt contains the index of this failed event.
		if res.Error != nil {
			return res, nil
		}
	}

	return res, nil
}
