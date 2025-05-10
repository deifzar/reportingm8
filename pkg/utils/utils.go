package utils

import (
	"errors"
	"maps"
	"net"
)

func GetSecurityPosture(stats map[string]int) (string, error) {
	for level, nvul := range stats {
		if level == "critical" && nvul > 0 {
			return "extreme", nil
		}
		if level == "high" && nvul > 0 {
			return "high", nil
		}
		if level == "medium" && nvul > 0 {
			return "moderate", nil
		}
		if level == "low" && nvul > 0 {
			return "low", nil
		}
		if level == "info" && nvul > 0 {
			return "minimal", nil
		}
	}
	var err = errors.New("cannot calculate security posture")
	return "", err
}

func MergeMaps(map1, map2 map[string]map[string]int) map[string]map[string]int {
	UniqueMap := make(map[string]map[string]int)

	// for loop for the first map
	for key, mapaux := range map1 {
		maps.Copy(mapaux, map2[key])
		UniqueMap[key] = mapaux
	}
	// return merged result
	return UniqueMap
}

func IsValidIPAddress(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress != nil
}
