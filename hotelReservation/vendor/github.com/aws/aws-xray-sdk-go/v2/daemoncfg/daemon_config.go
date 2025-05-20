// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package daemoncfg

import (
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-xray-sdk-go/v2/internal/logger"
	"github.com/pkg/errors"
)

var addressDelimiter = " " // delimiter between tcp and udp addresses
var udpKey = "udp"
var tcpKey = "tcp"

// DaemonEndpoints stores X-Ray daemon configuration about the ip address and port for UDP and TCP port. It gets the address
// string from "AWS_TRACING_DAEMON_ADDRESS" and then from recorder's configuration for DaemonAddr.
// A notation of '127.0.0.1:2000' or 'tcp:127.0.0.1:2000 udp:127.0.0.2:2001' or 'udp:127.0.0.1:2000 tcp:127.0.0.2:2001'
// are both acceptable. The first one means UDP and TCP are running at the same address.
// Notation 'hostname:2000' or 'tcp:hostname:2000 udp:hostname:2001' or 'udp:hostname:2000 tcp:hostname:2001' are also acceptable.
// By default it assumes a X-Ray daemon running at 127.0.0.1:2000 listening to both UDP and TCP traffic.
type DaemonEndpoints struct {
	// UDPAddr represents UDP endpoint for segments to be sent by emitter.
	UDPAddr *net.UDPAddr
	// TCPAddr represents TCP endpoint of the daemon to make sampling API calls.
	TCPAddr *net.TCPAddr
}

// GetDaemonEndpoints returns DaemonEndpoints.
func GetDaemonEndpoints() *DaemonEndpoints {
	daemonEndpoint, err := GetDaemonEndpointsFromString("") // only environment variable would be parsed

	if err != nil {
		panic(err)
	}

	if daemonEndpoint == nil { // env variable not set
		return GetDefaultDaemonEndpoints()
	}
	return daemonEndpoint // env variable successfully parsed
}

// GetDaemonEndpointsFromEnv resolves the daemon address if set in the environment variable.
func GetDaemonEndpointsFromEnv() (*DaemonEndpoints, error) {
	if envDaemonAddr := os.Getenv("AWS_XRAY_DAEMON_ADDRESS"); envDaemonAddr != "" {
		return resolveAddress(envDaemonAddr)
	}
	return nil, nil
}

// GetDefaultDaemonEndpoints returns the default UDP and TCP address of the daemon.
func GetDefaultDaemonEndpoints() *DaemonEndpoints {
	udpAddr := &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 2000,
	}

	tcpAddr := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 2000,
	}

	return &DaemonEndpoints{
		UDPAddr: udpAddr,
		TCPAddr: tcpAddr,
	}
}

// GetDaemonEndpointsFromString parses provided daemon address if the environment variable is invalid or not set.
// DaemonEndpoints is non nil if the env variable or provided address is valid.
func GetDaemonEndpointsFromString(dAddr string) (*DaemonEndpoints, error) {
	var daemonAddr string
	// Try to get the X-Ray daemon address from an environment variable
	if envDaemonAddr := os.Getenv("AWS_XRAY_DAEMON_ADDRESS"); envDaemonAddr != "" {
		daemonAddr = envDaemonAddr
		logger.Infof("using daemon endpoints from environment variable AWS_XRAY_DAEMON_ADDRESS: %v", envDaemonAddr)
	} else if dAddr != "" {
		daemonAddr = dAddr
	}
	if daemonAddr != "" {
		return resolveAddress(daemonAddr)
	}
	return nil, nil
}

func resolveAddress(dAddr string) (*DaemonEndpoints, error) {
	addr := strings.Split(dAddr, addressDelimiter)
	switch len(addr) {
	case 1:
		return parseSingleForm(addr[0])
	case 2:
		return parseDoubleForm(addr)
	}
	return nil, errors.New("invalid daemon address: " + dAddr)
}

func parseDoubleForm(addr []string) (*DaemonEndpoints, error) {
	addr1 := strings.Split(addr[0], ":") // tcp:127.0.0.1:2000  or udp:127.0.0.1:2000
	addr2 := strings.Split(addr[1], ":") // tcp:127.0.0.1:2000  or udp:127.0.0.1:2000

	if len(addr1) != 3 || len(addr2) != 3 {
		return nil, errors.New("invalid daemon address: " + addr[0] + " " + addr[1])
	}

	// validate ports
	_, pErr1 := strconv.Atoi(addr1[2])
	_, pErr2 := strconv.Atoi(addr1[2])

	if pErr1 != nil || pErr2 != nil {
		return nil, errors.New("invalid daemon address port")
	}

	addrMap := make(map[string]string)

	addrMap[addr1[0]] = addr1[1] + ":" + addr1[2]
	addrMap[addr2[0]] = addr2[1] + ":" + addr2[2]

	if addrMap[udpKey] == "" || addrMap[tcpKey] == "" { // for double form, tcp and udp keywords should be present
		return nil, errors.New("invalid daemon address")
	}

	udpAddr, uErr := resolveUDPAddr(addrMap[udpKey])
	if uErr != nil {
		return nil, uErr
	}

	tcpAddr, tErr := resolveTCPAddr(addrMap[tcpKey])
	if tErr != nil {
		return nil, tErr
	}

	return &DaemonEndpoints{
		UDPAddr: udpAddr,
		TCPAddr: tcpAddr,
	}, nil
}

func parseSingleForm(addr string) (*DaemonEndpoints, error) { // format = "ip:port"
	a := strings.Split(addr, ":") // 127.0.0.1:2000

	if len(a) != 2 {
		return nil, errors.New("invalid daemon address: " + addr)
	}

	// validate port
	_, pErr1 := strconv.Atoi(a[1])

	if pErr1 != nil {
		return nil, errors.New("invalid daemon address port")
	}

	udpAddr, uErr := resolveUDPAddr(addr)
	if uErr != nil {
		return nil, uErr
	}
	tcpAddr, tErr := resolveTCPAddr(addr)
	if tErr != nil {
		return nil, tErr
	}

	return &DaemonEndpoints{
		UDPAddr: udpAddr,
		TCPAddr: tcpAddr,
	}, nil
}

func resolveUDPAddr(s string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr(udpKey, s)
}

func resolveTCPAddr(s string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr(tcpKey, s)
}
