package scripts

import "embed"

//go:embed *
var Scripts embed.FS

type COMMAND string

const (
	Fd COMMAND = "fd"
	Fg COMMAND = "fg"
	Vf COMMAND = "vf"
	Fs COMMAND = "fs"
)
