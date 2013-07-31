package main

import (
	"bufio"
	"fmt"
	"github.com/op/go-logging"
	"strings"
)

type FilterFunc func(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error)

type Dispatcher struct {
	Core
	*logging.Logger
}

func NewDispatcher(connections <-chan Connection, core Core) (d Dispatcher) {
	d = Dispatcher{core, logging.MustGetLogger(LOGGER_NAME)}
	go d.dispatcherFunc(connections)
	return
}

func (d *Dispatcher) dispatcherFunc(connections <-chan Connection) {
	for connection := range connections {
		go func(conn Connection) {
			defer conn.Close()

			session := NewSession(d.Core, conn)

			// Send welcome message
			conn.Write([]byte(fmt.Sprintf("OK MPD %s\n", PROTOCOL_VERSION)))

			reader := bufio.NewScanner(conn)
			for reader.Scan() {
				req := reader.Text()
				d.Info("--> %s", req)
				resp, err := d.HandleRequest(&session, req, 0)
				for _, line := range resp {
					if err != nil {
						d.Warning("Command failed: %v", err)
					}

					d.Info("<-- %s", line)
					conn.Write(append([]byte(line), '\n'))
				}
			}
		}(connection)
	}
}

func (d *Dispatcher) HandleRequest(session *Session, req string, curCommandListIdx int) ([]string, error) {
	d.Debug("HandleRequest: %s", req)
	session.commandListIndex = curCommandListIdx

	response := []string{}
	filterChain := []FilterFunc{
		d.CatchMpdAckErrorsFilter,
		d.AuthenticateFilter,
		d.CommandListFilter,
		d.IdleFilter,
		d.AddOkFilter,
		d.CallHandlerFilter,
	}

	return d.CallNextFilter(session, req, response, filterChain)
}

func (d *Dispatcher) CallNextFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	if len(filterChain) > 0 {
		nextFilter := filterChain[0]
		return nextFilter(session, req, resp, filterChain[1:])
	} else {
		return resp, nil
	}
}

func (d *Dispatcher) CatchMpdAckErrorsFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("CatchMpdAckErrorsFilter")
	resp, err := d.CallNextFilter(session, req, resp, filterChain)

	if err != nil {
		ackErr := err.(MpdAckError)
		if session.commandListIndex < 0 {
			ackErr.Index = session.commandListIndex
		}

		resp = []string{ackErr.AckString()}
	}

	return resp, err
}

func (d *Dispatcher) AuthenticateFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("AuthenticateFilter")

	if session.Authenticated {
		return d.CallNextFilter(session, req, resp, filterChain)
	} else if d.Core.Config.Mpd.Password == "" {
		session.Authenticated = true
		return d.CallNextFilter(session, req, resp, filterChain)
	} else {
		commandName := strings.Split(req, " ")[0]

		if command, ok := MPD_COMMANDS[commandName]; ok && !command.AuthRequired {
			return d.CallNextFilter(session, req, resp, filterChain)
		} else {
			return nil, MpdAckError{
				Code:    ACK_ERROR_PERMISSION,
				Command: commandName,
				Message: fmt.Sprintf("you don't have permission for \"%s\"", commandName),
			}
		}
	}
}

func (d *Dispatcher) CommandListFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("CommandListFilter")

	if d.isReceivingCommandList(session, req) {
		d.Debug("CommandListFilter append to command list")
		session.commandList = append(session.commandList, req)
		return []string{}, nil
	} else {
		resp, err := d.CallNextFilter(session, req, resp, filterChain)

		if err != nil {
			return resp, err
		} else if d.isReceivingCommandList(session, req) || d.isProcessingCommandList(session, req) {
			if len(resp) > 0 && resp[len(resp)-1] == "OK" {
				resp = resp[:len(resp)-1]
			}
		}

		return resp, nil
	}
}

func (d *Dispatcher) isReceivingCommandList(session *Session, req string) bool {
	return session.commandListReceiving && req != "command_list_end"
}

func (d *Dispatcher) isProcessingCommandList(session *Session, req string) bool {
	return session.commandListIndex != 0 && req != "command_list_end"
}

func (d *Dispatcher) AddOkFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("AddOkFilter")

	resp, err := d.CallNextFilter(session, req, resp, filterChain)

	if err != nil {
		return resp, err
	} else if !d.hasError(resp) {
		resp = append(resp, "OK")
	}

	return resp, nil
}

func (d *Dispatcher) hasError(resp []string) bool {
	return resp != nil && (len(resp) > 0 && strings.HasPrefix(resp[len(resp)-1], "ACK"))
}

func (d *Dispatcher) CallHandlerFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("CallHandlerFilter")

	cmd, params, err := d.findMpdCommand(req)
	if err != nil {
		return []string{}, err
	}

	handlerResponse, err := cmd.Handler(session, params)
	if err != nil {
		return []string{}, err
	}

	return d.CallNextFilter(session, req, handlerResponse, filterChain)
}

func (d *Dispatcher) findMpdCommand(req string) (*MpdCommand, map[string]string, error) {
	commandName := strings.Split(req, " ")[0]
	mpdCommand, ok := MPD_COMMANDS[commandName]

	if ok {
		matches := mpdCommand.Pattern.FindStringSubmatch(req)

		if matches == nil {
			return &mpdCommand, nil, MpdAckError{
				Code:    ACK_ERROR_ARG,
				Command: commandName,
				Message: "incorrect arguments",
			}
		}

		groups := mpdCommand.Pattern.SubexpNames()[1:]
		params := toMap(groups, matches[1:])

		d.Debug("COMMAND:%s ARGUMENTS:%q", commandName, params)

		return &mpdCommand, params, nil
	}

	return nil, nil, MpdAckError{
		Code:    ACK_ERROR_UNKNOWN,
		Command: commandName,
		Message: fmt.Sprintf("unknown command \"%s\"", commandName),
	}
}

func toMap(groups []string, matches []string) (params map[string]string) {
	params = map[string]string{}
	for i, group := range groups {
		if group != "" {
			params[group] = matches[i]
		}
	}

	return
}

func (d *Dispatcher) IdleFilter(session *Session, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("IdleFilter")
	noidle := "noidle"

	if session.isCurrentlyIdle() && req != noidle {
		d.Debug("Client sent us %s, only %s is allowed while in the idle state", req, noidle)
		session.Close()
		return nil, nil
	}

	if !session.isCurrentlyIdle() && req == noidle {
		// noidle was called before idle
		return nil, nil
	}

	resp, err := d.CallNextFilter(session, req, resp, filterChain)

	if session.isCurrentlyIdle() {
		return nil, nil
	}

	return resp, err
}
