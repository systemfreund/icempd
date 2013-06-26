package main

import (
	"container/list"
)

type Request string

type Response struct {
	list.List
}

type FilterFunc func(req *Request, resp *Response, filterChain []FilterFunc) *Response

type MpdDispatcher struct {
	authenticated bool
	CommandListReceiving bool
	CommandListOk bool
	CommandList list.List
	CommandListIndex int
}

func (d *MpdDispatcher) HandleRequest(req *Request, curCommandListIdx int) *Response {
	d.CommandListIndex = curCommandListIdx;

	response := new(Response)
	response.Init()

	filterChain := []FilterFunc { 
		d.CatchMpdAckErrorsFilter, 
		d.AuthenticateFilter,
	}
	return d.CallNextFilter(req, response, filterChain)
}

func (d *MpdDispatcher) CallNextFilter(req *Request, resp *Response, filterChain []FilterFunc) *Response {
	if len(filterChain) > 0 {
		nextFilter := filterChain[0]
		return nextFilter(req, resp, filterChain[1:])
	} else {
		return resp
	}
}

func (d *MpdDispatcher) CatchMpdAckErrorsFilter(req *Request, resp *Response, filterChain []FilterFunc) *Response {
	logger.Debug("CatchMpdAckErrorsFilter")
	return d.CallNextFilter(req, resp, filterChain)
}

func (d *MpdDispatcher) AuthenticateFilter(req *Request, resp *Response, filterChain []FilterFunc) *Response {
	logger.Debug("AuthenticateFilter")
	return d.CallNextFilter(req, resp, filterChain)
}