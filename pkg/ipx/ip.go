package ipx

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Ascii numbers 0-9
const (
	ascii_0 = 48
	ascii_9 = 57
)

func Uint64ToNetAddr(value uint64) string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", (value>>32)&0xff, (value>>40)&0xff, (value>>48)&0xff, (value>>56)&0xff, value&0xffffffff)
}

func NetAddrToUint64(strHost string) uint64 {
	var dwIp uint32
	var dwPort uint32

	slcIpPort := strings.Split(strHost, ":")
	if len(slcIpPort) != 2 {
		return ((uint64(dwIp)) << 32) + uint64(dwPort)
	}

	nPort, _ := strconv.Atoi(slcIpPort[1])
	dwPort = uint32(nPort)

	ips := strings.Split(slcIpPort[0], ".")
	if len(ips) != 4 {
		return ((uint64(dwIp)) << 32) + uint64(dwPort)
	}

	b0, _ := strconv.Atoi(ips[3])
	b1, _ := strconv.Atoi(ips[2])
	b2, _ := strconv.Atoi(ips[1])
	b3, _ := strconv.Atoi(ips[0])
	dwIp += uint32(b0) << 24
	dwIp += uint32(b1) << 16
	dwIp += uint32(b2) << 8
	dwIp += uint32(b3)
	return ((uint64(dwIp)) << 32) + uint64(dwPort)
}

func ParseUint64(d []byte) (uint64, bool) {
	var n uint64
	d_len := len(d)
	if d_len == 0 {
		return 0, false
	}
	for i := 0; i < d_len; i++ {
		j := d[i]
		if j < ascii_0 || j > ascii_9 {
			return 0, false
		}
		n = n*10 + (uint64(j - ascii_0))
	}
	return n, true
}

func IpV4ToUint32(ip string) uint32 {
	var n uint32
	ips := strings.Split(ip, ".")
	if len(ips) != 4 {
		return n
	}
	b0, _ := strconv.Atoi(ips[0])
	b1, _ := strconv.Atoi(ips[1])
	b2, _ := strconv.Atoi(ips[2])
	b3, _ := strconv.Atoi(ips[3])
	n += uint32(b0) << 24
	n += uint32(b1) << 16
	n += uint32(b2) << 8
	n += uint32(b3)
	return n
}

//GetOutboundIP 获得本机内网IP，这个函数只能在启动的时候掉，运行中掉有风险
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func GetOutboundIPInt() uint64 {
	ipint := strings.ReplaceAll(GetOutboundIP(), ".", "")
	v, _ := strconv.Atoi(ipint)
	return uint64(v)
}

//GetExternalIP 获取公网IP
func GetExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return strings.TrimSpace(string(content))
}

//GetLocalIP 获得内网IP
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}

		}
	}
	return ""
}

var privateIPBlocks []*net.IPNet

// IsPrivateIP 是否是局域网IP
func IsPrivateIP(ipStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, errors.New("Invalid IP")
	}

	if privateIPBlocks == nil {
		for _, cidr := range []string{
			"127.0.0.0/8",    // IPv4 loopback
			"10.0.0.0/8",     // RFC1918
			"172.16.0.0/12",  // RFC1918
			"192.168.0.0/16", // RFC1918
			"169.254.0.0/16", // RFC3927 link-local
			"::1/128",        // IPv6 loopback
			"fe80::/10",      // IPv6 link-local
			"fc00::/7",       // IPv6 unique local addr
		} {
			_, block, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			privateIPBlocks = append(privateIPBlocks, block)
		}
	}

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true, nil
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}
