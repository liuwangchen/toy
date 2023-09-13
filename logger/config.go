// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package logger

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type LogProperty struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type LogFilter struct {
	Disable  bool          `yaml:"disable"`
	Tag      string        `yaml:"tag"`
	Level    string        `yaml:"level"`
	Type     string        `yaml:"type"`
	Property []LogProperty `yaml:"property"`
}

type LogConfig struct {
	Filter []LogFilter `yaml:"filter"`
}

// Load configuration; see examples/example.xml for documentation
func (log Logger) LoadConfigurationFromFile(fileName string) error {
	// Open the configuration file
	fd, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("LoadConfiguration: Error: Could not open %q for reading: %s\n", fileName, err)
	}

	content, err := ioutil.ReadAll(fd)
	if err != nil {
		return fmt.Errorf("LoadConfiguration: Error: Could not read %q: %s\n", fileName, err)
	}
	return log.LoadConfigurationFromContent(content)

}

// Load configuration; see examples/example.xml for documentation
func (log Logger) LoadConfigurationFromContent(contents []byte) error {
	if len(contents) == 0 {
		return fmt.Errorf("LoadConfiguration: Error: content empty")
	}

	xc := new(LogConfig)
	if err := yaml.Unmarshal(contents, xc); err != nil {
		return fmt.Errorf("LoadConfiguration: Error: Could not parse XML configuration %s\n", err)
	}
	return log.LoadConfiguration(xc)
}

// Load configuration; see examples/example.xml for documentation
func (log Logger) LoadConfiguration(xc *LogConfig) error {
	if xc == nil {
		return fmt.Errorf("LoadConfiguration: Error: config nil")
	}

	log.Close()

	for _, filter := range xc.Filter {
		if filter.Disable {
			continue
		}
		var lw LogWriter
		var lvl level

		if len(filter.Tag) == 0 {
			return fmt.Errorf("LoadConfiguration: Error: Required child <%s> for filter missing\n", "tag")
		}
		if len(filter.Type) == 0 {
			return fmt.Errorf("LoadConfiguration: Error: Required child <%s> for filter missing\n", "type")
		}
		if len(filter.Level) == 0 {
			return fmt.Errorf("LoadConfiguration: Error: Required child <%s> for filter missing\n", "level")
		}

		switch filter.Level {
		case "FINEST":
			lvl = FINEST
		case "FINE":
			lvl = FINE
		case "DEBUG":
			lvl = DEBUG
		case "TRACE":
			lvl = TRACE
		case "INFO":
			lvl = INFO
		case "WARNING":
			lvl = WARNING
		case "ERROR":
			lvl = ERROR
		case "FATAL":
			lvl = FATAL
		default:
			return fmt.Errorf("LoadConfiguration: Error: Required child <%s> for filter has unknown value : %s\n", "level", filter.Level)
		}

		var err error

		switch filter.Type {
		case "console":
			lw, err = toConsoleLogWriter(filter.Property)
		case "file":
			lw, err = toFileLogWriter(filter.Property)
		case "socket":
			lw, err = toSocketLogWriter(filter.Property)
		case "kafka":
			lw, err = toProducerLogWriter(filter.Type, filter.Property)
		default:
			return fmt.Errorf("LoadConfiguration: Error: Could not load XML configuration: unknown filter type \"%s\"\n", filter.Type)
		}

		if err != nil {
			return err
		}

		log[filter.Tag] = &Filter{lvl, lw}
	}
	return nil
}

func toConsoleLogWriter(props []LogProperty) (ConsoleLogWriter, error) {
	var (
		color bool
	)
	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "color":
			color, _ = strconv.ParseBool(prop.Value)
		default:
			return nil, fmt.Errorf("LoadConfiguration: Warning: Unknown property \"%s\" for console filter\n", prop.Name)
		}
	}

	if color {
		return NewColorConsoleLogWriter(), nil
	}

	return NewConsoleLogWriter(), nil
}

// Parse a number with K/M/G suffixes based on thousands (1000) or 2^10 (1024)
func strToNumSuffix(str string, mult int) int {
	num := 1
	if len(str) > 1 {
		switch str[len(str)-1] {
		case 'G', 'g':
			num *= mult
			fallthrough
		case 'M', 'm':
			num *= mult
			fallthrough
		case 'K', 'k':
			num *= mult
			str = str[0 : len(str)-1]
		}
	}
	parsed, _ := strconv.Atoi(str)
	return parsed * num
}

func toFileLogWriter(props []LogProperty) (*FileLogWriter, error) {
	file := ""
	format := "[%D %T] [%L] (%S) %M"
	maxlines := 0
	maxsize := 0
	daily := false
	hour := false
	rotate := false

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "filename":
			file = strings.Trim(prop.Value, " \r\n")
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
		case "maxlines":
			maxlines = strToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1000)
		case "maxsize":
			maxsize = strToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1024)
		case "daily":
			daily = strings.Trim(prop.Value, " \r\n") != "false"
		case "hour":
			hour = strings.Trim(prop.Value, " \r\n") != "false"
		case "rotate":
			rotate = strings.Trim(prop.Value, " \r\n") != "false"
		default:
			return nil, fmt.Errorf("LoadConfiguration: Warning: Unknown property \"%s\" for file filter\n", prop.Name)
		}
	}

	// Check properties
	if len(file) == 0 {
		return nil, fmt.Errorf("LoadConfiguration: Error: Required property \"%s\" for file filter missing\n", "filename")
	}

	flw, err := NewFileLogWriter(file, rotate)
	if err != nil {
		return nil, err
	}
	flw.SetFormat(format)
	flw.SetRotateLines(maxlines)
	flw.SetRotateSize(maxsize)
	flw.SetRotateDaily(daily)
	flw.SetRotateHour(hour)
	return flw, nil
}

func toSocketLogWriter(props []LogProperty) (SocketLogWriter, error) {
	endpoint := ""
	protocol := "udp"

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "endpoint":
			endpoint = strings.Trim(prop.Value, " \r\n")
		case "protocol":
			protocol = strings.Trim(prop.Value, " \r\n")
		default:
			return nil, fmt.Errorf("LoadConfiguration: Warning: Unknown property \"%s\" for file filter\n", prop.Name)
		}
	}

	// Check properties
	if len(endpoint) == 0 {
		return nil, fmt.Errorf("LoadConfiguration: Error: Required property \"%s\" for file filter missing\n", "endpoint")
	}

	return NewSocketLogWriter(protocol, endpoint), nil
}
