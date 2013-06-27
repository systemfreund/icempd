package main

import (
	"fmt"
	"strings"
)

type FilterFunc func(req string, resp []string, filterChain []FilterFunc) ([]string, error)

type MpdDispatcher struct {
	Session *MpdSession

	Authenticated bool
	commandListReceiving bool
	commandListOk bool
	commandList []string
	commandListIndex int
}

func (d *MpdDispatcher) HandleRequest(req string, curCommandListIdx int) ([]string, error) {
	logger.Debug("HandleRequest: %s", req)
	d.commandListIndex = curCommandListIdx;

	response := []string{};
	filterChain := []FilterFunc { 
		d.CatchMpdAckErrorsFilter, 
		d.AuthenticateFilter,
		d.CommandListFilter,
		d.AddOkFilter,
		d.CallHandlerFilter,
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
		if d.commandListIndex < 0 {
			ackErr.Index = d.commandListIndex
		}

		resp = []string { ackErr.AckString() }
	}

	return resp, err
}

func (d *MpdDispatcher) AuthenticateFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	logger.Debug("AuthenticateFilter")

	if d.Authenticated {
		return d.CallNextFilter(req, resp, filterChain)
	} else if d.Session.Config.Mpd.Password == "" {
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
		d.commandList = append(d.commandList, req)
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
	return d.commandListReceiving && req != "command_list_end"
}

func (d *MpdDispatcher) isProcessingCommandList(req string) bool {
	return d.commandListIndex != 0 && req != "command_list_end"
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
	return resp != nil && (len(resp) > 0 && strings.HasPrefix(resp[len(resp) - 1], "ACK"))
}

func (d *MpdDispatcher) CallHandlerFilter(req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	logger.Debug("CallHandlerFilter")

	cmd, params, err := d.findMpdCommand(req)
	if err != nil {
		return []string{}, err
	}

	handlerResponse, err := cmd.Handler(d.Session, params)
	if err != nil {
		return []string{}, err
	}

	return d.CallNextFilter(req, handlerResponse, filterChain)
}

func (d *MpdDispatcher) findMpdCommand(req string) (*MpdCommand, map[string]string, error) {
	commandName := strings.Split(req, " ")[0]
	mpdCommand, ok := MPD_COMMANDS[commandName]

	if ok {
		matches := mpdCommand.Pattern.FindStringSubmatch(req)

		if matches == nil {
			return &mpdCommand, nil, MpdAckError{
				Code: ACK_ERROR_ARG,
				Command: commandName,
				Message: "incorrect arguments",
			}
		}

		// Convert to parameter map
		params := map[string]string {}
		groups := mpdCommand.Pattern.SubexpNames()[1:]
		for i, group := range groups {
			params[group] = matches[i + 1]
		}
		
		logger.Debug("COMMAND:%s ARGUMENTS:%q", commandName, params)

		return &mpdCommand, params, nil
	}

	return nil, nil, MpdAckError{
		Code: ACK_ERROR_UNKNOWN,
		Command: commandName,
		Message: fmt.Sprintf("unknown command \"%s\"", commandName),
	}
}