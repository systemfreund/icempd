package main

import (
	"fmt"
	"strings"
)

type FilterFunc func(req string, resp []string, filterChain []FilterFunc) ([]string, error)

type MpdDispatcher struct {
	Config Configuration
	Authenticated bool
	CommandListReceiving bool
	CommandListOk bool
	CommandList []string
	CommandListIndex int
}

func (d *MpdDispatcher) HandleRequest(req string, curCommandListIdx int) ([]string, error) {
	d.CommandListIndex = curCommandListIdx;

	response := []string {};

	filterChain := []FilterFunc { 
		d.CatchMpdAckErrorsFilter, 
		d.AuthenticateFilter,
		d.CommandListFilter,
		d.AddOkFilter,
	}

	return d.CallNextFilter(req, response, filterChain)
}

func (d *MpdDispatcher) CallNextFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	if len(filterChain) > 0 {
		nextFilter := filterChain[0]
		return nextFilter(req, resp, filterChain[1:])
	} else {
		return resp, nil
	}
}

func (d *MpdDispatcher) CatchMpdAckErrorsFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	logger.Debug("CatchMpdAckErrorsFilter")
	resp, err := d.CallNextFilter(req, resp, filterChain)

	if err != nil {
		ackErr := err.(MpdAckError)
		if d.CommandListIndex < 0 {
			ackErr.Index = d.CommandListIndex
		}

		resp = []string { ackErr.AckString() }
	}

	return resp, err
}

func (d *MpdDispatcher) AuthenticateFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	logger.Debug("AuthenticateFilter")

	if d.Authenticated {
		return d.CallNextFilter(req, resp, filterChain)
	} else if d.Config.Mpd.Password == "" {
		d.Authenticated = true
		return d.CallNextFilter(req, resp, filterChain)
	} else {
		commandName := strings.Split(req, " ")[0]
		
		if command, ok := MPD_COMMANDS[commandName]; ok && !command.AuthRequired {
			return d.CallNextFilter(req, resp, filterChain)
		} else {
			return nil, MpdAckError{
				Code: ACK_ERROR_PERMISSION,
				Command: commandName,
				Message: fmt.Sprintf("you don't have permission for \"%s\"", commandName),
			}
		}
	}
}

func (d *MpdDispatcher) CommandListFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	logger.Debug("CommandListFilter")

	if d.isReceivingCommandList(req) {
		logger.Debug("CommandListFilter append to command list")
		d.CommandList = append(d.CommandList, req)
		return []string {}, nil
	} else {
		resp, err := d.CallNextFilter(req, resp, filterChain)
		
		if err != nil {
			return resp, err
		} else if d.isReceivingCommandList(req) || d.isProcessingCommandList(req) {
			if len(resp) > 0 && resp[len(resp) - 1] == "OK" {
				resp = resp[:len(resp) - 1]
			}
		}

		return resp, nil
	}
}

func (d *MpdDispatcher) isReceivingCommandList(req string) bool {
	return d.CommandListReceiving && req != "command_list_end"
}

func (d *MpdDispatcher) isProcessingCommandList(req string) bool {
	return d.CommandListIndex != -1 && req != "command_list_end"
}

func (d *MpdDispatcher) AddOkFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	logger.Debug("AddOkFilter")

	resp, err := d.CallNextFilter(req, resp, filterChain)

	if err != nil {
		return resp, err
	} else if !d.hasError(resp) {
		resp = append(resp, "OK")
	}
	
	return resp, nil
}

func (d *MpdDispatcher) hasError(resp []string) bool {
	return len(resp) > 0 && strings.HasPrefix(resp[len(resp) - 1], "ACK")
}