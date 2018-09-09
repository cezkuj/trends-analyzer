package wiremock

import(
"os/exec"

func runWiremockIfNotStarted() error{
   return exec.Run("docker run -itd -v $PWD/mappings:/home/wiremock/mappings -p 8080:8080 --name wiremock wiremock")
}

func stopWiremock(){
   return exec.Run("docker stop wiremock")

}
