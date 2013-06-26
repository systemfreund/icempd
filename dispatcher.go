package main

import (
	"container/list"
	"fmt"
	"strings"
)

type Response struct {
	list.List
}

type FilterFunc func(req string, resp *Response, filterChain []FilterFunc) (*Response, error)

type MpdDispatcher struct {
	Config Configuration
	Authenticated bool
	CommandListReceiving bool
	CommandListOk bool
	CommandList list.List
	CommandListIndex int
}

func (d *MpdDispatcher) HandleRequest(req string, curCommandListIdx int) (*Response, error) {
	d.CommandListIndex = curCommandListIdx;

	response := new(Response)
	response.Init()

	filterChain := []FilterFunc { 
		d.CatchMpdAckErrorsFilter, 
		d.AuthenticateFilter,
	}

	return d.CallNextFilter(req, response, filterChain)
}

func (d *MpdDispatcher) CallNextFilter(req string, resp *Response, filterChain []FilterFunc) (*Response, error) {
	if len(filterChain) > 0 {
		nextFilter := filterChain[0]
		return nextFilter(req, resp, filterChain[1:])
	} else {
		return resp, nil
	}
}

func (d *MpdDispatcher) CatchMpdAckErrorsFilter(req string, resp *Response, filterChain []FilterFunc) (*Response, error) {
	logger.Debug("CatchMpdAckErrorsFilter")
	resp, err := d.CallNextFilter(req, resp, filterChain)
	var ackErr MpdAckError

	if err != nil {
		ackErr = err.(MpdAckError)
		if d.CommandListIndex < 0 {
			ackErr.Index = d.CommandListIndex
		}

		// Overwrite with error response
		resp = new(Response)
		resp.PushFront(ackErr.AckString())
	}

	return resp, err
}

func (d *MpdDispatcher) AuthenticateFilter(req string, resp *Response, filterChain []FilterFunc) (*Response, error) {
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