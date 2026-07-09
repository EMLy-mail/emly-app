// Package ipcpb holds the generated protobuf types for the EMLyUpdater ⇄
// EMLy named-pipe protocol. This must stay in sync with the copy in
// emly-updater/internal/ipc/ipcpb — see ../../../../proto/updateripc.proto
// for the manual-sync note. Regenerate after editing the proto:
//
//	protoc --go_out=. --go_opt=paths=source_relative -I ../../../../proto ../../../../proto/updateripc.proto
//
// protoc and protoc-gen-go are not required by CI or `go build`; the
// generated file is committed (and thus effectively vendored already).
package ipcpb

//go:generate protoc --go_out=. --go_opt=paths=source_relative -I ../../../../proto ../../../../proto/updateripc.proto
