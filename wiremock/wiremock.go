package wiremock

import (
	"os/exec"
)

func runWiremockIfNotStarted() error {
	cmd := exec.Command("docker run -itd -v $PWD/mappings:/home/wiremock/mappings -p 8080:8080 --name wiremock wiremock")
	return cmd.Run()
}

func stopWiremock() error {
	cmd := exec.Command("docker stop wiremock")
	return cmd.Run()
}
