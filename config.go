package checker

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// CheckerConfig holds the configuration for a Checker instance.
type CheckerConfig struct {
	Interval  int
	Server    ServerConfig
	Checks    []CheckConfig
	Notifiers []NotifierConfig
}

// CheckConfig defines the configuration for a specific check.
type CheckConfig struct {
	Maker string
	Name  string
	Args  any
}

// NotifierConfig defines the configuration for a specific notifier.
type NotifierConfig struct {
	Maker string
	Args  any
}

// ServerConfig definies the configuration for the server for peer-to-peer monitoring
type ServerConfig struct {
	Enabled bool
	Listen  string
}

// CheckMaker is an interface for creating checks from configuration.
type CheckMaker interface {
	Maker() string
	FromConfig(c CheckConfig) (Check, error)
	UnmarshalArgs(j json.RawMessage) (any, error)
}

// NotifierMaker is an interface for creating notifiers from configuration.
type NotifierMaker interface {
	Maker() string
	FromConfig(c NotifierConfig) (Notifier, error)
	UnmarshalArgs(j json.RawMessage) (any, error)
}

// ArgsUnmarshaler is an interface for unmarshaling JSON arguments.
type ArgsUnmarshaler interface {
	Maker() string
	UnmarshalArgs(j json.RawMessage) (any, error)
}

// WithRecursion provides recursive parsing capabilities for nested configurations.
type WithRecursion struct {
}

// CheckRecursion evaluates a nested CheckConfig into a Check.
func (wr *WithRecursion) CheckRecursion(cc CheckConfig) (Check, error) {
	return checkFromConfig(cc)
}

// NotifierRecursion evaluates a nested NotifierConfig into a Notifier.
func (wr *WithRecursion) NotifierRecursion(nc NotifierConfig) (Notifier, error) {
	return notifierFromConfig(nc)
}

var checkMakers = make(map[string]CheckMaker)
var notifierMakers = make(map[string]NotifierMaker)
var argsUnmarshaler = make(map[string]ArgsUnmarshaler)

// AddCheckMaker registers one or more CheckMaker instances.
func AddCheckMaker(ms ...CheckMaker) error {
	for _, m := range ms {
		if _, ok := checkMakers[m.Maker()]; ok {
			return fmt.Errorf("can't add maker. Maker %s already exists", m.Maker())
		}
		checkMakers[m.Maker()] = m
		argsUnmarshaler[m.Maker()] = m
	}
	return nil
}

// AddNotifierMaker registers one or more NotifierMaker instances.
func AddNotifierMaker(ms ...NotifierMaker) error {
	for _, m := range ms {
		if _, ok := notifierMakers[m.Maker()]; ok {
			return fmt.Errorf("can't add maker. Maker %s already exists", m.Maker())
		}
		notifierMakers[m.Maker()] = m
		argsUnmarshaler[m.Maker()] = m
	}
	return nil
}

// ReadConfig loads the Checker configuration from the provided io.Reader.
func (chkr *Checker) ReadConfig(r io.Reader) error {
	dec := json.NewDecoder(r)
	conf := CheckerConfig{}
	if err := dec.Decode(&conf); err != nil {
		return fmt.Errorf("can't decode config: %v", err)
	}
	chkr.serverConfig = conf.Server
	chkr.interval = time.Duration(conf.Interval) * time.Second
	for _, cc := range conf.Checks {
		chk, err := checkFromConfig(cc)
		if err != nil {
			return err
		}
		chkr.AddCheck(cc.Name, chk)
	}
	for _, nc := range conf.Notifiers {
		not, err := notifierFromConfig(nc)
		if err != nil {
			return err
		}
		chkr.AddNotifier(not)
	}
	return nil
}

func checkFromConfig(cc CheckConfig) (Check, error) {
	cm, ok := checkMakers[cc.Maker]
	if !ok {
		return nil, fmt.Errorf("no check maker '%s'", cc.Maker)
	}
	chk, err := cm.FromConfig(cc)
	if err != nil {
		return nil, fmt.Errorf("can't make check %s, maker %s from config: %v", cc.Name, cc.Maker, err)
	}
	return chk, nil
}

func notifierFromConfig(nc NotifierConfig) (Notifier, error) {
	nm, ok := notifierMakers[nc.Maker]
	if !ok {
		return nil, fmt.Errorf("no notifier maker '%s'", nc.Maker)
	}
	not, err := nm.FromConfig(nc)
	if err != nil {
		return nil, fmt.Errorf("can't make notifier %s from config: %v", nc.Maker, err)
	}
	return not, nil
}

// UnmarshalJSON customizes the JSON unmarshaling for CheckConfig.
func (c *CheckConfig) UnmarshalJSON(b []byte) error {
	cc := struct {
		Maker string
		Name  string
		Args  json.RawMessage
	}{}
	if err := json.Unmarshal(b, &cc); err != nil {
		return fmt.Errorf("can't unmarshal Check config: %v", err)
	}
	c.Maker = cc.Maker
	c.Name = cc.Name
	unm, ok := argsUnmarshaler[c.Maker]
	if !ok {
		return fmt.Errorf("arguments unmarshaller '%s' unknown", c.Maker)
	}
	args, err := unm.UnmarshalArgs(cc.Args)
	if err != nil {
		return fmt.Errorf("can't unmarshal arguments for Maker %s, Check %s: %v", c.Maker, c.Name, err)
	}
	c.Args = args
	return nil
}

// UnmarshalJSON customizes the JSON unmarshaling for NotifierConfig.
func (n *NotifierConfig) UnmarshalJSON(b []byte) error {
	nc := struct {
		Maker string
		Args  json.RawMessage
	}{}
	if err := json.Unmarshal(b, &nc); err != nil {
		return fmt.Errorf("can't unmarshal Check config: %v", err)
	}
	n.Maker = nc.Maker
	unm, ok := argsUnmarshaler[n.Maker]
	if !ok {
		return fmt.Errorf("arguments unmarshaller '%s' unknown", n.Maker)
	}
	args, err := unm.UnmarshalArgs(nc.Args)
	if err != nil {
		return fmt.Errorf("can't unmarshal arguments for Maker %s: %v", n.Maker, err)
	}
	n.Args = args
	return nil
}
