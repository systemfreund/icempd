package main

import (
	"bufio"
	"fmt"
	"strings"
	"github.com/op/go-logging"
)

type FilterFunc func(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error)

type Dispatcher struct {
	config Configuration
	*logging.Logger
}

func NewDispatcher(config Configuration, sessions <-chan MpdSession) (d Dispatcher) {
	d = Dispatcher{config, logging.MustGetLogger(LOGGER_NAME)}
	go d.dispatcherFunc(sessions)
	return
}

func (d *Dispatcher) dispatcherFunc(sessions <-chan MpdSession) {
	for session := range sessions {
		go func(session MpdSession) {
			defer session.Close()

			// Send welcome message
			session.Write([]byte(fmt.Sprintf("OK MPD %s\n", PROTOCOL_VERSION)))

			reader := bufio.NewScanner(session)
			for reader.Scan() {
				req := reader.Text()
				d.Info("--> %s", req)
				resp, _ := d.HandleRequest(&session, req, 0)
				for _, line := range resp {
					d.Info("<-- %s", line)	
					session.Write(append([]byte(line), '\n'))
				}
			}
		}(session)
	}
}

func (d *Dispatcher) HandleRequest(session *MpdSession, req string, curCommandListIdx int) ([]string, error) {
	d.Debug("HandleRequest: %s", req)
	session.commandListIndex = curCommandListIdx;

	response := []string{};
	filterChain := []FilterFunc { 
		d.CatchMpdAckErrorsFilter, 
		d.AuthenticateFilter,
		d.CommandListFilter,
		d.AddOkFilter,
		d.CallHandlerFilter,
	}

	return d.CallNextFilter(session, req, response, filterChain)
}

func (d *Dispatcher) CallNextFilter(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	if len(filterChain) > 0 {
		nextFilter := filterChain[0]
		return nextFilter(session, req, resp, filterChain[1:])
	} else {
		return resp, nil
	}
}

func (d *Dispatcher) CatchMpdAckErrorsFilter(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("CatchMpdAckErrorsFilter")
	resp, err := d.CallNextFilter(session, req, resp, filterChain)

	if err != nil {
		ackErr := err.(MpdAckError)
		if session.commandListIndex < 0 {
			ackErr.Index = session.commandListIndex
		}

		resp = []string { ackErr.AckString() }
	}

	return resp, err
}

func (d *Dispatcher) AuthenticateFilter(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("AuthenticateFilter")

	if session.Authenticated {
		return d.CallNextFilter(session, req, resp, filterChain)
	} else if d.config.Mpd.Password == "" {
		session.Authenticated = true
		return d.CallNextFilter(session, req, resp, filterChain)
	} else {
		commandName := strings.Split(req, " ")[0]
		
		if command, ok := MPD_COMMANDS[commandName]; ok && !command.AuthRequired {
			return d.CallNextFilter(session, req, resp, filterChain)
		} else {
			return nil, MpdAckError{
				Code: ACK_ERROR_PERMISSION,
				Command: commandName,
				Message: fmt.Sprintf("you don't have permission for \"%s\"", commandName),
			}
		}
	}
}

func (d *Dispatcher) CommandListFilter(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
	d.Debug("CommandListFilter")

	if d.isReceivingCommandList(session, req) {
		d.Debug("CommandListFilter append to command list")
		session.commandList = append(session.commandList, req)
		return []string {}, nil
	} else {
		resp, err := d.CallNextFilter(session, req, resp, filterChain)
		
		if err != nil {
			return resp, err
		} else if d.isReceivingCommandList(session, req) || d.isProcessingCommandList(session, req) {
			if len(resp) > 0 && resp[len(resp) - 1] == "OK" {
				resp = resp[:len(resp) - 1]
			}
		}

		return resp, nil
	}
}

func (d *Dispatcher) isReceivingCommandList(session *MpdSession, req string) bool {
	return session.commandListReceiving && req != "command_list_end"
}

func (d *Dispatcher) isProcessingCommandList(session *MpdSession, req string) bool {
	return session.commandListIndex != 0 && req != "command_list_end"
}

func (d *Dispatcher) AddOkFilter(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
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
	return resp != nil && (len(resp) > 0 && strings.HasPrefix(resp[len(resp) - 1], "ACK"))
}

func (d *Dispatcher) CallHandlerFilter(session *MpdSession, req string, resp []string, filterChain []FilterFunc) ([]string, error) {
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
		
		d.Debug("COMMAND:%s ARGUMENTS:%q", commandName, params)

		return &mpdCommand, params, nil
	}

	return nil, nil, MpdAckError{
		Code: ACK_ERROR_UNKNOWN,
		Command: commandName,
		Message: fmt.Sprintf("unknown command \"%s\"", commandName),
	}
}