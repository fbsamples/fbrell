package app

import (
	"flag"
)

var (
	ID        uint64
	Secret    string
	Namespace string
)

func init() {
	flag.Uint64Var(
		&ID, "rell.client.id", 184484190795, "Default Client ID.")
	flag.StringVar(
		&Secret, "rell.client.secret", "", "Default Client Secret.")
	flag.StringVar(
		&Namespace, "rell.client.namespace", "", "Default Client Namespace.")
}
