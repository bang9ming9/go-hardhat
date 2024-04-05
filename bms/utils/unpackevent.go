package bmsutils

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

var (
	ErrEventNotFind error = errors.New("could not found event")
)

func UnpackEvents[T any](aBI *abi.ABI, name string, receipts ...*types.Receipt) ([]*T, error) {
	logs := make([]*types.Log, 0)
	for _, receipt := range receipts {
		logs = append(logs, receipt.Logs...)
	}

	e, elogs, err := FindEventLogWithABI(logs, aBI, name)
	if err != nil {
		return nil, err
	}

	events := make([]*T, 0)
	for _, log := range elogs {
		event := new(T)
		if err := aBI.UnpackIntoInterface(event, name, log.Data); err != nil {
			return nil, errors.Wrap(err, "aBI.UnpackIntoInterface")
		} else {
			var indexed abi.Arguments
			for _, arg := range e.Inputs {
				if arg.Indexed {
					indexed = append(indexed, arg)
				}
			}
			if err := abi.ParseTopics(event, indexed, log.Topics[1:]); err != nil {
				return nil, errors.Wrap(err, "abi.ParseTopics")
			}
			events = append(events, event)
		}
	}
	return events, nil
}

func UnpackEventsIntoMap(aBI *abi.ABI, name string, receipts ...*types.Receipt) ([]map[string]interface{}, error) {
	logs := make([]*types.Log, 0)
	for _, receipt := range receipts {
		logs = append(logs, receipt.Logs...)
	}

	e, elogs, err := FindEventLogWithABI(logs, aBI, name)
	if err != nil {
		return nil, err
	}

	events := make([]map[string]interface{}, 0)
	for _, log := range elogs {
		event := make(map[string]interface{})
		if err := aBI.UnpackIntoMap(event, name, log.Data); err != nil {
			return nil, errors.Wrap(err, "aBI.UnpackIntoMap")
		} else {
			var indexed abi.Arguments
			for _, arg := range e.Inputs {
				if arg.Indexed {
					indexed = append(indexed, arg)
				}
			}
			if err := abi.ParseTopicsIntoMap(event, indexed, log.Topics[1:]); err != nil {
				return nil, errors.Wrap(err, "abi.ParseTopics")
			}
			events = append(events, event)
		}
	}
	return events, nil
}

func FindEventLog(logs []*types.Log, aBI *abi.ABI, name string) ([]*types.Log, error) {
	_, elogs, err := FindEventLogWithABI(logs, aBI, name)
	return elogs, err
}

func FindEventLogWithABI(logs []*types.Log, aBI *abi.ABI, name string) (abi.Event, []*types.Log, error) {
	e, ok := aBI.Events[name]
	if !ok {
		return abi.Event{}, nil, errors.Wrap(ErrEventNotFind, "name")
	}

	elogs := make([]*types.Log, 0)
	for _, log := range logs {
		if log.Topics[0] == e.ID {
			elogs = append(elogs, log)
		}
	}
	return e, elogs, nil
}
