//go:build debug

package utils

const isDebugBuild = true

func IsRunningInDebugMode() bool {
	return isDebugBuild
}
