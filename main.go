package main

import (
	kubectlspillsecrets "github.com/tanakapayam/kubectl-spill-secrets/lib"
)

func main() {
	kubectlspillsecrets.ParseArgs()
	kubectlspillsecrets.SpillSecrets()
}
