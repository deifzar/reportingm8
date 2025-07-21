package utils

import (
	"maps"
	"net"
)

func GetSecurityPosture(stats map[string]int) string {
	for level, nvul := range stats {
		if level == "critical" && nvul > 0 {
			return "extreme"
		}
		if level == "high" && nvul > 0 {
			return "high"
		}
		if level == "medium" && nvul > 0 {
			return "moderate"
		}
		if level == "low" && nvul > 0 {
			return "low"
		}
		if level == "info" && nvul > 0 {
			return "minimal"
		}
	}
	return "unknown"
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
