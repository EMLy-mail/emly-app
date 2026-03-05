//go:build !debug

package utils

const isDebugBuild = false

func IsRunningInDebugMode() bool {
	return isDebugBuild
}
